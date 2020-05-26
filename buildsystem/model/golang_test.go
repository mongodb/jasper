package model

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/evergreen-ci/utility"
	"github.com/mongodb/jasper/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGolangVariantPackage(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		t.Run("FailsForNoRef", func(t *testing.T) {
			vp := GolangVariantPackage{}
			assert.Error(t, vp.Validate())
		})
		t.Run("SucceedsIfNameSet", func(t *testing.T) {
			vp := GolangVariantPackage{Name: "name"}
			assert.NoError(t, vp.Validate())
		})
		t.Run("SucceedsIfPathSet", func(t *testing.T) {
			vp := GolangVariantPackage{Path: "path"}
			assert.NoError(t, vp.Validate())
		})
		t.Run("SucceedsIfTagSet", func(t *testing.T) {
			vp := GolangVariantPackage{Tag: "tag"}
			assert.NoError(t, vp.Validate())
		})
		t.Run("FailsIfNameAndPathSet", func(t *testing.T) {
			vp := GolangVariantPackage{
				Name: "name",
				Path: "path",
			}
			assert.Error(t, vp.Validate())
		})
		t.Run("FailsIfNameAndTagSet", func(t *testing.T) {
			vp := GolangVariantPackage{
				Name: "name",
				Tag:  "tag",
			}
			assert.Error(t, vp.Validate())
		})
		t.Run("FailsIfPathAndTagSet", func(t *testing.T) {
			vp := GolangVariantPackage{
				Path: "path",
				Tag:  "tag",
			}
			assert.Error(t, vp.Validate())
		})
		t.Run("FailsIfAllSet", func(t *testing.T) {
			vp := GolangVariantPackage{
				Name: "name",
				Path: "path",
				Tag:  "tag",
			}
			assert.Error(t, vp.Validate())
		})
	})
}

func TestGolangVariant(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		for testName, testCase := range map[string]func(t *testing.T, v *GolangVariant){
			"Passes": func(t *testing.T, gv *GolangVariant) {
				assert.NoError(t, gv.Validate())
			},
			"FailsWithoutName": func(t *testing.T, gv *GolangVariant) {
				gv.Name = ""
				assert.Error(t, gv.Validate())
			},
			"FailsWithoutDistros": func(t *testing.T, gv *GolangVariant) {
				gv.Distros = nil
				assert.Error(t, gv.Validate())
			},
			"FailsWithoutPackages": func(t *testing.T, gv *GolangVariant) {
				gv.Packages = nil
				assert.Error(t, gv.Validate())
			},
			"FailsForInvalidPackage": func(t *testing.T, gv *GolangVariant) {
				gv.Packages = []GolangVariantPackage{{}}
				assert.Error(t, gv.Validate())
			},
			"FailsForDuplicatePackageName": func(t *testing.T, gv *GolangVariant) {
				gv.Packages = []GolangVariantPackage{
					{Name: "name"},
					{Name: "name"},
				}
				assert.Error(t, gv.Validate())
			},
			"FailsForDuplicatePackagePath": func(t *testing.T, gv *GolangVariant) {
				gv.Packages = []GolangVariantPackage{
					{Path: "path"},
					{Path: "path"},
				}
				assert.Error(t, gv.Validate())
			},
			"FailsForDuplicatePackageTag": func(t *testing.T, gv *GolangVariant) {
				gv.Packages = []GolangVariantPackage{
					{Tag: "tag"},
					{Tag: "tag"},
				}
				assert.Error(t, gv.Validate())
			},
		} {
			t.Run(testName, func(t *testing.T) {
				gv := GolangVariant{
					VariantDistro: VariantDistro{
						Name:    "var_name",
						Distros: []string{"distro1", "distro2"},
					},
					GolangVariantParameters: GolangVariantParameters{
						Packages: []GolangVariantPackage{
							{Name: "name"},
							{Path: "path"},
							{Tag: "tag"},
						},
					},
				}
				testCase(t, &gv)
			})
		}
	})
}

func TestGolangPackage(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		for testName, testCase := range map[string]func(t *testing.T, gp *GolangPackage){
			"Passes": func(t *testing.T, gp *GolangPackage) {
				assert.NoError(t, gp.Validate())
			},
			"FailsWithoutPath": func(t *testing.T, gp *GolangPackage) {
				gp.Path = ""
				assert.Error(t, gp.Validate())
			},
			"FailsWithDuplicateTags": func(t *testing.T, gp *GolangPackage) {
				gp.Tags = []string{"tag1", "tag1"}
			},
		} {
			t.Run(testName, func(t *testing.T) {
				gp := GolangPackage{
					Name: "name",
					Path: "path",
					Tags: []string{"tag1"},
				}
				testCase(t, &gp)
			})
		}
	})
}

func TestGolangGetPackageIndexByName(t *testing.T) {
	for testName, testCase := range map[string]func(t *testing.T, g *Golang){
		"Succeeds": func(t *testing.T, g *Golang) {
			gp, i, err := g.GetPackageIndexByName("package1")
			require.NoError(t, err)
			assert.Equal(t, 0, i)
			assert.Equal(t, "package1", gp.Name)
		},
		"FailsForPackageNotFound": func(t *testing.T, g *Golang) {
			gp, i, err := g.GetPackageIndexByName("")
			assert.Error(t, err)
			assert.Equal(t, -1, i)
			assert.Nil(t, gp)
		},
	} {
		t.Run(testName, func(t *testing.T) {
			g := Golang{
				Packages: []GolangPackage{
					{Name: "package1"},
					{Name: "package2"},
				},
			}
			testCase(t, &g)
		})
	}
}

func TestGolangGetUnnamedPackagesByPath(t *testing.T) {
	for testName, testCase := range map[string]func(t *testing.T, g *Golang){
		"Succeeds": func(t *testing.T, g *Golang) {
			gp, i, err := g.GetUnnamedPackageIndexByPath("path1")
			require.NoError(t, err)
			assert.Equal(t, 0, i)
			assert.Equal(t, "path1", gp.Path)
		},
		"FailsForPackageNotFound": func(t *testing.T, g *Golang) {
			gp, i, err := g.GetUnnamedPackageIndexByPath("")
			assert.Error(t, err)
			assert.Equal(t, -1, i)
			assert.Nil(t, gp)
		},
		"FailsForNamedPackageWithPath": func(t *testing.T, g *Golang) {
			gp, i, err := g.GetUnnamedPackageIndexByPath("path3")
			assert.Error(t, err)
			assert.Equal(t, -1, i)
			assert.Nil(t, gp)
		},
	} {
		t.Run(testName, func(t *testing.T) {
			g := Golang{
				Packages: []GolangPackage{
					{Path: "path1"},
					{Path: "path2"},
					{Name: "name3", Path: "path3"},
				},
			}
			testCase(t, &g)
		})
	}
}

func TestGolangGetPackagesByTag(t *testing.T) {
	for testName, testCase := range map[string]func(t *testing.T, g *Golang){
		"Succeeds": func(t *testing.T, g *Golang) {
			gps := g.GetPackagesByTag("tag2")
			require.Len(t, gps, 1)
			assert.Equal(t, "path1", gps[0].Path)
		},
		"FailsForPackageNotFound": func(t *testing.T, g *Golang) {
			gps := g.GetPackagesByTag("")
			assert.Empty(t, gps)
		},
	} {
		t.Run(testName, func(t *testing.T) {
			g := Golang{
				Packages: []GolangPackage{
					{Path: "path1", Tags: []string{"tag1", "tag2"}},
					{Path: "path2", Tags: []string{"tag1", "tag3"}},
				},
			}
			testCase(t, &g)
		})
	}
}

func TestGolangValidate(t *testing.T) {
	for testName, testCase := range map[string]func(t *testing.T, g *Golang){
		"Passes": func(t *testing.T, g *Golang) {
			assert.NoError(t, g.Validate())
		},
		"FailsWithoutRootPackage": func(t *testing.T, g *Golang) {
			g.RootPackage = ""
			assert.Error(t, g.Validate())
		},
		"PassesWithGOROOTInEnvironment": func(t *testing.T, g *Golang) {
			goroot := os.Getenv("GOROOT")
			if goroot == "" {
				t.Skip("GOROOT is not defined in environment")
			}
			delete(g.Environment, "GOROOT")
			assert.NoError(t, g.Validate())
			assert.Equal(t, goroot, g.Environment["GOROOT"])
		},
		"FailsWithoutGOROOTEnvVar": func(t *testing.T, g *Golang) {
			if goroot, ok := os.LookupEnv("GOROOT"); ok {
				defer func() {
					os.Setenv("GOROOT", goroot)
				}()
				require.NoError(t, os.Unsetenv("GOROOT"))
			}
			delete(g.Environment, "GOROOT")
			assert.Error(t, g.Validate())
		},
		"PassesIfGOPATHInEnvironmentAndIsWithinWorkingDirectory": func(t *testing.T, g *Golang) {
			gopath := os.Getenv("GOPATH")
			if gopath == "" {
				t.Skip("GOPATH not defined in environment")
			}
			g.WorkingDirectory = filepath.Dir(gopath)
			delete(g.Environment, "GOPATH")
			assert.NoError(t, g.Validate())
			relGopath, err := filepath.Rel(g.WorkingDirectory, gopath)
			require.NoError(t, err)
			assert.Equal(t, relGopath, g.Environment["GOPATH"])
		},
		"FailsWithoutGOPATHEnvVar": func(t *testing.T, g *Golang) {
			if gopath, ok := os.LookupEnv("GOPATH"); ok {
				defer func() {
					os.Setenv("GOPATH", gopath)
				}()
				require.NoError(t, os.Unsetenv("GOPATH"))
			}
			delete(g.Environment, "GOPATH")
			assert.Error(t, g.Validate())
		},
		"FailsWithoutPackages": func(t *testing.T, g *Golang) {
			g.Packages = nil
			assert.Error(t, g.Validate())
		},
		"FailsWithInvalidPackage": func(t *testing.T, g *Golang) {
			g.Packages = []GolangPackage{{}}
			assert.Error(t, g.Validate())
		},
		"PassesWithUniquePackageNames": func(t *testing.T, g *Golang) {
			g.Packages = []GolangPackage{
				{Path: "path1"},
				{Name: "name2", Path: "path2"},
			}
			assert.NoError(t, g.Validate())
		},
		"FailsWithDuplicatePackageName": func(t *testing.T, g *Golang) {
			g.Packages = []GolangPackage{
				{Name: "name", Path: "path1"},
				{Name: "name", Path: "path2"},
			}
			assert.Error(t, g.Validate())
		},
		"PassesWithUniquePackagePaths": func(t *testing.T, g *Golang) {
			g.Packages = []GolangPackage{
				{Path: "path1"},
				{Path: "path2"},
			}
			assert.NoError(t, g.Validate())
		},
		"PassesWithDuplicatePackagePathButUniqueNames": func(t *testing.T, g *Golang) {
			g.Packages = []GolangPackage{
				{Name: "name1", Path: "path1"},
				{Name: "name2", Path: "path1"},
			}
			g.Variants = []GolangVariant{
				{
					VariantDistro: VariantDistro{
						Name:    "variant",
						Distros: []string{"distro"},
					},
					GolangVariantParameters: GolangVariantParameters{
						Packages: []GolangVariantPackage{
							{Name: "name1"},
							{Name: "name2"},
						},
					},
				},
			}
			assert.NoError(t, g.Validate())
		},
		"FailsWithDuplicateUnnamedPackagePath": func(t *testing.T, g *Golang) {
			g.Packages = []GolangPackage{
				{Path: "path"},
				{Path: "path"},
			}
			assert.Error(t, g.Validate())
		},
		"FailsWithoutVariants": func(t *testing.T, g *Golang) {
			g.Variants = nil
			assert.Error(t, g.Validate())
		},
		"FailsWithInvalidVariant": func(t *testing.T, g *Golang) {
			g.Variants = []GolangVariant{{}}
			assert.Error(t, g.Validate())
		},
		"FailsWithDuplicateVariantName": func(t *testing.T, g *Golang) {
			g.Variants = []GolangVariant{
				{
					VariantDistro: VariantDistro{
						Name:    "variant",
						Distros: []string{"distro"},
					},
					GolangVariantParameters: GolangVariantParameters{
						Packages: []GolangVariantPackage{
							{Path: "path1"},
						},
					},
				},
				{
					VariantDistro: VariantDistro{
						Name:    "variant",
						Distros: []string{"distro"},
					},
					GolangVariantParameters: GolangVariantParameters{
						Packages: []GolangVariantPackage{
							{Path: "path2"},
						},
					},
				},
			}
		},
		"PassesWithValidGolangVariantPackageName": func(t *testing.T, g *Golang) {
			g.Packages = []GolangPackage{
				{Name: "name", Path: "path"},
			}
			g.Variants = []GolangVariant{
				{
					VariantDistro: VariantDistro{
						Name:    "variant",
						Distros: []string{"distro"},
					},
					GolangVariantParameters: GolangVariantParameters{
						Packages: []GolangVariantPackage{
							{Name: "name"},
						},
					},
				},
			}
			assert.NoError(t, g.Validate())
		},
		"FailsWithInvalidGolangVariantPackageName": func(t *testing.T, g *Golang) {
			g.Variants = []GolangVariant{
				{
					VariantDistro: VariantDistro{
						Name:    "variant",
						Distros: []string{"distro"},
					},
					GolangVariantParameters: GolangVariantParameters{
						Packages: []GolangVariantPackage{
							{Name: "nonexistent"},
						},
					},
				},
			}
			assert.Error(t, g.Validate())
		},
		"PassesWithValidGolangVariantPackagePath": func(t *testing.T, g *Golang) {
			g.Packages = []GolangPackage{
				{Path: "path"},
			}
			g.Variants = []GolangVariant{
				{
					VariantDistro: VariantDistro{
						Name:    "variant",
						Distros: []string{"distro"},
					},
					GolangVariantParameters: GolangVariantParameters{
						Packages: []GolangVariantPackage{
							{Path: "path"},
						},
					},
				},
			}
			assert.NoError(t, g.Validate())
		},
		"FailsWithDuplicateGolangPackageReferences": func(t *testing.T, g *Golang) {
			g.Packages = []GolangPackage{
				{Path: "path", Tags: []string{"tag"}},
			}
			g.Variants = []GolangVariant{
				{
					VariantDistro: VariantDistro{
						Name:    "variant",
						Distros: []string{"distro"},
					},
					GolangVariantParameters: GolangVariantParameters{
						Packages: []GolangVariantPackage{
							{Path: "path"},
							{Tag: "tag"},
						},
					},
				},
			}
			assert.Error(t, g.Validate())
		},
		"FailsWithInvalidGolangVariantPackagePath": func(t *testing.T, g *Golang) {
			g.Variants = []GolangVariant{
				{
					VariantDistro: VariantDistro{
						Name:    "variant",
						Distros: []string{"distro"},
					},
					GolangVariantParameters: GolangVariantParameters{
						Packages: []GolangVariantPackage{
							{Path: "nonexistent"},
						},
					},
				},
			}
			assert.Error(t, g.Validate())
		},
		"PassesWithValidGolangVariantPackageTag": func(t *testing.T, g *Golang) {
			g.Packages = []GolangPackage{
				{Path: "path", Tags: []string{"tag"}},
			}
			g.Variants = []GolangVariant{
				{
					VariantDistro: VariantDistro{
						Name:    "variant",
						Distros: []string{"distro"},
					},
					GolangVariantParameters: GolangVariantParameters{
						Packages: []GolangVariantPackage{
							{Tag: "tag"},
						},
					},
				},
			}
			assert.NoError(t, g.Validate())
		},
		"FailsWithInvalidGolangVariantPackageTag": func(t *testing.T, g *Golang) {
			g.Variants = []GolangVariant{
				{
					VariantDistro: VariantDistro{
						Name:    "variant",
						Distros: []string{"distro"},
					},
					GolangVariantParameters: GolangVariantParameters{
						Packages: []GolangVariantPackage{
							{Tag: "nonexistent"},
						},
					},
				},
			}
			assert.Error(t, g.Validate())
		},
	} {
		t.Run(testName, func(t *testing.T) {
			g := Golang{
				RootPackage: "root_package",
				Environment: map[string]string{
					"GOPATH": "gopath",
					"GOROOT": "goroot",
				},
				Packages: []GolangPackage{
					{Path: "path1"},
					{Path: "path2"},
				},
				Variants: []GolangVariant{
					{
						VariantDistro: VariantDistro{
							Name:    "variant",
							Distros: []string{"distro"},
						},
						GolangVariantParameters: GolangVariantParameters{
							Packages: []GolangVariantPackage{
								{Path: "path1"},
							},
						},
					},
				},
			}
			testCase(t, &g)
		})
	}
}

func TestDiscoverPackages(t *testing.T) {
	for testName, testCase := range map[string]func(t *testing.T, g *Golang, rootPath string){
		"FailsForPackageNotFound": func(t *testing.T, g *Golang, rootPath string) {
			g.RootPackage = "foo"
			assert.Error(t, g.DiscoverPackages())
		},
		"DoesNotDiscoverPackageWithoutTestFiles": func(t *testing.T, g *Golang, rootPath string) {
			assert.NoError(t, g.DiscoverPackages())
			assert.Empty(t, g.Packages)
		},
		"DiscoversPackageIfTestFilesPresent": func(t *testing.T, g *Golang, rootPath string) {
			f, err := os.Create(filepath.Join(rootPath, "fake_test.go"))
			require.NoError(t, err)
			require.NoError(t, f.Close())

			assert.NoError(t, g.DiscoverPackages())
			require.Len(t, g.Packages, 1)
			assert.Equal(t, ".", g.Packages[0].Path)
			assert.Empty(t, g.Packages[0].Name)
			assert.Empty(t, g.Packages[0].Tags)
		},
		"DoesNotModifyPackageDefinitionIfAlreadyDefined": func(t *testing.T, g *Golang, rootPath string) {
			gp := GolangPackage{
				Name: "package_name",
				Path: ".",
				Tags: []string{"tag"},
			}
			g.Packages = []GolangPackage{gp}
			f, err := os.Create(filepath.Join(rootPath, "fake_test.go"))
			require.NoError(t, err)
			require.NoError(t, f.Close())

			assert.NoError(t, g.DiscoverPackages())
			require.Len(t, g.Packages, 1)
			assert.Equal(t, gp.Path, g.Packages[0].Path)
			assert.Equal(t, gp.Name, g.Packages[0].Name)
			assert.Equal(t, gp.Tags, g.Packages[0].Tags)
		},
		"IgnoresVendorDirectory": func(t *testing.T, g *Golang, rootPath string) {
			vendorDir := filepath.Join(rootPath, golangVendorDir)
			require.NoError(t, os.Mkdir(vendorDir, 0777))
			f, err := os.Create(filepath.Join(vendorDir, "fake_test.go"))
			require.NoError(t, err)
			require.NoError(t, f.Close())

			assert.NoError(t, g.DiscoverPackages())
			assert.Empty(t, g.Packages)
		},
	} {
		t.Run(testName, func(t *testing.T) {
			rootPackage := filepath.Join("github.com", "fake_user", "fake_repo")
			gopath, err := ioutil.TempDir(testutil.BuildDirectory(), "gopath")
			require.NoError(t, err)
			defer func() {
				assert.NoError(t, os.RemoveAll(gopath))
			}()
			rootPath := filepath.Join(gopath, "src", rootPackage)
			require.NoError(t, os.MkdirAll(rootPath, 0777))

			g := Golang{
				Environment: map[string]string{
					"GOPATH": gopath,
					"GOROOT": "some_goroot",
				},
				RootPackage:      rootPackage,
				WorkingDirectory: filepath.Dir(gopath),
			}
			testCase(t, &g, rootPath)
		})
	}
}

func TestGolangRuntimeOptions(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		for testName, testCase := range map[string]struct {
			opts        GolangRuntimeOptions
			expectError bool
		}{
			"PassesWithAllUniqueFlags": {
				opts: []string{"-cover", "-coverprofile", "-race"},
			},
			"FailsWithDuplicateFlags": {
				opts:        []string{"-race", "-race"},
				expectError: true,
			},
			"FailsWithDuplicateEquivalentFlags": {
				opts:        []string{"-race", "-test.race"},
				expectError: true,
			},
			"FailsWithVerboseFlag": {
				opts:        []string{"-v"},
				expectError: true,
			},
		} {
			t.Run(testName, func(t *testing.T) {
				err := testCase.opts.Validate()
				if testCase.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
	t.Run("Merge", func(t *testing.T) {
		for testName, testCase := range map[string]struct {
			opts      GolangRuntimeOptions
			overwrite GolangRuntimeOptions
			expected  GolangRuntimeOptions
		}{
			"AllUniqueFlagsAreAppended": {
				opts:      []string{"-cover", "-race=true"},
				overwrite: []string{"-coverprofile", "-outputdir=./dir"},
				expected:  []string{"-cover", "-race=true", "-coverprofile", "-outputdir=./dir"},
			},
			"DuplicateFlagsAreCombined": {
				opts:      []string{"-cover"},
				overwrite: []string{"-cover"},
				expected:  []string{"-cover"},
			},
			"TestFlagsAreCheckedAgainstEquivalentFlags": {
				opts:      []string{"-test.race"},
				overwrite: []string{"-race"},
				expected:  []string{"-race"},
			},
			"ConflictingTestFlagsAreOverwritten": {
				opts:      []string{"-test.race=true"},
				overwrite: []string{"-test.race=false"},
				expected:  []string{"-test.race=false"},
			},
			"UniqueFlagsAreAppendedAndDuplicateFlagsAreCombined": {
				opts:      []string{"-cover"},
				overwrite: []string{"-cover", "-coverprofile"},
				expected:  []string{"-cover", "-coverprofile"},
			},
			"ConflictingFlagValuesAreOverwritten": {
				opts:      []string{"-race=false"},
				overwrite: []string{"-race=true"},
				expected:  []string{"-race=true"},
			},
			"UniqueFlagsAreAppendedAndConflictingFlagsAreOverwritten": {
				opts:      []string{"-cover", "-race=false"},
				overwrite: []string{"-race=true"},
				expected:  []string{"-cover", "-race=true"},
			},
			"DuplicateFlagsAreCombinedAndConflictingFlagsAreOverwritten": {
				opts:      []string{"-cover", "-race=false"},
				overwrite: []string{"-cover", "-race=true"},
				expected:  []string{"-cover", "-race=true"},
			},
		} {
			t.Run(testName, func(t *testing.T) {
				merged := testCase.opts.Merge(testCase.overwrite)
				assert.Len(t, merged, len(testCase.expected))
				for _, flag := range merged {
					assert.True(t, utility.StringSliceContains(testCase.expected, flag))
				}
			})
		}

	})
}
