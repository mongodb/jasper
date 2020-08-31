package cli

import "github.com/urfave/cli"

// Constants representing the scripting.Harness interface as CLI commands.
const (
	ScriptingCommand          = "scripting"
	ScriptingSetupCommand     = "setup"
	ScriptingRunCommand       = "run"
	ScriptingRunScriptCommand = "run-script"
	ScriptingBuildCommand     = "build"
	ScriptingTestCommand      = "test"
	ScriptingCleanupCommand   = "cleanup"
)

func Scripting() cli.Command {
	return cli.Command{
		Name: ScriptingCommand,
		Subcommands: []cli.Command{
			scriptingSetup(),
		},
	}
}

// kim: TODO: implement scripting.Harness interface methods.
func scriptingSetup() cli.Command {
	return cli.Command{
		Name: ScriptingSetupCommand,
	}
}

func scriptingRun() cli.Command {
	return cli.Command{
		Name: ScriptingRunCommand,
	}
}

func scriptingRunScript() cli.Command {
	return cli.Command{
		Name: ScriptingRunScriptCommand,
	}
}

func scriptingBuild() cli.Command {
	return cli.Command{
		Name: ScriptingBuildCommand,
	}
}

func scriptingTest() cli.Command {
	return cli.Command{
		Name: ScriptingTestCommand,
	}
}

func scriptingCleanup() cli.Command {
	return cli.Command{
		Name: ScriptingCleanupCommand,
	}
}
