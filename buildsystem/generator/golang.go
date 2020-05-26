package generator

import (
	"strings"

	"github.com/evergreen-ci/shrub"
	"github.com/mongodb/jasper/buildsystem/model"
	"github.com/pkg/errors"
)

// Golang represents a configuration for generating an evergreen configuration
// from a project that uses golang.
type Golang struct {
	model.Golang
}

func NewGolang(g model.Golang) *Golang {
	return &Golang{
		Golang: g,
	}
}

// Generate creates the evergreen configuration from the given golang build
// configuration.
func (g *Golang) Generate() (*shrub.Configuration, error) {
	conf, err := shrub.BuildConfiguration(func(c *shrub.Configuration) {
		for _, gv := range g.Variants {
			variant := c.Variant(gv.Name)
			variant.DistroRunOn = gv.Distros

			var tasksForVariant []*shrub.Task
			// Make one task per package in this variant. We cannot make one
			// task per package, because we have to account for variant-level
			// options possibly overriding package-level options, which requires
			// making separate tasks with different commands.
			for _, gvp := range gv.Packages {
				var gps []model.GolangPackage
				var pkgRef string
				gps, pkgRef, err := g.getPackagesAndRef(gvp)
				if err != nil {
					panic(errors.Wrapf(err, "package definition for variant '%s'", gv.Name))
				}

				newTasks, err := g.generateVariantTasksForRef(c, gv, gps, pkgRef)
				if err != nil {
					panic(errors.Wrapf(err, "generating task for package ref '%s' in variant '%s'", pkgRef, gv.Name))
				}
				tasksForVariant = append(tasksForVariant, newTasks...)
			}

			projectPath, err := g.RelProjectPath()
			if err != nil {
				panic(errors.Wrap(err, "getting project path as a relative path"))
			}
			getProjectCmd := shrub.CmdGetProject{
				Directory: projectPath,
			}

			// Only use a task group for this variant if it meets the threshold
			// number of tasks. Otherwise, just run regular tasks for this
			// variant.
			if len(variant.TaskSpecs) >= minTasksForTaskGroup {
				tg := c.TaskGroup(gv.Name + "_group").SetMaxHosts(len(variant.TaskSpecs) / 2)
				tg.SetupTask = shrub.CommandSequence{getProjectCmd.Resolve()}

				for _, task := range variant.TaskSpecs {
					_ = tg.Task(task.Name)
				}
				_ = variant.AddTasks(tg.GroupName)
			} else {
				for _, task := range tasksForVariant {
					task.Commands = append([]*shrub.CommandDefinition{getProjectCmd.Resolve()}, task.Commands...)
					_ = variant.AddTasks(task.Name)
				}
			}
		}
	})

	if err != nil {
		return nil, errors.Wrap(err, "generating evergreen configuration")
	}

	return conf, nil
}

func (g *Golang) getPackagesAndRef(gvp model.GolangVariantPackage) ([]model.GolangPackage, string, error) {
	if gvp.Tag != "" {
		gps := g.GetPackagesByTag(gvp.Tag)
		if len(gps) == 0 {
			return nil, "", errors.Errorf("no packages matched tag '%s'", gvp.Tag)
		}
		return gps, gvp.Tag, nil
	}

	if gvp.Name != "" {
		gp, _, err := g.GetPackageIndexByName(gvp.Name)
		if err != nil {
			return nil, "", errors.Wrapf(err, "finding definition for package named '%s'", gvp.Name)
		}
		return []model.GolangPackage{*gp}, gvp.Name, nil
	} else if gvp.Path != "" {
		gp, _, err := g.GetUnnamedPackageIndexByPath(gvp.Path)
		if err != nil {
			return nil, "", errors.Wrapf(err, "finding definition for package path '%s'", gvp.Path)
		}
		return []model.GolangPackage{*gp}, gvp.Path, nil
	}

	return nil, "", errors.New("empty package reference")
}

func (g *Golang) generateVariantTasksForRef(c *shrub.Configuration, gv model.GolangVariant, gps []model.GolangPackage, pkgRef string) ([]*shrub.Task, error) {
	var tasks []*shrub.Task

	for _, gp := range gps {
		scriptCmd, err := g.subprocessScriptingCmd(gv, gp)
		if err != nil {
			return nil, errors.Wrapf(err, "generating %s command for package '%s' in variant '%s'", shrub.CmdSubprocessScripting{}.Name(), gp.Path, gv.Name)
		}
		var taskName string
		if len(gps) > 1 {
			taskName = getTaskName(gv.Name, pkgRef, gp.Path)
		} else {
			taskName = getTaskName(gv.Name, pkgRef)
		}
		tasks = append(tasks, c.Task(taskName).Command(scriptCmd))
	}

	return tasks, nil
}

func (g *Golang) subprocessScriptingCmd(gv model.GolangVariant, gp model.GolangPackage) (*shrub.CmdSubprocessScripting, error) {
	gopath, err := g.RelGopath()
	if err != nil {
		return nil, errors.Wrap(err, "getting GOPATH as a relative path")
	}
	projectPath, err := g.RelProjectPath()
	if err != nil {
		return nil, errors.Wrap(err, "getting project path as a relative path")
	}

	testOpts := gp.Options
	if gv.Options != nil {
		testOpts = testOpts.Merge(*gv.Options)
	}

	relPath := gp.Path
	if relPath != "." && !strings.HasPrefix(relPath, "./") {
		relPath = "./" + relPath
	}
	testOpts = append(testOpts, relPath)

	return &shrub.CmdSubprocessScripting{
		Harness:     "golang",
		WorkingDir:  g.WorkingDirectory,
		HarnessPath: gopath,
		// It is not a problem for the environment to set the
		// GOPATH here (relative to the working directory),
		// which conflicts with the actual GOPATH (an absolute
		// path). The GOPATH in the environment will be
		// overridden when subprocess.scripting runs to be an
		// absolute path relative to the working directory.
		Env:     g.Environment,
		TestDir: projectPath,
		TestOptions: &shrub.ScriptingTestOptions{
			Args: testOpts,
		},
	}, nil
}
