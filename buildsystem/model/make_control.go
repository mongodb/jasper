package model

import (
	"path/filepath"

	"github.com/evergreen-ci/utility"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

// MakeControl represents a Make generator constructed in parts from various
// files.
type MakeControl struct {
	TargetSequenceFiles   []string `yaml:"target_sequence_files"`
	TaskFiles             []string `yaml:"task_files"`
	VariantDistroFiles    []string `yaml:"variant_distro_files"`
	VariantParameterFiles []string `yaml:"variant_parameter_files"`
	EnvironmentFiles      []string `yaml:"environment_files"`

	WorkDir string `yaml:"-"`
}

// NewMakeControl creates a new representation of a Make control file from the
// given file.
func NewMakeControl(file string) (*MakeControl, error) {
	mc := MakeControl{
		WorkDir: filepath.Dir(file),
	}
	if err := utility.ReadYAMLFileStrict(file, &mc); err != nil {
		return nil, errors.Wrap(err, "unmarshalling from YAML file")
	}
	return &mc, nil
}

// Build creates a Make model from the files referenced in the MakeControl.
func (mc *MakeControl) Build() (*Make, error) {
	var m Make

	if err := withMatchingFiles(mc.WorkDir, mc.TargetSequenceFiles, func(file string) error {
		mtss := []MakeTargetSequence{}
		if err := utility.ReadYAMLFileStrict(file, &mtss); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		_ = m.MergeTargetSequences(mtss)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building target sequence definitions")
	}

	if err := withMatchingFiles(mc.WorkDir, mc.TaskFiles, func(file string) error {
		mts := []MakeTask{}
		if err := utility.ReadYAMLFileStrict(file, &mts); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		catcher := grip.NewBasicCatcher()
		for _, mt := range mts {
			catcher.Wrapf(mt.Validate(), "task '%s'", mt.Name)
		}
		if catcher.HasErrors() {
			return errors.Wrap(catcher.Resolve(), "invalid task definitions")
		}

		_ = m.MergeTasks(mts)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building task definitions")
	}

	if err := withMatchingFiles(mc.WorkDir, mc.VariantDistroFiles, func(file string) error {
		vds := []VariantDistro{}
		if err := utility.ReadYAMLFileStrict(file, &vds); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		catcher := grip.NewBasicCatcher()
		for _, vd := range vds {
			catcher.Wrapf(vd.Validate(), "variant '%s'", vd.Name)
		}
		if catcher.HasErrors() {
			return errors.Wrap(catcher.Resolve(), "invalid variant-distro mappings")
		}

		_ = m.MergeVariantDistros(vds)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building variant-distro mappings")
	}

	if err := withMatchingFiles(mc.WorkDir, mc.VariantParameterFiles, func(file string) error {
		nmvps := []NamedMakeVariantParameters{}
		if err := utility.ReadYAMLFileStrict(file, &nmvps); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		catcher := grip.NewBasicCatcher()
		for _, nmvp := range nmvps {
			catcher.Wrapf(nmvp.Validate(), "variant '%s'", nmvp.Name)
		}
		if catcher.HasErrors() {
			return errors.Wrap(catcher.Resolve(), "invalid variant parameters")
		}

		_ = m.MergeVariantParameters(nmvps)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building variant parameters")
	}

	if err := withMatchingFiles(mc.WorkDir, mc.EnvironmentFiles, func(file string) error {
		env := map[string]string{}
		if err := utility.ReadYAMLFileStrict(file, &env); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		_ = m.MergeEnvironments(env)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building environment variables")
	}

	if err := m.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid configuration")
	}

	return &m, nil
}
