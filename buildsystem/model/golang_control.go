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

	gps, err := gc.buildPackages()
	if err != nil {
		return nil, errors.Wrap(err, "building package definitions")
	}
	_ = g.MergePackages(gps...)

	vds, err := gc.buildVariantDistros()
	if err != nil {
		return nil, errors.Wrap(err, "building variant-distro mappings")
	}
	_ = g.MergeVariantDistros(vds...)

	ngvps, err := gc.buildVariantParameters()
	if err != nil {
		return nil, errors.Wrap(err, "building variant parameters")
	}
	_ = g.MergeVariantParameters(ngvps...)

	envs, err := gc.buildEnvironments()
	if err != nil {
		return nil, errors.Wrap(err, "building environment variables")
	}
	_ = g.MergeEnvironments(envs...)

	tags, err := gc.buildDefaultTags()
	if err != nil {
		return nil, errors.Wrap(err, "building default tags")
	}
	_ = g.MergeDefaultTags(tags...)

	if err := g.DiscoverPackages(); err != nil {
		return nil, errors.Wrap(err, "automatically discovering test packages")
	}

	g.ApplyDefaultTags()

	if err := g.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid build configuration")
	}

	return &g, nil
}

func (gc *GolangControl) buildPackages() ([]GolangPackage, error) {
	var all []GolangPackage
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

		all = append(all, gps...)

		return nil
	}); err != nil {
		return nil, errors.WithStack(err)
	}

	return all, nil
}

func (gc *GolangControl) buildVariantDistros() ([]VariantDistro, error) {
	var all []VariantDistro
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

		all = append(all, vds...)

		return nil
	}); err != nil {
		return nil, errors.WithStack(err)
	}

	return all, nil
}

func (gc *GolangControl) buildVariantParameters() ([]NamedGolangVariantParameters, error) {
	var all []NamedGolangVariantParameters

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

		all = append(all, ngvps...)

		return nil
	}); err != nil {
		return nil, errors.WithStack(err)
	}

	return all, nil
}

func (gc *GolangControl) buildEnvironments() ([]map[string]string, error) {
	var all []map[string]string

	if err := withMatchingFiles(gc.WorkingDirectory, gc.EnvironmentFiles, func(file string) error {
		env := map[string]string{}
		if err := utility.ReadYAMLFileStrict(file, &env); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		all = append(all, env)

		return nil
	}); err != nil {
		return nil, errors.WithStack(err)
	}

	return all, nil
}

func (gc *GolangControl) buildDefaultTags() ([]string, error) {
	var all []string
	if err := withMatchingFiles(gc.WorkingDirectory, gc.DefaultTagFiles, func(file string) error {
		tags := []string{}
		if err := utility.ReadYAMLFileStrict(file, &tags); err != nil {
			return errors.Wrap(err, "unmarshalling from YAML file")
		}

		all = append(all, tags...)

		return nil
	}); err != nil {
		return nil, errors.WithStack(err)
	}

	return all, nil
}
