package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/evergreen-ci/shrub"
	"github.com/evergreen-ci/utility"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Golang represents a configuration for generating an evergreen configuration
// from a project that uses golang.
type Golang struct {
	// Environment defines any environment variables. GOPATH and GOROOT are
	// required.
	Environment map[string]string `yaml:"environment,omitempty"`
	// RootPackage is the name of the root package for the project (e.g.
	// github.com/mongodb/jasper).
	RootPackage string `yaml:"root_package"`
	// Packages explicitly sets options for packages that should be tested.
	Packages []GolangPackage `yaml:"packages"`
	Variants []Variant       `yaml:"variants"`
}

func NewGolang(file string) (Generator, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "could not read configuration file")
	}

	g := Golang{}
	if err := yaml.UnmarshalStrict(b, &g); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal configuration file from YAML")
	}

	if err := g.DiscoverPackages(); err != nil {
		return nil, errors.Wrap(err, "error while automatically discovering test packages")
	}

	if err := g.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid golang generator configuration")
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
	// pkgPaths := map[string]struct{}{}
	for _, pkg := range g.Packages {
		catcher.Wrapf(pkg.Validate(), "invalid package definition '%s'", pkg.Path)

		if pkg.Name == "" {
			continue
		}
		if _, ok := pkgNames[pkg.Name]; ok {
			catcher.Errorf("cannot have duplicate package named '%s'", pkg.Name)
		}
		pkgNames[pkg.Name] = struct{}{}
		// kim: TODO: remove since we've separated names/paths in variant
		// definitions
		// if dupPkg, ok := pkgPaths[pkg.Name]; ok {
		//     catcher.Errorf("cannot have package named '%s' because it is ambiguous for package with path '%s'", pkg.Name, dupPkg.Path)
		// }
		// pkgPath[pkg.Path] = struct{}{}
	}
	return catcher.Resolve()
}

func (g *Golang) validateEnvVars() error {
	catcher := grip.NewBasicCatcher()
	for _, name := range []string{"GOPATH", "GOROOT"} {
		if _, ok := g.Environment[name]; ok {
			continue
		}
		if val := os.Getenv(name); val != "" {
			g.Environment[name] = val
			continue
		}
		catcher.Errorf("environment variable '%s' must be explicitly defined or already present in the environment", name)
	}
	return catcher.Resolve()
}

func (g *Golang) validateVariants() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(len(g.Variants) == 0, "must specify at least one variant")
	varNames := map[string]struct{}{}
	for _, v := range g.Variants {
		catcher.Wrapf(v.Validate(), "invalid definition for variant '%s'", v.Name)

		if _, ok := varNames[v.Name]; ok {
			catcher.Errorf("cannot have duplicate variant name '%s'", v.Name)
		}
		varNames[v.Name] = struct{}{}
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
	gopath, ok := g.Environment["GOPATH"]
	if !ok {
		return errors.New("cannot discover packages if GOPATH is not defined")
	}
	rootPackagePath := filepath.Join(gopath, "src", g.RootPackage)

	if err := filepath.Walk(rootPackagePath, func(path string, info os.FileInfo, err error) error {
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
		// If package has already been added, skip adding it.
		dir := filepath.Dir(path)
		if _, err := g.GetPackageByPath(dir); err == nil {
			return nil
		}

		pkg := GolangPackage{
			Path: dir,
		}

		g.Packages = append(g.Packages, pkg)

		return nil
	}); err != nil {
		return errors.Wrapf(err, "could not walk the file system tree starting from path '%s'", rootPackagePath)
	}

	return nil
}

const (
	// minTasksForTaskGroup is the minimum number of tasks that have to be in a
	// task group in order for it to be worthwhile to create a task group. Since
	// max hosts must be at least and we don't want to use single-host task
	// groups, we must have at least four tasks in the group to make a task
	// group.
	minTasksForTaskGroup = 4
)

// Generate creates the evergreen configuration from the given golang build
// configuration.
func (g *Golang) Generate(output io.Writer) error {
	conf, err := shrub.BuildConfiguration(func(c *shrub.Configuration) {
		gopath := "go"
		projectDir := filepath.Join(gopath, "src", g.RootPackage)

		for _, v := range g.Variants {

			variant := c.Variant(v.Name)
			variant.DistroRunOn = v.Distros

			var tasksForVariant []*shrub.Task
			// Make one task per package in this variant. We cannot make one
			// task per package, because we have to account for variant-level
			// options possibly overriding package-level options, which requires
			// making separate tasks with different commands.
			for _, pkgRef := range v.Packages {
				taskName := fmt.Sprintf("%s_%s", v.Name, pkgRef)

				var pkgs []GolangPackage
				if pkgRef.Tag != "" {
					// We have to expand tagged packages now because shrub doesn't
					// support tags.
					pkgs := g.GetPackagesByTag(pkgRef.Tag)
					if len(pkgs) == 0 {
						panic(errors.Errorf("no packages matched tag '%s' for variant '%s'", pkgRef.Tag, v.Name))
					}
				} else {
					var pkg *GolangPackage
					var err error
					if pkgRef.Name != "" {
						pkg, err = g.GetPackageByName(pkgRef.Name)
						if err != nil {
							panic(err)
						}
					} else if pkgRef.Path != "" {
						pkg, err = g.GetPackageByPath(pkgRef.Path)
						if err != nil {
							panic(err)
						}
					} else {
						panic(errors.Errorf("empty package reference in variant '%s'", v.Name))
					}

					pkgs = []GolangPackage{*pkg}
				}

				for _, pkg := range pkgs {
					testOpts := pkg.Options
					if v.Options != nil {
						testOpts = v.Options.Merge(testOpts)
					}

					// kim: TODO: for each package within a variant, check for an
					// existing package name/path.
					// kim: TODO: implement once subprocess.scripting supports
					// (scripting.Harness).Test()
					// kim: TODO: handle precedence of variant overriding package
					// options when specifying options.
					scriptCmd := shrub.CmdSubprocessScripting{
						HarnessPath: gopath,
						Env:         g.Environment,
						// TestDir:     filepath.Join(gopath, "src", g.RootPackage, pkg.Path),
						// TestOptions: testOpts,
					}

					t := c.Task(taskName).Command(scriptCmd)
					tasksForVariant = append(tasksForVariant, t)
				}
			}

			// Don't bother using making a task group for this variant if it
			// doesn't meet the threshold number of tasks
			// If this variant runs the threshold number of packages, add
			if len(tasksForVariant) >= minTasksForTaskGroup {
				tg := c.TaskGroup(v.Name + "_group").SetMaxHosts(len(tasksForVariant) / 2)
				getProjectCmd := shrub.CmdGetProject{
					Directory: projectDir,
				}
				tg.SetupTask = shrub.CommandSequence{getProjectCmd.Resolve()}

				for _, task := range tasksForVariant {
					tg.Task(task.Name)
				}
				variant = variant.AddTasks(tg.GroupName)
			} else {
				for _, task := range tasksForVariant {
					variant = variant.AddTasks(task.Name)
				}
			}
		}
	})

	if err != nil {
		return errors.Wrap(err, "failed to build configuration")
	}

	if err := writeConfig(conf, output); err != nil {
		return errors.Wrap(err, "failed to write generated configuration")
	}

	return nil
}

// writeConfig writes the evergreen configuration to the given output as JSON.
func writeConfig(conf *shrub.Configuration, output io.Writer) error {
	data, err := json.Marshal(conf)
	if err != nil {
		return errors.Wrap(err, "could not marshal configuration as JSON")
	}

	if _, err = io.Copy(output, bytes.NewBuffer(data)); err != nil {
		return errors.Wrapf(err, "could not write configuration")
	}

	return nil
}

func (g *Golang) GetPackageByName(name string) (*GolangPackage, error) {
	for _, pkg := range g.Packages {
		if pkg.Name == name {
			return &pkg, nil
		}
	}
	return nil, errors.Errorf("package with name '%s' not found", name)
}

func (g *Golang) GetPackageByPath(path string) (*GolangPackage, error) {
	for _, pkg := range g.Packages {
		if pkg.Path == path {
			return &pkg, nil
		}
	}
	return nil, errors.Errorf("package with path '%s' not found", path)
}

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

func (g *GolangPackage) Validate() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(g.Path == "", "need to specify package path")
	tags := map[string]struct{}{}
	for _, tag := range g.Tags {
		if _, ok := tags[tag]; ok {
			catcher.Errorf("duplicate tag '%s'", tag)
		}
		tags[tag] = struct{}{}
	}
	return catcher.Resolve()
}

// Variant defines a mapping between distros that run packages and the packages
// to run.
type Variant struct {
	Name    string   `yaml:"name"`
	Distros []string `yaml:"distros"`
	// Packages lists the names or paths of the package relative to the root
	// package, which can be one of the following:
	// - A name of a package explicitly listed in the packages section.
	// - A path of a package explicitly listed in the packages section.
	// - A path of a package that matches a path to a directory containing tests
	//   relative to the root package.
	// - A package tag, beginning with a period (e.g. ".tag").
	Packages []VariantPackage `yaml:"packages"`
	// Options are variant-specific options that modify test execution.
	// Explicitly setting these values will override options specified under the
	// definitions of packages.
	Options *GolangRuntimeOptions `yaml:"options,omitempty"`
}

type VariantPackage struct {
	Name string `yaml:"name,omitempty"`
	Path string `yaml:"path,omitempty"`
	Tag  string `yaml:"tag,omitempty"`
}

func (vp *VariantPackage) Validate() error {
	var numRefs int
	for _, ref := range []string{vp.Name, vp.Path, vp.Tag} {
		if ref != "" {
			numRefs++
		}
	}
	if numRefs != 1 {
		return errors.New("must specify exactly one of the following: name, path, or tag")
	}
	return nil
}

func (v *Variant) Validate() error {
	catcher := grip.NewBasicCatcher()
	catcher.NewWhen(v.Name == "", "missing variant name")
	catcher.NewWhen(len(v.Distros) == 0, "need to specify at least one distro")
	catcher.NewWhen(len(v.Packages) == 0, "need to specify at least one package")
	for _, pkg := range v.Packages {
		catcher.Wrap(pkg.Validate(), "invalid package reference")
	}
	if v.Options != nil {
		catcher.Wrap(v.Options.Validate(), "invalid runtime options")
	}
	return catcher.Resolve()
}

// kim: TODO: handle variant/task precedence, maybe make more tasks based on
// variant overriding settings to handle variant-specific settings.
type GolangRuntimeOptions struct {
	// TODO: consider removing
	// Pattern      *string `yaml:"pattern,omitempty"`
	// TimeoutSecs  *int    `yaml:"timeout_secs,omitempty"`
	// RunCount     *int    `yaml:"run_count,omitempty"`
	// RaceDetector *bool   `yaml:"race_detector,omitempty"`
	// Short        *bool   `yaml:"short,omitempty"`
	// Coverage     *bool   `yaml:"coverage,omitempty"`
	// CoverageMode *string `yaml:"coverage_mode,omitempty"`
	Args []string `yaml:"args,omitempty"`
}

// kim: TODO: convert model into shrub config and run

// const (
//     GolangCoverageModeSet    = "set"
//     GolangCoverageModeCount  = "count"
//     GolangCoverageModeAtomic = "atomic"
// )

// ValidGolangCoverageMode
// func ValidGolangCoverageMode(mode string) bool {
//     switch {
//     case GolangCoverageModeSet, GolangCoverageModeCount, GolangCoverageModeAtomic:
//         return true
//     default:
//         return false
//     }
// }

func (opts *GolangRuntimeOptions) Validate() error {
	// catcher.NewWhen(opts.TimeoutSecs != nil && g.TimeoutSecs < 0, "timeout cannot be negative")
	// catcher.NewWhen(opts.RunCount != nil && g.RunCount < 0, "run count cannot be negative")
	// catcher.NewWhen(opts.CoverageMode != nil && ValidGolangCoverageMode(*g.CoverageMode), "invalid coverage mode '%s'", *g.CoverageMode)
	// kim: TODO: check for duplicate settings (e.g. args=["-race"] and RaceDetector
	//= true; args=["-covermode=count", "-covermode=set"])
	// TODO: should probably check for duplicate defined args.
	return nil
}

// func (opts *GolangRuntimeOptions) validateDuplicateArgsOptions() error {
//     isFlag := func(arg, flagName string) bool {
//         return arg == flagName || strings.HasPrefix(arg, flagName+"=")
//     }
//     catcher := grip.NewBasicCatcher()
//     for _, arg := range opts.Args {
//         // catcher.NewWhen(isFlag(arg, "-run") && g.Pattern != nil && *g.Pattern, "cannot specify the '-run' flag and also specify a run pattern")
//         // catcher.NewWhen(isFlag(arg, "-race") && g.RaceDetector != nil && *g.RaceDetector, "cannot specify the '-race' flag in args and also set race detector to true")
//         // catcher.NewWhen(isFlag(arg, "-cover") && g.Coverage != nil && *g.Coverage, "cannot specify the '-cover' flag in args and also set coverage to true")
//         // catcher.NewWhen(isFlag(arg, "-short") && g.Short != nil && *g.Short, "cannot specify the '-short' flag in args and also set short to true")
//         // catcher.NewWhen(isFlag(arg, "-covermode") && g.CoverageMode != nil && *g.CoverageMode != "", "cannot specify the '-coveragemode'"
//     }
//     return catcher.Resolve()
// }

// TODO: this is probably not the best behavior since ideally it should
// overwrite options specified in opts, but it's easiest to deal with for now.
func (opts *GolangRuntimeOptions) Merge(toOverwrite GolangRuntimeOptions) GolangRuntimeOptions {
	if len(opts.Args) != 0 {
		toOverwrite.Args = opts.Args
	}
	return toOverwrite
	// if g.Pattern != nil {
	//     opts.Pattern = g.Pattern
	// }
	// if g.TimeoutSecs != nil {
	//     opts.TimeoutSecs = g.TimeoutSecs
	// }
	// if g.RunCount != nil {
	//     opts.RunCount = g.RunCount
	// }
	// if g.CoverageMode != nil {
	//     opts.CoverageMode = g.CoverageMode
	// }
	// if g.RaceDetector != nil {
	//     opts.RaceDetector = g.RaceDetector
	// }
	// if g.Short != nil {
	//     opts.Short = g.Short
	// }
	// if g.Coverage != nil {
	//     opts.Coverage = g.Coverage
	// }
	// if g.CoverageMode != nil {
	//     opts.CoverageMode = g.CoverageMode
	// }
	// if g.Args != nil {
	//     opts.Args = g.Args
	// }
}

// kim: TODO: maybe get rid of and just have a merge function or something.
// kim: TODO: need to handle overwriting args by looking for matching prefix
// - <flag> =
// or exact flag match
// func (g *GolangRuntimeOptions) TestOptions() scripting.TestOptions {
//     var args []string
//     if p.RaceDetector != nil && *p.RaceDetector {
//         args = append(args, "-race")
//     }
//     if p.Short != nil && *p.Short {
//         args = append(args, "-short")
//     }
//     if p.Coverage != nil && *p.Coverage {
//         args = append(args, "-cover")
//     }
//     if p.CoverageMode != nil && p.CoverageMode != "" {
//         args = append(args, fmt.Sprintf("-covermode=%s", p.CoverageMode))
//     }
//     return scripting.TestOptions{
//         Name:    p.Name,
//         Timeout: p.TimeoutSecs * time.Second,
//         Pattern: p.Pattern,
//         Count:   p.RunCount,
//         Args:    args,
//     }
// }
