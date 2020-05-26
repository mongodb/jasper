package generator

import (
	"github.com/evergreen-ci/shrub"
	"github.com/evergreen-ci/utility"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

// Make represents an evergreen config generator for Makefile-based projects.
type Make struct {
	Tasks    map[string]MakeTask    `yaml:"tasks,omitempty"`
	Variants map[string]MakeVariant `yaml:"variants"`
	// Environment defines global environment variables. Definitions can be
	// overridden at the task or variant level.
	Environment      map[string]string `yaml:"environment,omitempty"`
	WorkingDirectory string            `yaml:"-"`
}

// NewMake creates a new evergreen config generator for Make from a single file
// that contains all the necessary generation information.
func NewMake(file string, workingDir string) (*Make, error) {
	m := Make{WorkingDirectory: workingDir}
	if err := utility.ReadYAMLFileStrict(file, &m); err != nil {
		return nil, errors.Wrap(err, "unmarshalling from YAML file")
	}

	if err := m.Validate(); err != nil {
		return nil, errors.Wrap(err, "make generator configuration")
	}

	return &m, nil
}

// MakeBuilder is used to build a Make generator from its component parts.
type MakeBuilder struct {
	Tasks       map[string]MakeTask
	Variants    map[string]MakeVariant
	Environment map[string]string
}

// MergeTasks merges task definitions.
func (m *Make) MergeTasks(mts map[string]MakeTask) *Make {
	for newName, newMT := range mts {
		m.Tasks[newName] = newMT
	}
	return m
}

// MergeVariantDistros merges variant-distro mappings.
func (m *Make) MergeVariantDistros(vds map[string]VariantDistro) *Make {
	for name, newVD := range vds {
		mv := m.Variants[name]
		mv.VariantDistro = newVD
		m.Variants[name] = mv
	}
	return m
}

// MergeVariants merges variant parameters.
func (m *Make) MergeVariantParameters(mvps map[string]MakeVariantParameters) *Make {
	for name, newMVP := range mvps {
		mv := m.Variants[name]
		mv.MakeVariantParameters = newMVP
		m.Variants[name] = mv
	}
	return m
}

// MergeEnvironments merges the given environments with the existing environment
// variables. Duplicates are overwritten in the order in which environments are
// passed into the function.
func (m *Make) MergeEnvironments(envs ...map[string]string) *Make {
	m.Environment = MergeEnvironments(append([]map[string]string{m.Environment}, envs...)...)
	return m
}

func (m *Make) Validate() error {
	catcher := grip.NewBasicCatcher()
	catcher.Wrap(m.validateTasks(), "invalid task definitions")
	catcher.Wrap(m.validateVariants(), "invalid variant definitions")
	return catcher.Resolve()
}

func (m *Make) validateTasks() error {
	catcher := grip.NewBasicCatcher()
	for name, mt := range m.Tasks {
		catcher.Wrapf(mt.Validate(), "invalid task '%s'", name)
	}
	return catcher.Resolve()
}

func (m *Make) validateVariants() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(len(m.Variants) == 0, "must have at least one variant")
	for variantName, mv := range m.Variants {
		catcher.Wrapf(mv.Validate(), "invalid definitions for variant '%s'", variantName)

		taskNames := map[string]struct{}{}
		for _, mvt := range mv.Tasks {
			res, err := m.getTasksAndRef(mvt)
			if err != nil {
				catcher.Wrapf(err, "invalid task reference in variant '%s'", variantName)
				continue
			}
			for taskName := range res.tasks {
				if _, ok := taskNames[taskName]; ok {
					catcher.Errorf("duplicate reference to task name '%s' in variant '%s'", taskName, variantName)
				}
				taskNames[taskName] = struct{}{}
			}
		}
	}
	return catcher.Resolve()
}

type taskReferenceResult struct {
	tasks map[string]MakeTask
	ref   string
}

func (m *Make) getTasksAndRef(mvt MakeVariantTask) (*taskReferenceResult, error) {
	if mvt.Tag != "" {
		tasks := m.GetTasksByTag(mvt.Tag)
		if len(tasks) == 0 {
			return nil, errors.Errorf("no tasks matched tag '%s'", mvt.Tag)
		}
		return &taskReferenceResult{tasks: tasks, ref: mvt.Tag}, nil
	}

	if mvt.Name != "" {
		task, err := m.GetTaskByName(mvt.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "finding definition for package named '%s'", mvt.Name)
		}
		tasks := map[string]MakeTask{mvt.Name: *task}
		return &taskReferenceResult{tasks: tasks, ref: mvt.Name}, nil
	}

	return nil, errors.New("empty package reference")
}

func (m *Make) Generate() (*shrub.Configuration, error) {
	conf, err := shrub.BuildConfiguration(func(c *shrub.Configuration) {
		for vName, mv := range m.Variants {
			variant := c.Variant(vName)
			variant.DistroRunOn = mv.Distros

			var tasksForVariant []*shrub.Task
			for _, mvt := range mv.Tasks {
				res, err := m.getTasksAndRef(mvt)
				if err != nil {
					panic(err)
				}
				tasksForVariant = append(tasksForVariant, m.generateVariantTasksForRef(c, vName, mv, res.tasks)...)
			}

			getProjectCmd := shrub.CmdGetProject{
				Directory: m.WorkingDirectory,
			}

			if len(variant.TaskSpecs) >= minTasksForTaskGroup {
				tg := c.TaskGroup(vName + "_group").SetMaxHosts(len(variant.TaskSpecs) / 2)
				tg.SetupTask = shrub.CommandSequence{getProjectCmd.Resolve()}

				for _, task := range variant.TaskSpecs {
					_ = tg.Task(task.Name)
				}
				_ = variant.AddTasks(tg.GroupName)
			} else {
				for _, task := range tasksForVariant {
					task.Commands = append([]*shrub.CommandDefinition{getProjectCmd.Resolve()}, task.Commands...)
					_ = variant.AddTasks(task.Name)
				}
			}
		}
	})

	if err != nil {
		return nil, errors.Wrap(err, "generating evergreen configuration")
	}
	return conf, nil
}

func (m *Make) generateVariantTasksForRef(c *shrub.Configuration, vName string, mv MakeVariant, mts map[string]MakeTask) []*shrub.Task {
	var tasks []*shrub.Task
	for name, mt := range mts {
		cmds := m.subprocessExecCmds(mv, mt)
		tasks = append(tasks, c.Task(getTaskName(vName, name)).Command(cmds...))
	}
	return tasks
}

func (m *Make) subprocessExecCmds(mv MakeVariant, mt MakeTask) []shrub.Command {
	env := MergeEnvironments(m.Environment, mv.Environment, mt.Environment)
	var cmds []shrub.Command
	for _, target := range mt.Targets {
		cmds = append(cmds, &shrub.CmdExec{
			Binary: "make",
			// kim: TODO: people should maybe be able to specify additional
			// make args for tasks.
			Args: []string{target},
			Env:  env,
		})
	}
	return cmds
}

func (m *Make) GetTaskByName(name string) (*MakeTask, error) {
	if mt, ok := m.Tasks[name]; ok {
		return &mt, nil
	}
	return nil, errors.Errorf("task with name '%s' not found", name)
}

func (m *Make) GetTasksByTag(tag string) map[string]MakeTask {
	tasks := map[string]MakeTask{}
	for name, mt := range m.Tasks {
		if utility.StringSliceContains(mt.Tags, tag) {
			tasks[name] = mt
		}
	}
	return tasks
}

// MakeTask represents a task that runs Make targets.
type MakeTask struct {
	Targets []string `yaml:"targets"`
	Tags    []string `yaml:"tags,omitempty"`
	// Environment defines task-specific environment variables. This has higher
	// precedence than global environment variables but lower precedence than
	// variant-specific environment variables.
	Environment map[string]string `yaml:"environment,omitempty"`
}

func (mt *MakeTask) Validate() error {
	catcher := grip.NewBasicCatcher()
	// catcher.NewWhen(mt.Name == "", "missing task name")
	catcher.NewWhen(len(mt.Targets) == 0, "need to specify at least one target")
	tags := map[string]struct{}{}
	for _, tag := range mt.Tags {
		if _, ok := tags[tag]; ok {
			catcher.Errorf("duplicate tag '%s'", tag)
		}
		tags[tag] = struct{}{}
	}
	return catcher.Resolve()
}

type MakeVariant struct {
	VariantDistro         `yaml:",inline"`
	MakeVariantParameters `yaml:",inline"`
}

type MakeVariantParameters struct {
	// Environment defines variant-specific environment variables. This has
	// higher precedence than global or task-specific environment variables.
	Environment map[string]string `yaml:"environment,omitempty"`
	Tasks       []MakeVariantTask `yaml:"tasks"`
}

func (mvp *MakeVariantParameters) Validate() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(len(mvp.Tasks) == 0, "need to specify at least one task")
	for _, mvt := range mvp.Tasks {
		catcher.Wrap(mvt.Validate(), "invalid task reference")
	}
	return catcher.Resolve()
}

func (mv *MakeVariant) Validate() error {
	catcher := grip.NewBasicCatcher()
	catcher.Add(mv.VariantDistro.Validate())
	catcher.Add(mv.MakeVariantParameters.Validate())
	return catcher.Resolve()
}

// MakeVariantTask represents a reference to a task or group of tasks by either
// task name or tasks containing the tag.
type MakeVariantTask struct {
	Name string `yaml:"name,omitempty"`
	Tag  string `yaml:"tag,omitempty"`
}

func (mvt *MakeVariantTask) Validate() error {
	if mvt.Name == "" && mvt.Tag == "" {
		return errors.New("must specify either a task name or tag")
	}
	if mvt.Name != "" && mvt.Tag != "" {
		return errors.New("cannot specify both task name and tag")
	}
	return nil
}
