package model

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/evergreen-ci/utility"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

// Golang represents a configuration for generating an evergreen configuration
// from a project that uses golang.
type Golang struct {
	// Environment defines any environment variables. GOPATH and GOROOT are
	// required. If the working directory is specified, GOPATH must be specified
	// as a subdirectory of the working directory.
	// kim: TODO: make it optionally a separate file
	Environment map[string]string `yaml:"environment"`
	// RootPackage is the name of the root package for the project (e.g.
	// github.com/mongodb/jasper).
	RootPackage string `yaml:"root_package"`
	// Packages explicitly sets options for packages that should be tested.
	Packages []GolangPackage `yaml:"packages,omitempty"`
	// Variants describe the mapping between packages and distros to run them
	// on.
	Variants []GolangVariant `yaml:"variants"`

	// WorkingDirectory is the absolute path to the base directory where the
	// GOPATH directory is located.
	WorkingDirectory string `yaml:"-"`
}

// NewGolang returns a model of a Golang build configuration from a single file
// and working directory where the GOPATH directory is located.
func NewGolang(file, workingDir string) (*Golang, error) {
	g := Golang{
		WorkingDirectory: workingDir,
	}
	if err := utility.ReadYAMLFileStrict(file, &g); err != nil {
		return nil, errors.Wrap(err, "unmarshalling from YAML file")
	}

	if err := g.DiscoverPackages(); err != nil {
		return nil, errors.Wrap(err, "automatically discovering test packages")
	}

	if err := g.Validate(); err != nil {
		return nil, errors.Wrap(err, "golang generator configuration")
	}

	return &g, nil
}

func (g *Golang) Validate() error {
	catcher := grip.NewBasicCatcher()

	catcher.NewWhen(g.RootPackage == "", "must specify the import path of the root package of the project")
	catcher.Wrap(g.validatePackages(), "invalid package definitions")
	catcher.Wrap(g.validateEnvVars(), "invalid environment variables")
	catcher.Wrap(g.validateVariants(), "invalid variant definitions")

	return catcher.Resolve()
}

func (g *Golang) validatePackages() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(len(g.Packages) == 0, "must have at least one package to test")
	pkgNames := map[string]struct{}{}
	unnamedPkgPaths := map[string]struct{}{}
	for _, pkg := range g.Packages {
		catcher.Wrapf(pkg.Validate(), "invalid package definition '%s'", pkg.Path)

		if pkg.Name == "" {
			if _, ok := unnamedPkgPaths[pkg.Path]; ok {
				catcher.Errorf("cannot have duplicate unnamed package definitions for path '%s'", pkg.Path)
			}
			unnamedPkgPaths[pkg.Path] = struct{}{}
			continue
		}
		if _, ok := pkgNames[pkg.Name]; ok {
			catcher.Errorf("cannot have duplicate package named '%s'", pkg.Name)
		}
		pkgNames[pkg.Name] = struct{}{}
	}
	return catcher.Resolve()
}

func (g *Golang) validateEnvVars() error {
	catcher := grip.NewBasicCatcher()
	for _, name := range []string{"GOPATH", "GOROOT"} {
		if val, ok := g.Environment[name]; ok && val != "" {
			g.Environment[name] = filepath.ToSlash(val)
			continue
		}
		if val := os.Getenv(name); val != "" {
			g.Environment[name] = filepath.ToSlash(val)
			continue
		}
		catcher.Errorf("environment variable '%s' must be explicitly defined or already present in the environment", name)
	}
	if catcher.HasErrors() {
		return catcher.Resolve()
	}

	// According to the semantics of the generator's GOPATH, it must be relative
	// to the working directory (if specified).
	relGopath, err := g.RelGopath()
	if err != nil {
		catcher.Wrap(err, "converting GOPATH to relative path")
	} else {
		g.Environment["GOPATH"] = relGopath
	}

	return catcher.Resolve()
}

// RelGopath returns the GOPATH in the environment relative to the working
// directory (if it is defined).
func (g *Golang) RelGopath() (string, error) {
	gopath := filepath.ToSlash(g.Environment["GOPATH"])
	workingDir := filepath.ToSlash(g.WorkingDirectory)
	if workingDir != "" && strings.HasPrefix(gopath, workingDir) {
		return filepath.Rel(workingDir, gopath)
	}
	if filepath.IsAbs(gopath) {
		return "", errors.New("GOPATH is absolute path, but needs to be relative path")
	}
	return gopath, nil
}

// AbsGopath converts the relative GOPATH in the environment into an absolute
// path based on the working directory.
func (g *Golang) AbsGopath() (string, error) {
	gopath := filepath.ToSlash(g.Environment["GOPATH"])
	workingDir := filepath.ToSlash(g.WorkingDirectory)
	if workingDir != "" && !strings.HasPrefix(gopath, workingDir) {
		return filepath.Join(workingDir, gopath), nil
	}
	if !filepath.IsAbs(gopath) {
		return "", errors.New("GOPATH is relative path, but needs to be absolute path")
	}
	return gopath, nil
}

func (g *Golang) validateVariants() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(len(g.Variants) == 0, "must specify at least one variant")
	varNames := map[string]struct{}{}
	for _, gv := range g.Variants {
		catcher.Wrapf(gv.Validate(), "invalid definition for variant '%s'", gv.Name)

		if _, ok := varNames[gv.Name]; ok {
			catcher.Errorf("cannot have duplicate variant name '%s'", gv.Name)
		}
		varNames[gv.Name] = struct{}{}

		pkgNames := map[string]struct{}{}
		pkgPaths := map[string]struct{}{}
		for _, gvp := range gv.Packages {
			pkgs, _, err := g.GetPackagesAndRef(gvp)
			if err != nil {
				catcher.Wrapf(err, "invalid package reference in variant '%s'", gv.Name)
				continue
			}
			for _, pkg := range pkgs {
				if pkg.Name != "" {
					if _, ok := pkgNames[pkg.Name]; ok {
						catcher.Errorf("duplicate reference to package name '%s' in variant '%s'", pkg.Name, gv.Name)
					}
					pkgNames[pkg.Name] = struct{}{}
				} else if pkg.Path != "" {
					if _, ok := pkgPaths[pkg.Path]; ok {
						catcher.Errorf("duplicate reference to package path '%s' in variant '%s'", pkg.Path, gv.Name)
					}
					pkgPaths[pkg.Path] = struct{}{}
				}
			}
		}
	}
	return catcher.Resolve()
}

// golangTestFileSuffix is the suffix indicating that a golang file is meant to
// be run as a test.
const (
	golangTestFileSuffix = "_test.go"
	golangVendorDir      = "vendor"
)

// DiscoverPackages discovers directories containing tests in the local file
// system and adds them if they are not already defined.
func (g *Golang) DiscoverPackages() error {
	if err := g.validateEnvVars(); err != nil {
		return errors.Wrap(err, "invalid environment variables")
	}

	projectPath, err := g.AbsProjectPath()
	if err != nil {
		return errors.Wrap(err, "getting project path as an absolute path")
	}

	if err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		fileName := filepath.Base(info.Name())
		if fileName == golangVendorDir {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if !strings.Contains(fileName, golangTestFileSuffix) {
			return nil
		}
		dir := filepath.Dir(path)
		dir, err = filepath.Rel(projectPath, dir)
		if err != nil {
			return errors.Wrapf(err, "making package path '%s' relative to root package", path)
		}
		// If package has already been added, skip adding it.
		if _, err = g.GetPackageByPath(dir); err == nil {
			return nil
		}

		pkg := GolangPackage{
			Path: dir,
		}

		g.Packages = append(g.Packages, pkg)

		return nil
	}); err != nil {
		return errors.Wrapf(err, "walking the file system tree starting from path '%s'", projectPath)
	}

	return nil
}

// GetPackagesAndRef returns all packages that match the reference specified in
// the given GolangVariantPackage and the reference string from the
// GolangVariantPackage.
func (g *Golang) GetPackagesAndRef(gvp GolangVariantPackage) ([]GolangPackage, string, error) {
	if gvp.Tag != "" {
		pkgs := g.GetPackagesByTag(gvp.Tag)
		if len(pkgs) == 0 {
			return nil, "", errors.Errorf("no packages matched tag '%s'", gvp.Tag)
		}
		return pkgs, gvp.Tag, nil
	}

	if gvp.Name != "" {
		pkg, err := g.GetPackageByName(gvp.Name)
		if err != nil {
			return nil, "", errors.Wrapf(err, "finding definition for package named '%s'", gvp.Name)
		}
		return []GolangPackage{*pkg}, gvp.Name, nil
	} else if gvp.Path != "" {
		pkg, err := g.GetPackageByPath(gvp.Path)
		if err != nil {
			return nil, "", errors.Wrapf(err, "finding definition for package path '%s'", gvp.Path)
		}
		return []GolangPackage{*pkg}, gvp.Path, nil
	}

	return nil, "", errors.New("empty package reference")
}

// RelProjectPath returns the path to the project relative to the working
// directory.
func (g *Golang) RelProjectPath() (string, error) {
	gopath, err := g.RelGopath()
	if err != nil {
		return "", errors.Wrap(err, "getting GOPATH as a relative path")
	}
	return filepath.Join(gopath, "src", g.RootPackage), nil
}

// AbsProjectPath returns the absolute path to the project.
func (g *Golang) AbsProjectPath() (string, error) {
	gopath, err := g.AbsGopath()
	if err != nil {
		return "", errors.Wrap(err, "getting GOPATH as an absolute path")
	}
	return filepath.Join(gopath, "src", g.RootPackage), nil
}

// GetPackageByName returns the package matching the name.
func (g *Golang) GetPackageByName(name string) (*GolangPackage, error) {
	for _, pkg := range g.Packages {
		if pkg.Name == name {
			return &pkg, nil
		}
	}
	return nil, errors.Errorf("package with name '%s' not found", name)
}

// GetPackageByPath returns the package matching the path.
func (g *Golang) GetPackageByPath(path string) (*GolangPackage, error) {
	for _, pkg := range g.Packages {
		if pkg.Path == path {
			return &pkg, nil
		}
	}
	return nil, errors.Errorf("package with path '%s' not found", path)
}

// GetPackagesByTag returns the packages that match the tag.
func (g *Golang) GetPackagesByTag(tag string) []GolangPackage {
	var pkgs []GolangPackage
	for _, pkg := range g.Packages {
		if utility.StringSliceContains(pkg.Tags, tag) {
			pkgs = append(pkgs, pkg)
		}
	}
	return pkgs
}

type GolangPackage struct {
	// Name is an alias for the package.
	Name string `yaml:"name,omitempty"`
	// Path is the path of the package relative to the root package. For
	// example, "." refers to the root package while "util" refers to a
	// subpackage called "util" within the root package.
	Path string `yaml:"path"`
	// Tags are labels that allow you to logically group related packages.
	Tags    []string             `yaml:"tags,omitempty"`
	Options GolangRuntimeOptions `yaml:"options,omitempty"`
}

func (gp *GolangPackage) Validate() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(gp.Path == "", "need to specify package path")
	tags := map[string]struct{}{}
	for _, tag := range gp.Tags {
		if _, ok := tags[tag]; ok {
			catcher.Errorf("duplicate tag '%s'", tag)
		}
		tags[tag] = struct{}{}
	}
	catcher.Wrap(gp.Options.Validate(), "invalid golang options")
	return catcher.Resolve()
}

// GolangVariant defines a mapping between distros that run packages and the
// golang packages to run.
type GolangVariant struct {
	VariantDistro `yaml:",inline"`
	// Packages lists a package name, path or or tagged group of packages
	// relative to the root package.
	Packages []GolangVariantPackage `yaml:"packages"`
	// Options are variant-specific options that modify test execution.
	// Explicitly setting these values will override options specified under the
	// definitions of packages.
	Options *GolangRuntimeOptions `yaml:"options,omitempty"`
}

func (gv *GolangVariant) Validate() error {
	catcher := grip.NewBasicCatcher()
	catcher.Add(gv.VariantDistro.Validate())
	catcher.NewWhen(len(gv.Packages) == 0, "need to specify at least one package")
	pkgPaths := map[string]struct{}{}
	pkgNames := map[string]struct{}{}
	pkgTags := map[string]struct{}{}
	for _, pkg := range gv.Packages {
		catcher.Wrap(pkg.Validate(), "invalid package reference")
		if path := pkg.Path; path != "" {
			if _, ok := pkgPaths[path]; ok {
				catcher.Errorf("duplicate reference to package path '%s'", path)
			}
			pkgPaths[path] = struct{}{}
		}
		if name := pkg.Name; name != "" {
			if _, ok := pkgNames[name]; ok {
				catcher.Errorf("duplicate reference to package name '%s'", name)
			}
			pkgNames[name] = struct{}{}
		}
		if tag := pkg.Tag; tag != "" {
			if _, ok := pkgTags[tag]; ok {
				catcher.Errorf("duplicate reference to package tag '%s'", tag)
			}
			pkgTags[tag] = struct{}{}

		}
	}
	if gv.Options != nil {
		catcher.Wrap(gv.Options.Validate(), "invalid runtime options")
	}
	return catcher.Resolve()
}

// GolangVariantPackage is a specifier that references a golang package.
type GolangVariantPackage struct {
	Name string `yaml:"name,omitempty"`
	Path string `yaml:"path,omitempty"`
	Tag  string `yaml:"tag,omitempty"`
}

// Validate ensures that exactly one kind of reference is specified in the
// variant package reference.
func (gvp *GolangVariantPackage) Validate() error {
	var numRefs int
	for _, ref := range []string{gvp.Name, gvp.Path, gvp.Tag} {
		if ref != "" {
			numRefs++
		}
	}
	if numRefs != 1 {
		return errors.New("must specify exactly one of the following: name, path, or tag")
	}
	return nil
}

// GolangRuntimeOptions specify additional options to modify behavior of runtime
// execution.
type GolangRuntimeOptions []string

// Validate ensures that options to the scripting environment (i.e. the go
// binary) do not contain duplicate flags.
func (gro GolangRuntimeOptions) Validate() error {
	seen := map[string]struct{}{}
	catcher := grip.NewBasicCatcher()
	for _, flag := range gro {
		flag = cleanupFlag(flag)
		// Don't allow the verbose flag because the scripting harness sets
		// verbose.
		if flagIsVerbose(flag) {
			catcher.New("verbose flag is already specified")
		}
		if _, ok := seen[flag]; ok {
			catcher.Errorf("duplicate flag '%s'", flag)
			continue
		}
		seen[flag] = struct{}{}
	}
	return catcher.Resolve()
}

// flagIsVerbose returns whether or not the flag is the golang flag for verbose
// testing.
func flagIsVerbose(flag string) bool {
	flag = cleanupFlag(flag)
	return flag == "v"
}

// golangTestPrefix is the optional prefix that each golang test flag can have.
// Flags with this prefix have identical meaning to their non-prefixed flag.
// (e.g. "test.v" and "v" are identical).
const golangTestPrefix = "test."

// cleanupFlag cleans up the golang-style flag and returns just the name of the
// flag. Golang flags have the form -<flag_name>[=value].
func cleanupFlag(flag string) string {
	flag = strings.TrimPrefix(flag, "-")
	flag = strings.TrimPrefix(flag, golangTestPrefix)

	// We only care about the flag name, not its set value.
	flagAndValue := strings.Split(flag, "=")
	flag = flagAndValue[0]

	return flag
}

// Merge returns the merged set of GolangRuntimeOptions. Unique flags between
// the two flag sets are added together. Duplicate flags are handled by
// overwriting conflicting flags with overwrite's flags.
func (gro GolangRuntimeOptions) Merge(overwrite GolangRuntimeOptions) GolangRuntimeOptions {
	merged := gro
	for _, flag := range overwrite {
		if i := merged.flagIndex(flag); i != -1 {
			merged = append(merged[:i], merged[i+1:]...)
			merged = append(merged, flag)
		} else {
			merged = append(merged, flag)
		}
	}

	return merged
}

// flagIndex returns the index where the flag is set if it is present. If it is
// absent, this returns -1.
func (gro GolangRuntimeOptions) flagIndex(flag string) int {
	flag = cleanupFlag(flag)
	for i, f := range gro {
		f = cleanupFlag(f)
		if f == flag {
			return i
		}
	}
	return -1
}
