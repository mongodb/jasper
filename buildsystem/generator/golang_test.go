package generator

import (
	"path/filepath"
	"testing"

	"github.com/evergreen-ci/shrub"
	"github.com/mongodb/jasper/buildsystem/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: more test coverage for generation
func TestGolangGenerate(t *testing.T) {
	for testName, testCase := range map[string]func(t *testing.T, g *Golang){
		"Succeeds": func(t *testing.T, g *Golang) {
			conf, err := g.Generate()
			require.NoError(t, err)
			expected := [][]string{
				{"variant1", "path1"},
				{"variant1", "name2"},
				{"variant2", "name2"},
			}
			require.Len(t, conf.Tasks, len(expected))

			for _, parts := range expected {
				task := conf.Task(getTaskName(parts...))
				require.Len(t, task.Commands, 2)

				getProjectCmd := task.Commands[0]
				assert.Equal(t, shrub.CmdGetProject{}.Name(), getProjectCmd.CommandName)
				projectPath, err := g.RelProjectPath()
				require.NoError(t, err)
				assert.Equal(t, projectPath, getProjectCmd.Params["directory"])

				scriptingCmd := task.Commands[1]
				assert.Equal(t, shrub.CmdSubprocessScripting{}.Name(), scriptingCmd.CommandName)
				gopath, err := g.RelGopath()
				require.NoError(t, err)
				assert.Equal(t, gopath, scriptingCmd.Params["harness_path"])
				assert.Equal(t, g.WorkingDirectory, scriptingCmd.Params["working_dir"])
				assert.Equal(t, projectPath, scriptingCmd.Params["test_dir"])
				env, ok := scriptingCmd.Params["env"].(map[string]interface{})
				require.True(t, ok)
				assert.EqualValues(t, g.Environment["GOROOT"], env["GOROOT"])
			}
		},
	} {
		t.Run(testName, func(t *testing.T) {
			rootPackage := filepath.Join("github.com", "fake_user", "fake_repo")
			gopath := "gopath"

			g := Golang{
				Golang: model.Golang{
					Environment: map[string]string{
						"GOPATH": gopath,
						"GOROOT": "some_goroot",
					},
					RootPackage: rootPackage,
					Packages: []model.GolangPackage{
						{
							Path: "path1",
						},
						{
							Name: "name2",
							Path: "path2",
						},
					},
					Variants: []model.GolangVariant{
						{
							VariantDistro: model.VariantDistro{
								Name:    "variant1",
								Distros: []string{"distro1"},
							},
							GolangVariantParameters: model.GolangVariantParameters{
								Packages: []model.GolangVariantPackage{
									{Path: "path1"},
									{Name: "name2"},
								},
							},
						},
						{
							VariantDistro: model.VariantDistro{
								Name:    "variant2",
								Distros: []string{"distro2"},
							},
							GolangVariantParameters: model.GolangVariantParameters{
								Packages: []model.GolangVariantPackage{
									{Name: "name2"},
								},
							},
						},
					},
					WorkingDirectory: filepath.Dir(gopath),
				},
			}
			testCase(t, &g)
		})
	}
}
