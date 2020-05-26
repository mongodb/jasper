package model

import (
	"github.com/evergreen-ci/utility"
	"github.com/pkg/errors"
)

type GolangControl struct {
	// TODO: this configuration is required to know the full go package name but
	// doesn't have a designated config file where it can be placed.
	RootPackage string `yaml:"root_package"`

	VariantDistroFiles    []string `yaml:"variant_distro_files"`
	VariantParameterFiles []string `yaml:"variant_parameter_files"`
	PackageFiles          []string `yaml:"package_files"`
	EnvironmentFiles      []string `yaml:"environment_files"`

	WorkDir string `yaml:"-"`
}

func NewGolangControl(file, workingDir string) (*GolangControl, error) {
	gc := GolangControl{
		WorkDir: workingDir,
	}
	if err := utility.ReadYAMLFileStrict(file, &gc); err != nil {
		return nil, errors.Wrap(err, "unmarshalling from YAML file")
	}
	return &gc, nil
}

func (gc *GolangControl) Build() (*Golang, error) {
	g := Golang{RootPackage: gc.RootPackage}

	if err := withMatchingFiles(gc.WorkDir, gc.PackageFiles, func(file string) error {
		gps := []GolangPackage{}
		if err := utility.ReadYAMLFileStrict(file, &gps); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		_ = g.MergePackages(gps)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building package definitions")
	}

	if err := withMatchingFiles(gc.WorkDir, gc.VariantDistroFiles, func(file string) error {
		vds := []VariantDistro{}
		if err := utility.ReadYAMLFileStrict(file, &vds); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		_ = g.MergeVariantDistros(vds)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building variant-distro mappings")
	}

	if err := withMatchingFiles(gc.WorkDir, gc.VariantParameterFiles, func(file string) error {
		gvps := []NamedGolangVariantParameters{}
		if err := utility.ReadYAMLFileStrict(file, &gvps); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		_ = g.MergeVariantParameters(gvps)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building variant parameters")
	}

	if err := withMatchingFiles(gc.WorkDir, gc.EnvironmentFiles, func(file string) error {
		env := map[string]string{}
		if err := utility.ReadYAMLFileStrict(file, &env); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		_ = g.MergeEnvironments(env)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building environment variables")
	}

	if err := g.DiscoverPackages(); err != nil {
		return nil, errors.Wrap(err, "automatically discovering test packages")
	}

	if err := g.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid configuration")
	}

	return &g, nil
}
