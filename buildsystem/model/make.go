package model

import (
	"github.com/evergreen-ci/utility"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

// Make represents the configuration for Make-based projects.
type Make struct {
	// kim: TODO: add definitions for specific targets (e.g. tags)
	Tasks    map[string]MakeTask    `yaml:"tasks,omitempty"`
	Variants map[string]MakeVariant `yaml:"variants"`
	// Environment defines global environment variables. Definitions can be
	// overridden at the task or variant level.
	Environment map[string]string `yaml:"environment,omitempty"`
}

// NewMake creates a new evergreen config generator for Make from a single file
// that contains all the necessary generation information.
func NewMake(file string) (*Make, error) {
	m := Make{}
	if err := utility.ReadYAMLFileStrict(file, &m); err != nil {
		return nil, errors.Wrap(err, "unmarshalling from YAML file")
	}

	if err := m.Validate(); err != nil {
		return nil, errors.Wrap(err, "make generator configuration")
	}

	return &m, nil
}

// Validate checks that the task and variant definitions are valid.
func (m *Make) Validate() error {
	catcher := grip.NewBasicCatcher()
	catcher.Wrap(m.validateTasks(), "invalid task definitions")
	catcher.Wrap(m.validateVariants(), "invalid variant definitions")
	return catcher.Resolve()
}

// validateTasks checks that:
// - Tasks are defined.
// - Each task is valid.
func (m *Make) validateTasks() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(len(m.Tasks) == 0, "must have at least one task")
	for name, mt := range m.Tasks {
		catcher.Wrapf(mt.Validate(), "invalid task '%s'", name)
	}
	return catcher.Resolve()
}

// validateVariants checks that:
// - Variants are defined
// - Each task referenced in a variant references a defined task.
// - Each variant does not specify a duplicate task.
func (m *Make) validateVariants() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(len(m.Variants) == 0, "must have at least one variant")
	for variantName, mv := range m.Variants {
		catcher.Wrapf(mv.Validate(), "invalid definitions for variant '%s'", variantName)

		taskNames := map[string]struct{}{}
		for _, mvt := range mv.Tasks {
			tasks, err := m.GetTasksAndRef(mvt)
			if err != nil {
				catcher.Wrapf(err, "invalid task reference in variant '%s'", variantName)
				continue
			}
			for taskName := range tasks {
				if _, ok := taskNames[taskName]; ok {
					catcher.Errorf("duplicate reference to task name '%s' in variant '%s'", taskName, variantName)
				}
				taskNames[taskName] = struct{}{}
			}
		}
	}
	return catcher.Resolve()
}

// GetTasksAndRef returns the tasks that match the reference specified in the
// given MakeVariantTask.
func (m *Make) GetTasksAndRef(mvt MakeVariantTask) (map[string]MakeTask, error) {
	if mvt.Tag != "" {
		tasks := m.GetTasksByTag(mvt.Tag)
		if len(tasks) == 0 {
			return nil, errors.Errorf("no tasks matched tag '%s'", mvt.Tag)
		}
		return tasks, nil
	}

	if mvt.Name != "" {
		task, err := m.GetTaskByName(mvt.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "finding definition for package named '%s'", mvt.Name)
		}
		tasks := map[string]MakeTask{mvt.Name: *task}
		return tasks, nil
	}

	return nil, errors.New("empty package reference")
}

// GetTaskByName finds the definition of the task by the task name.
func (m *Make) GetTaskByName(name string) (*MakeTask, error) {
	if mt, ok := m.Tasks[name]; ok {
		return &mt, nil
	}
	return nil, errors.Errorf("task with name '%s' not found", name)
}

// GetTasksByTag finds the definition of tasks matching the given tag.
func (m *Make) GetTasksByTag(tag string) map[string]MakeTask {
	tasks := map[string]MakeTask{}
	for name, mt := range m.Tasks {
		if utility.StringSliceContains(mt.Tags, tag) {
			tasks[name] = mt
		}
	}
	return tasks
}

// MakeTask represents a task that runs a group of Make targets.
type MakeTask struct {
	Targets []string `yaml:"targets"`
	Tags    []string `yaml:"tags,omitempty"`
	// Environment defines task-specific environment variables. This has higher
	// precedence than global environment variables but lower precedence than
	// variant-specific environment variables.
	Environment map[string]string `yaml:"environment,omitempty"`
}

// Validate checks that targets are defined and all tags are unique.
func (mt *MakeTask) Validate() error {
	catcher := grip.NewBasicCatcher()
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

// MakeVariant defines a variant that runs Make tasks.
type MakeVariant struct {
	VariantDistro         `yaml:",inline"`
	MakeVariantParameters `yaml:",inline"`
}

// MakeVariantParameters describe Make-specific variant configuration.
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

// MakeVariantTask represents either a reference to a task by task name, or a
// group of tasks containing a particular tag.
type MakeVariantTask struct {
	Name string `yaml:"name,omitempty"`
	Tag  string `yaml:"tag,omitempty"`
}

// Validate checks that exactly one kind of reference is specified in a task
// reference for a variant.
func (mvt *MakeVariantTask) Validate() error {
	if mvt.Name == "" && mvt.Tag == "" {
		return errors.New("must specify either a task name or tag")
	}
	if mvt.Name != "" && mvt.Tag != "" {
		return errors.New("cannot specify both task name and tag")
	}
	return nil
}

// MergeTasks merges task definitions with the existing ones by task name. For a
// given task name, existing tasks are overwritten if they are already defined.
func (m *Make) MergeTasks(mts map[string]MakeTask) *Make {
	for newName, newMT := range mts {
		m.Tasks[newName] = newMT
	}
	return m
}

// MergeVariantDistros merges variant-distro mappings with the existing ones by
// variant name. For a given variant name, existing variant-distro mappings are
// overwritten if they are already defined.
func (m *Make) MergeVariantDistros(vds map[string]VariantDistro) *Make {
	for name, newVD := range vds {
		mv := m.Variants[name]
		mv.VariantDistro = newVD
		m.Variants[name] = mv
	}
	return m
}

// MergeVariants merges variant parameters with the existing ones by name. For a
// given variant name, existing variant options are overwritten if they are
// already defined.
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
