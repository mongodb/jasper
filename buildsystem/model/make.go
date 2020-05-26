package model

import (
	"github.com/evergreen-ci/utility"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

// Make represents the configuration for Make-based projects.
type Make struct {
	// kim: TODO: add definitions for specific targets (e.g. tags)
	Tasks    []MakeTask    `yaml:"tasks,omitempty"`
	Variants []MakeVariant `yaml:"variants"`
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
// - Each task name is unique.
// - Each task is valid.
func (m *Make) validateTasks() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(len(m.Tasks) == 0, "must have at least one task")
	taskNames := map[string]struct{}{}
	for _, mt := range m.Tasks {
		if _, ok := taskNames[mt.Name]; ok {
			catcher.Errorf("duplicate task name '%s'", mt.Name)
		}
		catcher.Wrapf(mt.Validate(), "invalid task '%s'", mt.Name)
	}
	return catcher.Resolve()
}

// validateVariants checks that:
// - Variants are defined.
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
			for _, mt := range tasks {
				if _, ok := taskNames[mt.Name]; ok {
					catcher.Errorf("duplicate reference to task '%s' in variant '%s'", mt.Name, variantName)
				}
				taskNames[mt.Name] = struct{}{}
			}
		}
	}
	return catcher.Resolve()
}

// GetTasksAndRef returns the tasks that match the reference specified in the
// given MakeVariantTask.
func (m *Make) GetTasksAndRef(mvt MakeVariantTask) ([]MakeTask, error) {
	if mvt.Tag != "" {
		tasks := m.GetTasksByTag(mvt.Tag)
		if len(tasks) == 0 {
			return nil, errors.Errorf("no tasks matched tag '%s'", mvt.Tag)
		}
		return tasks, nil
	}

	if mvt.Name != "" {
		task, _, err := m.GetTaskIndexByName(mvt.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "finding definition for task '%s'", mvt.Name)
		}
		return []MakeTask{*task}, nil
	}

	return nil, errors.New("empty package reference")
}

// GetTaskIndexByName finds the definition of the task and its index by the task
// name.
func (m *Make) GetTaskIndexByName(name string) (mt *MakeTask, index int, err error) {
	for i, t := range m.Tasks {
		if t.Name == name {
			return &t, i, nil
		}
	}
	return nil, -1, errors.Errorf("task with name '%s' not found", name)
}

// GetTasksByTag finds the definition of tasks matching the given tag.
func (m *Make) GetTasksByTag(tag string) []MakeTask {
	var tasks []MakeTask
	for _, mt := range m.Tasks {
		if utility.StringSliceContains(mt.Tags, tag) {
			tasks = append(tasks, mt)
		}
	}
	return tasks
}

func (m *Make) GetVariantIndexByName(name string) (mv *MakeVariant, index int, err error) {
	for i, v := range m.Variants {
		if v.Name == name {
			return &v, i, nil
		}
	}
	return nil, -1, errors.Errorf("variant with name '%s' not found", name)
}

// MakeTask represents a task that runs a group of Make targets.
type MakeTask struct {
	Name    string   `yaml:"name"`
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
	catcher.NewWhen(mt.Name == "", "missing task name")
	catcher.NewWhen(len(mt.Targets) == 0, "must specify at least one target")
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

// Validate checks that the variant-distro mapping and the Make-specific
// parameters are valid.
func (mv *MakeVariant) Validate() error {
	catcher := grip.NewBasicCatcher()
	catcher.Add(mv.VariantDistro.Validate())
	catcher.Add(mv.MakeVariantParameters.Validate())
	return catcher.Resolve()
}

// MakeVariantParameters describe Make-specific variant configuration.
type MakeVariantParameters struct {
	// Environment defines variant-specific environment variables. This has
	// higher precedence than global or task-specific environment variables.
	Environment map[string]string `yaml:"environment,omitempty"`
	Tasks       []MakeVariantTask `yaml:"tasks"`
}

// NamedMakeVariantParameters describes Make-specific variant configuration
// associated with a particular variant name.
type NamedMakeVariantParameters struct {
	Name                  string `yaml:"name"`
	MakeVariantParameters `yaml:",inline"`
}

// Validate checks that tasks are defined and each task reference is valid.
func (mvp *MakeVariantParameters) Validate() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(len(mvp.Tasks) == 0, "must specify at least one task")
	for _, mvt := range mvp.Tasks {
		catcher.Wrap(mvt.Validate(), "invalid task reference")
	}
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
func (m *Make) MergeTasks(mts []MakeTask) *Make {
	for _, mt := range mts {
		if _, i, err := m.GetTaskIndexByName(mt.Name); err == nil {
			m.Tasks[i] = mt
		} else {
			m.Tasks = append(m.Tasks, mt)
		}
	}
	return m
}

// MergeVariantDistros merges variant-distro mappings with the existing ones by
// variant name. For a given variant name, existing variant-distro mappings are
// overwritten if they are already defined.
func (m *Make) MergeVariantDistros(vds []VariantDistro) *Make {
	for _, vd := range vds {
		if mv, i, err := m.GetVariantIndexByName(vd.Name); err == nil {
			mv.VariantDistro = vd
			m.Variants[i] = *mv
		} else {
			m.Variants = append(m.Variants, MakeVariant{
				VariantDistro: vd,
			})
		}
	}
	return m
}

// MergeVariantParameters merges variant parameters with the existing ones by
// name. For a given variant name, existing variant options are overwritten if
// they are already defined.
func (m *Make) MergeVariantParameters(mvps []NamedMakeVariantParameters) *Make {
	for _, mvp := range mvps {
		if mv, i, err := m.GetVariantIndexByName(mvp.Name); err == nil {
			mv.MakeVariantParameters = mvp.MakeVariantParameters
			m.Variants[i] = *mv
		} else {
			m.Variants = append(m.Variants, MakeVariant{
				VariantDistro:         VariantDistro{Name: mvp.Name},
				MakeVariantParameters: mvp.MakeVariantParameters,
			})
		}
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