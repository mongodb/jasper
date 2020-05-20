package generator

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// MakeControl represents a Make generator constructed in parts from various
// files.
type MakeControl struct {
	VariantDistroFiles    []string `yaml:"variant_distro_files"`
	VariantParameterFiles []string `yaml:"variant_parameter_files"`
	TaskFiles             []string `yaml:"task_files"`
	EnvironmentFiles      []string `yaml:"environment_files"`
}

// NewMakeControl creates a new representation of a Make control file from the
// given file.
func NewMakeControl(file string) (*MakeControl, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "reading control file")
	}
	mc := MakeControl{}
	if err := yaml.UnmarshalStrict(content, &mc); err != nil {
		return nil, errors.Wrap(err, "unmarshalling control file from YAML")
	}
	return &mc, nil
}

func (mc *MakeControl) Build() (*Make, error) {
	m := Make{}
	if err := withGlobMatches(mc.VariantDistroFiles, func(file string) error {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return errors.Wrap(err, "reading file")
		}

		vds := map[string]VariantDistro{}
		if err := yaml.UnmarshalStrict(content, &vds); err != nil {
			return errors.Wrap(err, "unmarshalling file from YAML")
		}

		_ = m.MergeVariantDistros(vds)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building variant-distro mappings")
	}

	if err := withGlobMatches(mc.VariantParameterFiles, func(file string) error {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return errors.Wrap(err, "reading file")
		}

		mvps := map[string]MakeVariantParameters{}
		if err := yaml.UnmarshalStrict(content, &mvps); err != nil {
			return errors.Wrap(err, "unmarshalling file from YAML")
		}

		_ = m.MergeVariantParameters(mvps)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building variant parameters")
	}

	if err := withGlobMatches(mc.TaskFiles, func(file string) error {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return errors.Wrap(err, "reading file")
		}

		mts := map[string]MakeTask{}
		if err := yaml.UnmarshalStrict(content, &mts); err != nil {
			return errors.Wrap(err, "unmarshalling file from YAML")
		}

		_ = m.MergeTasks(mts)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building task definitions")
	}

	if err := withGlobMatches(mc.EnvironmentFiles, func(file string) error {
		content, err := ioutil.ReadFile(file)

		if err != nil {
			return errors.Wrap(err, "reading file")
		}

		env := map[string]string{}
		if err := yaml.UnmarshalStrict(content, &env); err != nil {
			return errors.Wrap(err, "unmarshalling file from YAML")
		}

		_ = m.MergeEnvironments(env)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building environment variables")
	}

	return &m, nil
}
