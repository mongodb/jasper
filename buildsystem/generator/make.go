package generator

import (
	"github.com/evergreen-ci/shrub"
	"github.com/mongodb/jasper/buildsystem/model"
	"github.com/pkg/errors"
)

// Make represents an evergreen config generator for Make-based projects.
type Make struct {
	model.Make
	WorkingDirectory string `yaml:"-"`
}

// NewMake returns a generator for Make.
func NewMake(m model.Make, workingDir string) *Make {
	return &Make{
		Make:             m,
		WorkingDirectory: workingDir,
	}
}

func (m *Make) Generate() (*shrub.Configuration, error) {
	conf, err := shrub.BuildConfiguration(func(c *shrub.Configuration) {
		for _, mv := range m.Variants {
			variant := c.Variant(mv.Name)
			variant.DistroRunOn = mv.Distros

			var tasksForVariant []*shrub.Task
			// kim: TODO: turn each task into a function and each
			// variant-specific task passes in its own vars?
			for _, mvt := range mv.Tasks {
				tasks, err := m.GetTasksAndRef(mvt)
				if err != nil {
					panic(err)
				}
				tasksForVariant = append(tasksForVariant, m.generateVariantTasksForRef(c, mv, tasks)...)
			}

			// kim: TODO: turn into a function?
			getProjectCmd := shrub.CmdGetProject{
				Directory: m.WorkingDirectory,
			}

			if len(variant.TaskSpecs) >= minTasksForTaskGroup {
				tg := c.TaskGroup(mv.Name + "_group").SetMaxHosts(len(variant.TaskSpecs) / 2)
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

func (m *Make) generateVariantTasksForRef(c *shrub.Configuration, mv model.MakeVariant, mts []model.MakeTask) []*shrub.Task {
	var tasks []*shrub.Task
	for _, mt := range mts {
		cmds := m.subprocessExecCmds(mv, mt)
		tasks = append(tasks, c.Task(getTaskName(mv.Name, mt.Name)).Command(cmds...))
	}
	return tasks
}

func (m *Make) subprocessExecCmds(mv model.MakeVariant, mt model.MakeTask) []shrub.Command {
	env := model.MergeEnvironments(m.Environment, mv.Environment, mt.Environment)
	var cmds []shrub.Command
	for _, target := range mt.Targets {
		cmds = append(cmds, &shrub.CmdExec{
			Binary: "make",
			// kim: TODO: people should maybe be able to specify additional
			// make args for tasks.
			Args: []string{target},
			Env:  env,
		})
	}
	return cmds
}
