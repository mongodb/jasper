package generator

import (
	"io/ioutil"

	"github.com/evergreen-ci/shrub"
	"github.com/evergreen-ci/utility"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Make struct {
	Tasks    []MakeTask    `yaml:"tasks,omitempty"`
	Variants []MakeVariant `yaml:"variants"`
	// Environment defines global environment variables. Definitions can be
	// overridden at the task or variant level.
	Environment map[string]string `yaml:"environment,omitempty"`
}

// NewMake creates a new evergreen config generator for Make from a single file
// that contains all the necessary generation information.
func NewMake(file string) (*Make, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "reading configuration file")
	}

	var m Make
	if err := yaml.UnmarshalStrict(b, &m); err != nil {
		return nil, errors.Wrap(err, "unmarshalling configuration file from YAML")
	}

	if err := m.Validate(); err != nil {
		return nil, errors.Wrap(err, "make generator configuration")
	}

	return &m, nil
}

type MakeBuilder struct {
	Tasks       []MakeTask
	Variants    []MakeVariant
	Environment map[string]string
}

func (mb *MakeBuilder) MergeTasks(mts ...MakeTask) error {
	for _, newMT := range mts {
		for _, mt := range mb.Tasks {
			if mt.Name == newMT.Name {
				return errors.Errorf("duplicate definition of task '%s'", mt.Name)
			}
		}
	}
	mb.Tasks = append(mb.Tasks, mts...)
	return nil
}

// kim: TODO: merge duplicates by variant name?
// kim: TODO: merge into Variants instead of splitting into separate
// VariantDistros? We would have to ensure that MergeVariantDistros is
// called _before_ MergeVariants is called to ensure the variant-distro
// mappings exist before the tasks are assigned.
// kim: NOTE: this is special case that allows merging of duplicate
// names just because the variant distro mapping can be in a separate
// location from the variants. We might just combine the parameters to this
// with MergeVariants to ensure that they get merged at the same time and in
// the right order (VariantDistros first, followed by Variants for
// task references).

// kim: TODO: merge with above functionality to avoid this duplication handling
// business.
func (mb *MakeBuilder) MergeVariants(vds []VariantDistro, mvs []MakeVariant) error {
	for _, newVD := range vds {
		var found bool
		for i, mv := range mb.Variants {
			if mv.Name == newVD.Name {
				return errors.Errorf("duplicate variant distro mapping for variant '%s'", mv.Name)
			}
		}
		mb.Variants = append(mb.Variants, MakeVariant{VariantDistro: newVD})
	}

	for _, newMV := range mvs {
		var found bool
		for i, mv := range mb.Variants {
			if mv.Name == newMV.Name {
				mb.Variants[i] = mv
				found = true
				break
			}
		}
		if !found {
			mb.Variants = append(mb.Variants, newMV)
		}
	}
	return mb
}

// NOTE: this is the only case in which the result is to overwrite rather than
// error on duplicates.
func (mb *MakeBuilder) MergeEnvironments(envs ...map[string]string) *MakeBuilder {
	for _, env := range envs {
		mb.Environment = MergeEnvironments(mb.Environment, env)
	}
	return mb
}

func (mb *MakeBuilder) Build() (*Make, error) {
	return nil, errors.New("TODO: implement")
}

func (m *Make) Validate() error {
	catcher := grip.NewBasicCatcher()
	catcher.Wrap(m.validateTasks(), "invalid task definitions")
	catcher.Wrap(m.validateVariants(), "invalid variant definitions")
	return catcher.Resolve()
}

func (m *Make) validateTasks() error {
	catcher := grip.NewBasicCatcher()
	for _, mt := range m.Tasks {
		catcher.Wrapf(mt.Validate(), "invalid task '%s'", mt.Name)
	}
	return catcher.Resolve()
}

func (m *Make) validateVariants() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(len(m.Variants) == 0, "must have at least one variant")
	for _, mv := range m.Variants {
		catcher.Wrapf(mv.Validate(), "invalid definitions for variant '%s'", mv.Name)

		taskNames := map[string]struct{}{}
		for _, mvt := range mv.Tasks {
			tasks, _, err := m.getTasksAndRef(mvt)
			if err != nil {
				catcher.Wrapf(err, "invalid task reference in variant '%s'", mv.Name)
				continue
			}
			for _, mt := range tasks {
				if _, ok := taskNames[mt.Name]; ok {
					catcher.Errorf("duplicate reference to task name '%s' in variant '%s'", mt.Name, mv.Name)
				}
				taskNames[mt.Name] = struct{}{}
			}
		}
	}
	return catcher.Resolve()
}

func (m *Make) getTasksAndRef(mvt MakeVariantTask) ([]MakeTask, string, error) {
	if mvt.Tag != "" {
		tasks := m.GetTasksByTag(mvt.Tag)
		if len(tasks) == 0 {
			return nil, "", errors.Errorf("no tasks matched tag '%s'", mvt.Tag)
		}
		return tasks, mvt.Tag, nil
	}

	if mvt.Name != "" {
		task, err := m.GetTaskByName(mvt.Name)
		if err != nil {
			return nil, "", errors.Wrapf(err, "finding definition for package named '%s'", mvt.Name)
		}
		return []MakeTask{*task}, mvt.Name, nil
	}

	return nil, "", errors.New("empty package reference")
}

func (m *Make) Generate() (*shrub.Configuration, error) {
	conf, err := shrub.BuildConfiguration(func(c *shrub.Configuration) {
		// kim: TODO: make subprocess.exec commands from variant references ->
		// resolve task references -> make subprocess.exec commands with proper
		// environment
		// for _, mt := range m.Tasks {
		//     m.subprocessExecCmd(mt)
		//     _ = c.Task(mt.Name).Command()
		// }
	})

	if err != nil {
		return nil, errors.Wrap(err, "generating evergreen configuration")
	}
	return conf, nil
}

func (m *Make) subprocessExecCmd(mvt MakeVariantTask, mt MakeTask) (*shrub.CmdExec, error) {
	// kim: TODO: merge with priority: variant > task > global
	return nil, errors.New("TODO: implement")
}

func (m *Make) GetTaskByName(name string) (*MakeTask, error) {
	for _, mt := range m.Tasks {
		if mt.Name == name {
			return &mt, nil
		}
	}
	return nil, errors.Errorf("task with name '%s' not found", name)
}

func (m *Make) GetTasksByTag(tag string) []MakeTask {
	var tasks []MakeTask
	for _, mvt := range m.Tasks {
		if utility.StringSliceContains(mvt.Tags, tag) {
			tasks = append(tasks, mvt)
		}
	}
	return tasks
}

type MakeTask struct {
	Name    string   `yaml:"name"`
	Targets []string `yaml:"targets"`
	Tags    []string `yaml:"tags,omitempty"`
	// Environment defines task-specific environment variables. This has higher
	// precedence than global environment variables but lower precedence than
	// variant-specific environment variables.
	Environment map[string]string `yaml:"environment,omitempty"`
}

func (mt *MakeTask) Validate() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(mt.Name == "", "missing task name")
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

// kim: NOTE: if we actually want to split the variant-distro mappings from the
// variant-target definitions, we will have to add a name field and figure out
// how to consolidate them into the single MakeVariant struct.
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
