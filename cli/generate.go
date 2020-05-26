package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/mongodb/jasper/buildsystem/generator"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func Generate() cli.Command {
	return cli.Command{
		Name:  "generate",
		Usage: "Generate JSON evergreen configurations.",
		Subcommands: []cli.Command{
			generateGolang(),
			generateMake(),
		},
	}
}

const (
	workingDirFlagName    = "working_dir"
	generatorFileFlagName = "generator_file"
	controlFileFlagName   = "control_file"
	outputFileFlagName    = "output_file"
)

func generateGolang() cli.Command {
	return cli.Command{
		Name:  "golang",
		Usage: "Generate JSON evergreen config from golang build file(s).",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  workingDirFlagName,
				Usage: "The directory that contains the GOPATH as a subdirectory.",
			},
			cli.StringFlag{
				Name:  generatorFileFlagName,
				Usage: "The build files necessary to generate the evergreen config.",
			},
			cli.StringFlag{
				Name:  controlFileFlagName,
				Usage: "The control file referencing all the necessary build files.",
			},
			cli.StringFlag{
				Name:  outputFileFlagName,
				Usage: "The output file. If unspecified, the config will be written to stdout.",
			},
		},
		Before: mergeBeforeFuncs(
			requireStringFlag(workingDirFlagName),
			requireOneFlag(generatorFileFlagName, controlFileFlagName),
		),
		Action: func(c *cli.Context) error {
			workingDir := c.String(workingDirFlagName)
			genFile := c.String(generatorFileFlagName)
			ctrlFile := c.String(generatorFileFlagName)
			outputFile := c.String(outputFileFlagName)
			var err error
			if !filepath.IsAbs(workingDir) {
				workingDir, err = filepath.Abs(workingDir)
				if err != nil {
					return errors.Wrapf(err, "getting working directory '%s' as absolute path", workingDir)
				}
			}

			var gen *generator.Golang
			if genFile != "" {
				gen, err = generator.NewGolang(genFile, workingDir)
				if err != nil {
					return errors.Wrapf(err, "creating generator from build file '%s'", genFile)
				}
			} else if ctrlFile != "" {
				gen = &generator.Golang{}
				// kim: TODO: write Merge functions to gather all the parts
				// files
			}

			conf, err := gen.Generate()
			if err != nil {
				return errors.Wrap(err, "generating evergreen config from golang build file(s)")
			}

			output, err := json.MarshalIndent(conf, "", "\t")
			if err != nil {
				return errors.Wrap(err, "marshalling evergreen config as JSON")
			}
			if outputFile != "" {
				if err := ioutil.WriteFile(outputFile, output, 0644); err != nil {
					return errors.Wrapf(err, "writing JSON config to file '%s'", outputFile)
				}
			} else {
				fmt.Println(string(output))
			}

			return nil
		},
	}
}

func generateMake() cli.Command {
	return cli.Command{
		Name:  "make",
		Usage: "Generate JSON evergreen config from make build file(s).",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  workingDirFlagName,
				Usage: "The directory containing the project and build files.",
			},
			cli.StringFlag{
				Name:  generatorFileFlagName,
				Usage: "The build files necessary to generate the evergreen config (relative to the working directory).",
			},
			cli.StringFlag{
				Name:  controlFileFlagName,
				Usage: "The control file referencing all the necessary build files (relative to the working directory).",
			},
			cli.StringFlag{
				Name:  outputFileFlagName,
				Usage: "The output file. If unspecified, the config will be written to stdout.",
			},
		},
		Before: mergeBeforeFuncs(
			requireStringFlag(workingDirFlagName),
			requireOneFlag(generatorFileFlagName, controlFileFlagName),
			cleanupFilePathSeparators(generatorFileFlagName, controlFileFlagName, workingDirFlagName),
			requireRelativePath(generatorFileFlagName, workingDirFlagName),
			requireRelativePath(controlFileFlagName, workingDirFlagName),
		),
		Action: func(c *cli.Context) error {
			workingDir := c.String(workingDirFlagName)
			genFile := c.String(generatorFileFlagName)
			ctrlFile := c.String(controlFileFlagName)
			outputFile := c.String(outputFileFlagName)

			var gen *generator.Make
			var err error
			if genFile != "" {
				gen, err = generator.NewMake(genFile, workingDir)
				if err != nil {
					return errors.Wrapf(err, "creating generator from build file '%s'", genFile)
				}
			} else if ctrlFile != "" {
				mc, err := generator.NewMakeControl(ctrlFile)
				if err != nil {
					return errors.Wrapf(err, "building from control file '%s'", ctrlFile)
				}
				gen, err = mc.Build()
				if err != nil {
					return errors.Wrapf(err, "creating generator from control file '%s'", ctrlFile)
				}
			}
			conf, err := gen.Generate()
			if err != nil {
				return errors.Wrapf(err, "generating evergreen config from build file(s)")
			}

			output, err := json.MarshalIndent(conf, "", "\t")
			if err != nil {
				return errors.Wrap(err, "marshalling evergreen config as JSON")
			}
			if outputFile != "" {
				if err := ioutil.WriteFile(outputFile, output, 0644); err != nil {
					return errors.Wrapf(err, "writing JSON config to file '%s'", outputFile)
				}
			} else {
				fmt.Println(string(output))
			}

			return nil
		},
	}
}
