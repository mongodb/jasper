package generator

import (
	"io/ioutil"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// VariantDistro represents a mapping between a variant name and the distros
// that it runs on.
type VariantDistro struct {
	// kim: TODO: maybe remove Name since we use maps now for naming
	Name    string   `yaml:"-"`
	Distros []string `yaml:"distros"`
}

// kim: TODO: probably remove since we don't want to do duplicate
// reconciliation.
// func (vd *VariantDistro) Merge(add VariantDistro) (*VariantDistro, error) {
//     if vd.Name != add.Name {
//         return nil, errors.Errorf("cannot merge variant-distro mapping for variants named '%s' and '%s' because they have different names", vd.Name, add.Name)
//     }
//     merged := *vd
//     for _, newDistro := range add.Distros {
//         if utility.StringSliceContains(merged.Distros, newDistro) {
//             continue
//         }
//         merged.Distros = append(merged.Distros, newDistro)
//     }
//     return &merged, nil
// }

func (vd *VariantDistro) Validate() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(vd.Name == "", "missing variant name")
	catcher.NewWhen(len(vd.Distros) == 0, "need to specify at least one distro")
	return catcher.Resolve()
}

func NewEnvironment(file string) (map[string]string, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "reading environment configuration")
	}

	env := map[string]string{}
	if err := yaml.UnmarshalStrict(b, &env); err != nil {
		return nil, errors.Wrap(err, "unmarshalling environment variables from YAML")
	}

	return env, nil
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
