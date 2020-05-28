package model

import (
	"github.com/evergreen-ci/utility"
	"github.com/mongodb/grip"
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
	DefaultTagFiles       []string `yaml:"default_tag_files"`

	WorkingDirectory string `yaml:"-"`
}

func NewGolangControl(file, workingDir string) (*GolangControl, error) {
	gc := GolangControl{
		WorkingDirectory: workingDir,
	}
	if err := utility.ReadYAMLFileStrict(file, &gc); err != nil {
		return nil, errors.Wrap(err, "unmarshalling from YAML file")
	}
	return &gc, nil
}

func (gc *GolangControl) Build() (*Golang, error) {
	g := Golang{RootPackage: gc.RootPackage}

	if err := withMatchingFiles(gc.WorkingDirectory, gc.PackageFiles, func(file string) error {
		gps := []GolangPackage{}
		if err := utility.ReadYAMLFileStrict(file, &gps); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		catcher := grip.NewBasicCatcher()
		for _, gp := range gps {
			catcher.Wrapf(gp.Validate(), "package '%s'", gp.Name)
		}
		if catcher.HasErrors() {
			return errors.Wrap(catcher.Resolve(), "invalid package definitions")
		}

		_ = g.MergePackages(gps...)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building package definitions")
	}

	if err := withMatchingFiles(gc.WorkingDirectory, gc.VariantDistroFiles, func(file string) error {
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

		_ = g.MergeVariantDistros(vds...)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building variant-distro mappings")
	}

	if err := withMatchingFiles(gc.WorkingDirectory, gc.VariantParameterFiles, func(file string) error {
		ngvps := []NamedGolangVariantParameters{}
		if err := utility.ReadYAMLFileStrict(file, &ngvps); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		catcher := grip.NewBasicCatcher()
		for _, ngvp := range ngvps {
			catcher.Wrapf(ngvp.Validate(), "variant '%s'", ngvp.Name)
		}
		if catcher.HasErrors() {
			return errors.Wrap(catcher.Resolve(), "invalid variant parameters")
		}

		_ = g.MergeVariantParameters(ngvps...)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building variant parameters")
	}

	if err := withMatchingFiles(gc.WorkingDirectory, gc.EnvironmentFiles, func(file string) error {
		env := map[string]string{}
		if err := utility.ReadYAMLFileStrict(file, &env); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		_ = g.MergeEnvironments(env)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building environment variables")
	}

	if err := withMatchingFiles(gc.WorkingDirectory, gc.DefaultTagFiles, func(file string) error {
		tags := []string{}
		if err := utility.ReadYAMLFileStrict(file, &tags); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		_ = g.MergeDefaultTags(tags...)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "building default tags")
	}

	if err := g.DiscoverPackages(); err != nil {
		return nil, errors.Wrap(err, "automatically discovering test packages")
	}

	g.ApplyDefaultTags()

	if err := g.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid configuration")
	}

	return &g, nil
}
