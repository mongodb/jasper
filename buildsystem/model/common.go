package model

import "github.com/mongodb/grip"

// VariantDistro represents a mapping between a variant name and the distros
// that it runs on.
type VariantDistro struct {
	Name    string   `yaml:"name"`
	Distros []string `yaml:"distros"`
}

func (vd *VariantDistro) Validate() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(vd.Name == "", "missing variant name")
	catcher.NewWhen(len(vd.Distros) == 0, "need to specify at least one distro")
	return catcher.Resolve()
}

// MergeEnvironments returns the merged environment variable mappings where the
// given input environments are ordered in increasing priority. If there are
// duplicate environment variables, in the environments, the variable definition
// of the higher priority environment takes precedence.
func MergeEnvironments(envsByPriority ...map[string]string) map[string]string {
	merged := map[string]string{}
	for _, env := range envsByPriority {
		for k, v := range env {
			merged[k] = v
		}
	}
	return merged
}
