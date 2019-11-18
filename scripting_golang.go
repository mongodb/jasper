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

type golangEnvironment struct {
	opts *options.ScriptingGolang

	isConfigured bool
	cachedHash   string
	manager      Manager
}

func (e *golangEnvironment) ID() string { e.cachedHash = e.opts.ID(); return e.cachedHash }
func (e *golangEnvironment) Setup(ctx context.Context) error {
	if e.isConfigured && e.cachedHash == e.opts.ID() {
		return nil
	}

	if e.opts.Gopath == "" {
		e.opts.Gopath = filepath.Join("go", uuid.Must(uuid.NewV4()).String())
	}
	e.cachedHash = e.opts.ID()

	gobin := e.opts.Interpreter()
	cmd := e.manager.CreateCommand(ctx)

	for _, pkg := range e.opts.Packages {
		if e.opts.WithUpdate {
			cmd.AppendArgs(gobin, "get", "-u", pkg)
		} else {
			cmd.AppendArgs(gobin, "get", pkg)
		}
	}

	cmd.SetHook(func(res error) error {
		if res == nil {
			e.isConfigured = true
		}
		return nil
	})

	return cmd.Run(ctx)
}

func (e *golangEnvironment) Run(ctx context.Context, args []string) error {
	cmd := e.manager.CreateCommand(ctx).
		AddEnv("GOPATH", e.opts.Gopath).
		AddEnv("GOROOT", e.opts.Goroot).
		Add(append([]string{e.opts.Interpreter(), "run"}, args...))

	if e.opts.Context != "" {
		cmd.Directory(e.opts.Context)
	}

	return cmd.Run(ctx)
}

func (e *golangEnvironment) Build(ctx context.Context, dir string, args []string) error {
	return e.manager.CreateCommand(ctx).
		Directory(dir).
		AddEnv("GOPATH", e.opts.Gopath).
		AddEnv("GOROOT", e.opts.Goroot).
		Add(append([]string{e.opts.Interpreter(), "build"}, args...)).Run(ctx)

}

func (e *golangEnvironment) RunScript(ctx context.Context, script string) error {
	scriptChecksum := fmt.Sprintf("%x", sha1.Sum([]byte(script)))
	path := strings.Join([]string{e.manager.ID(), scriptChecksum}, "_") + ".go"
	if e.opts.Context != "" {
		path = filepath.Join(e.opts.Context, path)
	} else {
		path = filepath.Join(e.opts.Gopath, "tmp", path)

	}

	wo := options.WriteFile{
		Path:    path,
		Content: []byte(script),
	}

	if err := e.manager.WriteFile(ctx, wo); err != nil {
		return errors.Wrap(err, "problem writing file")
	}

	return e.manager.CreateCommand(ctx).AppendArgs(e.opts.Interpreter(), wo.Path).Run(ctx)
}

func (e *golangEnvironment) Cleanup(ctx context.Context) error {
	switch mgr := e.manager.(type) {
	case RemoteClient:
		return errors.Wrapf(mgr.CreateCommand(ctx).AppendArgs("rm", "-rf", e.opts.Gopath).Run(ctx),
			"problem removing remote golang environment '%s'", e.opts.Gopath)
	default:
		return errors.Wrapf(os.RemoveAll(e.opts.Gopath),
			"problem removing local golang environment '%s'", e.opts.Gopath)
	}
}
