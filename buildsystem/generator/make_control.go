package generator

import (
	"path/filepath"

	"github.com/evergreen-ci/utility"
	"github.com/pkg/errors"
)

// MakeControl represents a Make generator constructed in parts from various
// files.
type MakeControl struct {
	VariantDistroFiles    []string `yaml:"variant_distro_files"`
	VariantParameterFiles []string `yaml:"variant_parameter_files"`
	TaskFiles             []string `yaml:"task_files"`
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

func (mc *MakeControl) Build() (*Make, error) {
	m := Make{}
	if err := withMatchingFiles(mc.WorkDir, mc.VariantDistroFiles, func(file string) error {
		vds := map[string]VariantDistro{}
		if err := utility.ReadYAMLFileStrict(file, &vds); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		_ = m.MergeVariantDistros(vds)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building variant-distro mappings")
	}

	if err := withMatchingFiles(mc.WorkDir, mc.VariantParameterFiles, func(file string) error {
		mvps := map[string]MakeVariantParameters{}
		if err := utility.ReadYAMLFileStrict(file, &mvps); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		_ = m.MergeVariantParameters(mvps)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building variant parameters")
	}

	if err := withMatchingFiles(mc.WorkDir, mc.TaskFiles, func(file string) error {
		mts := map[string]MakeTask{}
		if err := utility.ReadYAMLFileStrict(file, &mts); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		_ = m.MergeTasks(mts)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building task definitions")
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

	return &m, nil
}
