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
	Name    string   `yaml:"name"`
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

// MergeEnvironments returns the merged environment variable mappings. If there
// are duplicate mappings in env and toMerge, the variable definition in toMerge
// takes precedence over the one in env.
func MergeEnvironments(env map[string]string, toMerge map[string]string) map[string]string {
	merged := map[string]string{}
	for k, v := range env {
		merged[k] = v
	}
	for k, v := range toMerge {
		merged[k] = v
	}
	return merged
}
