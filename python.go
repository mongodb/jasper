package jasper

import (
	"context"
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type pythonEnvironment struct {
	opts *options.ScriptingPython

	isConfigured bool
	cachedHash   string
	manager      Manager
}

func (e *pythonEnvironment) ID() string { e.cachedHash = e.opts.ID(); return e.cachedHash }

func (e *pythonEnvironment) Setup(ctx context.Context) error {
	if e.isConfigured && e.cachedHash == e.opts.ID() {
		return nil
	}

	if e.opts.VirtualEnvPath == "" {
		e.opts.VirtualEnvPath = filepath.Join(uuid.Must(uuid.NewV4()).String())
	}
	if e.opts.HostPythonInterpeter == "" {
		e.opts.HostPythonInterpeter = "python3"
	}

	cmd := e.manager.CreateCommand(ctx)
	venvpy := e.opts.Interpreter()

	if _, err := os.Stat(venvpy); os.IsNotExist(err) {
		cmd = cmd.AppendArgs(e.opts.HostPythonInterpeter, "-m", e.venvMod(), e.opts.VirtualEnvPath)
	}

	if e.opts.RequirementsFilePath != "" {
		cmd.AppendArgs(venvpy, "-m", "pip", "install", "-r", e.opts.RequirementsFilePath)
	}

	for _, pkg := range e.opts.Packages {
		cmd.AppendArgs(venvpy, "-m", "pip", "install", "-r", pkg)
	}

	cmd.SetHook(func(res error) error {
		if res == nil {
			e.isConfigured = true
		}
		return nil
	})

	return cmd.Run(ctx)
}

func (e *pythonEnvironment) venvMod() string {
	if e.opts.LegacyPython {
		return "virtualenv"
	}
	return "venv"
}

func (e *pythonEnvironment) Run(ctx context.Context, args []string) error {
	return e.manager.CreateCommand(ctx).Add(append([]string{e.opts.Interpreter()}, args...)).Run(ctx)
}

func (e *pythonEnvironment) RunScript(ctx context.Context, script string) error {
	scriptChecksum := fmt.Sprintf("%x", sha1.Sum([]byte(script)))
	wo := options.WriteFile{
		Path:    filepath.Join(e.opts.VirtualEnvPath, "tmp", strings.Join([]string{e.manager.ID(), scriptChecksum}, "_")+".py"),
		Content: []byte(script),
	}

	if err := e.manager.WriteFile(ctx, wo); err != nil {
		return errors.Wrap(err, "problem writing file")
	}

	return e.manager.CreateCommand(ctx).AppendArgs(e.opts.Interpreter(), wo.Path).Run(ctx)
}

func (e *pythonEnvironment) Cleanup(ctx context.Context) error {
	switch mgr := e.manager.(type) {
	case RemoteClient:
		return errors.Wrapf(mgr.CreateCommand(ctx).Add("rm", "-rf", e.opts.VirtualEnvPath).Run(ctx),
			"problem removing remote python environment '%s'", e.opts.VirtualEnvPath)
	default:
		return errors.Wrapf(os.RemoveAll(e.opts.VirtualEnvPath),
			"problem removing local python environment '%s'", e.opts.VirtualEnvPath)
	}
}
