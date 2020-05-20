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
	workingDirFlagName = "working_dir"
	filesFlagName      = "files"
	outputFileFlagName = "output_file"
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
			cli.StringSliceFlag{
				Name:  filesFlagName,
				Usage: "The build files necessary to generate the evergreen config.",
			},
			cli.StringFlag{
				Name:  outputFileFlagName,
				Usage: "The output file. If unspecified, the config will be written to stdout.",
			},
		},
		Before: mergeBeforeFuncs(
			requireStringFlag(workingDirFlagName),
			requireStringSliceFlag(filesFlagName),
		),
		Action: func(c *cli.Context) error {
			workingDir := c.String(workingDirFlagName)
			files := c.StringSlice(filesFlagName)
			outputFile := c.String(outputFileFlagName)
			var err error
			if !filepath.IsAbs(workingDir) {
				workingDir, err = filepath.Abs(workingDir)
				if err != nil {
					return errors.Wrapf(err, "getting working directory '%s' as absolute path", workingDir)
				}
			}

			for _, file := range files {
				gen, err := generator.NewGolang(file, workingDir)
				if err != nil {
					return errors.Wrapf(err, "creating generator from build file '%s'", file)
				}
				conf, err := gen.Generate()
				if err != nil {
					return errors.Wrapf(err, "generating evergreen config from build file '%s'", file)
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
				Usage: "The directory where the project will be cloned.",
			},
			cli.StringSliceFlag{
				Name:  filesFlagName,
				Usage: "The build files necessary to generate the evergreen config.",
			},
			cli.StringFlag{
				Name:  outputFileFlagName,
				Usage: "The output file. If unspecified, the config will be written to stdout.",
			},
		},
		Before: mergeBeforeFuncs(
			requireStringFlag(workingDirFlagName),
			requireStringSliceFlag(filesFlagName),
		),
		Action: func(c *cli.Context) error {
			workingDir := c.String(workingDirFlagName)
			files := c.StringSlice(filesFlagName)
			outputFile := c.String(outputFileFlagName)
			var err error
			if !filepath.IsAbs(workingDir) {
				workingDir, err = filepath.Abs(workingDir)
				if err != nil {
					return errors.Wrapf(err, "getting working directory '%s' as absolute path", workingDir)
				}
			}

			for _, file := range files {
				gen, err := generator.NewMake(file, workingDir)
				if err != nil {
					return errors.Wrapf(err, "creating generator from build file '%s'", file)
				}
				conf, err := gen.Generate()
				if err != nil {
					return errors.Wrapf(err, "generating evergreen config from build file '%s'", file)
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
			}

			return nil
		},
	}
}
