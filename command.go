package jasper

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/google/shlex"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/level"
	"github.com/mongodb/grip/message"
	"github.com/mongodb/grip/recovery"
	"github.com/mongodb/grip/send"
	"github.com/pkg/errors"
)

// Command TODO.
type Command struct {
	cmds     [][]string
	opts     CreateOptions
	priority level.Priority
	id       string

	continueOnError bool
	ignoreError     bool
	prerequisite    func() bool
}

func getRemoteCreateOpt(ctx context.Context, host string, args []string, dir string) (*CreateOptions, error) {
	var remoteCmd string

	if dir != "" {
		remoteCmd = fmt.Sprintf("cd %s && ", dir)
	}

	switch len(args) {
	case 0:
		return nil, errors.New("args invalid")
	case 1:
		remoteCmd += args[0]
	default:
		remoteCmd += strings.Join(args, " ")
	}

	return &CreateOptions{Args: []string{"ssh", host, remoteCmd}}, nil
}

func getLogOutput(out []byte) string {
	return strings.Trim(strings.Replace(string(out), "\n", "\n\t out -> ", -1), "\n\t out->")
}

func splitCmdToArgs(cmd string) []string {
	args, err := shlex.Split(cmd)
	if err != nil {
		grip.Error(message.WrapError(err, message.Fields{"input": cmd}))
		return nil
	}
	return args
}

// NewCommand TODO.
func NewCommand() *Command { return &Command{} }

// String TODO.
func (c *Command) String() string { return fmt.Sprintf("id='%s', cmd='%s'", c.id, c.getCmd()) }

// Add TODO.
func (c *Command) Add(args []string) *Command { c.cmds = append(c.cmds, args); return c }

// Extend TODO.
func (c *Command) Extend(cmds [][]string) *Command { c.cmds = append(c.cmds, cmds...); return c }

// Directory TODO.
func (c *Command) Directory(d string) *Command { c.opts.WorkingDirectory = d; return c }

// Host TODO.
func (c *Command) Host(h string) *Command { c.opts.Hostname = h; return c }

// Priority TODO.
func (c *Command) Priority(l level.Priority) *Command { c.priority = l; return c }

// ID TODO.
func (c *Command) ID(id string) *Command { c.id = id; return c }

// ContinueOnError TODO.
func (c *Command) ContinueOnError(cont bool) *Command { c.continueOnError = cont; return c }

// IgnoreError TODO.
func (c *Command) IgnoreError(ignore bool) *Command { c.ignoreError = ignore; return c }

// Environment TODO.
func (c *Command) Environment(e map[string]string) *Command { c.opts.Environment = e; return c }

// AddEnv TODO.
func (c *Command) AddEnv(k, v string) *Command { c.setupEnv(); c.opts.Environment[k] = v; return c }

// Prerequisite TODO.
func (c *Command) Prerequisite(chk func() bool) *Command { c.prerequisite = chk; return c }

// Append TODO.
func (c *Command) Append(cmds ...string) *Command {
	for _, cmd := range cmds {
		c.cmds = append(c.cmds, splitCmdToArgs(cmd))
	}
	return c
}

func (c *Command) setupEnv() {
	if c.opts.Environment == nil {
		c.opts.Environment = map[string]string{}
	}
}

// Run TODO.
func (c *Command) Run(ctx context.Context) error {
	if c.prerequisite != nil && !c.prerequisite() {
		grip.Debug(message.Fields{
			"op":  "noop after prerequisite returned false",
			"id":  c.id,
			"cmd": c.String(),
		})
		return nil
	}

	c.finalizeWriters()
	catcher := grip.NewBasicCatcher()

	var opts []*CreateOptions
	opts, err := c.getCreateOpts(ctx)
	if err != nil {
		catcher.Add(err)
		catcher.Add(c.Close())
		return catcher.Resolve()
	}

	for idx, opt := range opts {
		if err := ctx.Err(); err != nil {
			catcher.Add(errors.Wrap(err, "operation canceled"))
			catcher.Add(c.Close())
			return catcher.Resolve()
		}

		err := c.exec(ctx, opt, idx)
		if !c.ignoreError {
			catcher.Add(err)
		}

		if err != nil && !c.continueOnError {
			catcher.Add(c.Close())
			return catcher.Resolve()
		}
	}

	catcher.Add(c.Close())
	return catcher.Resolve()
}

// RunParallel TODO.
func (c *Command) RunParallel(ctx context.Context) error {
	// Avoid paying the copy-costs in between command structs by doing the work
	// before executing the commands.
	parallelCmds := make([]Command, len(c.cmds))

	for idx, cmd := range c.cmds {
		splitCmd := *c
		splitCmd.opts.closers = []func() error{}
		splitCmd.cmds = [][]string{cmd}
		parallelCmds[idx] = splitCmd
	}

	catcher := grip.NewBasicCatcher()
	errs := make(chan error, len(c.cmds))
	for _, parallelCmd := range parallelCmds {
		go func(innerCmd Command) {
			defer func() {
				errs <- recovery.HandlePanicWithError(recover(), nil, "parallel command encountered error")
			}()
			errs <- innerCmd.Run(ctx)
		}(parallelCmd)
	}

	for i := 0; i < len(c.cmds); i++ {
		select {
		case err := <-errs:
			if !c.ignoreError {
				catcher.Add(err)
			}
		case <-ctx.Done():
			catcher.Add(c.Close())
			catcherErr := catcher.Resolve()
			if catcherErr != nil {
				return errors.Wrapf(ctx.Err(), catcherErr.Error())
			}
			return ctx.Err()
		}
	}

	catcher.Add(c.Close())
	return catcher.Resolve()
}

// Close TODO.
func (c *Command) Close() error {
	catcher := grip.NewBasicCatcher()
	for _, closer := range c.opts.closers {
		catcher.Add(closer())
	}

	return catcher.Resolve()
}

// SetErrorSender TODO.
func (c *Command) SetErrorSender(l level.Priority, s send.Sender) *Command {
	writer := send.MakeWriterSender(s, l)
	c.opts.closers = append(c.opts.closers, writer.Close)
	c.opts.Output.Error = writer
	return c
}

// SetOutputSender TODO.
func (c *Command) SetOutputSender(l level.Priority, s send.Sender) *Command {
	writer := send.MakeWriterSender(s, l)
	c.opts.closers = append(c.opts.closers, writer.Close)
	c.opts.Output.Output = writer
	return c
}

// SetCombinedSender TODO.
func (c *Command) SetCombinedSender(l level.Priority, s send.Sender) *Command {
	writer := send.MakeWriterSender(s, l)
	c.opts.closers = append(c.opts.closers, writer.Close)
	c.opts.Output.Error = writer
	c.opts.Output.Output = writer
	return c
}

// SetErrorWriter TODO.
func (c *Command) SetErrorWriter(writer io.WriteCloser) *Command {
	c.opts.closers = append(c.opts.closers, writer.Close)
	c.opts.Output.Error = writer
	return c
}

// SetOutputWriter TODO.
func (c *Command) SetOutputWriter(writer io.WriteCloser) *Command {
	c.opts.closers = append(c.opts.closers, writer.Close)
	c.opts.Output.Output = writer
	return c
}

// SetCombinedWriter TODO.
func (c *Command) SetCombinedWriter(writer io.WriteCloser) *Command {
	c.opts.closers = append(c.opts.closers, writer.Close)
	c.opts.Output.Error = writer
	c.opts.Output.Output = writer
	return c
}

func (c *Command) finalizeWriters() {
	if c.opts.Output.Output == nil {
		c.opts.Output.Output = ioutil.Discard
	}

	if c.opts.Output.Error == nil {
		c.opts.Output.Error = ioutil.Discard
	}
}

func (c *Command) getEnv() []string {
	out := []string{}
	for k, v := range c.opts.Environment {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return out
}

func (c *Command) getCmd() string {
	env := strings.Join(c.getEnv(), " ")
	out := []string{}
	for _, cmd := range c.cmds {
		out = append(out, fmt.Sprintf("%s '%s';\n", env, strings.Join(cmd, " ")))
	}
	return strings.Join(out, "")
}

func getCreateOpt(ctx context.Context, args []string, dir string, env map[string]string) (*CreateOptions, error) {
	var opts *CreateOptions
	switch len(args) {
	case 0:
		return nil, errors.New("args invalid")
	case 1:
		if strings.Contains(args[0], " \"'") {
			spl, err := shlex.Split(args[0])
			if err != nil {
				return nil, errors.Wrap(err, "problem splitting argstring")
			}
			return getCreateOpt(ctx, spl, dir, env)
		}
		opts = &CreateOptions{Args: args}
	default:
		opts = &CreateOptions{Args: args}
	}
	opts.WorkingDirectory = dir

	for k, v := range env {
		opts.Environment[k] = v
	}

	return opts, nil
}

func (c *Command) getCreateOpts(ctx context.Context) ([]*CreateOptions, error) {
	out := []*CreateOptions{}
	catcher := grip.NewBasicCatcher()
	if c.opts.Hostname != "" {
		for _, args := range c.cmds {
			cmd, err := getRemoteCreateOpt(ctx, c.opts.Hostname, args, c.opts.WorkingDirectory)
			if err != nil {
				catcher.Add(err)
				continue
			}

			out = append(out, cmd)
		}
	} else {
		for _, args := range c.cmds {
			cmd, err := getCreateOpt(ctx, args, c.opts.WorkingDirectory, c.opts.Environment)
			if err != nil {
				catcher.Add(err)
				continue
			}

			out = append(out, cmd)
		}
	}
	if catcher.HasErrors() {
		return nil, catcher.Resolve()
	}

	return out, nil
}

func (c *Command) exec(ctx context.Context, opts *CreateOptions, idx int) error {
	msg := message.Fields{
		"id":  c.id,
		"cmd": strings.Join(opts.Args, " "),
		"idx": idx,
		"len": len(c.cmds),
	}

	var err error
	var newProc Process
	if c.opts.Output.Output == nil {
		var out bytes.Buffer
		opts.Output.Output = &out
		opts.Output.Error = &out
		newProc, err = newBasicProcess(ctx, opts)
		if err != nil {
			return errors.Wrapf(err, "problem starting command")
		}

		_, err = newProc.Wait(ctx)
		msg["out"] = getLogOutput(out.Bytes())
		msg["err"] = err
	} else {
		opts.Output.Error = c.opts.Output.Error
		opts.Output.Output = c.opts.Output.Output
		newProc, err = newBasicProcess(ctx, opts)
		if err != nil {
			return errors.Wrapf(err, "problem starting command")
		}

		_, err = newProc.Wait(ctx)
		msg["err"] = err
	}
	grip.Log(c.priority, msg)
	return err
}

// RunCommand TODO.
func RunCommand(ctx context.Context, id string, pri level.Priority, args []string, dir string, env map[string]string) error {
	return NewCommand().
		ID(id).
		Priority(pri).
		Add(args).
		Directory(dir).
		Environment(env).
		Run(ctx)
}

// RunRemoteCommand TODO.
func RunRemoteCommand(ctx context.Context, id string, pri level.Priority, host string, args []string, dir string) error {
	return NewCommand().
		ID(id).
		Priority(pri).
		Host(host).
		Add(args).
		Directory(dir).
		Run(ctx)
}

// RunCommandGroupContinueOnError TODO.
func RunCommandGroupContinueOnError(ctx context.Context, id string, pri level.Priority, cmds [][]string, dir string, env map[string]string) error {
	return NewCommand().
		ID(id).
		Priority(pri).
		Extend(cmds).
		Directory(dir).
		Environment(env).
		ContinueOnError(true).
		Run(ctx)
}

// RunRemoteCommandGroupContinueOnError TODO.
func RunRemoteCommandGroupContinueOnError(ctx context.Context, id string, pri level.Priority, host string, cmds [][]string, dir string) error {
	return NewCommand().
		ID(id).
		Priority(pri).
		Host(host).
		Extend(cmds).
		Directory(dir).
		ContinueOnError(true).
		Run(ctx)
}

// RunCommandGroup TODO.
func RunCommandGroup(ctx context.Context, id string, pri level.Priority, cmds [][]string, dir string, env map[string]string) error {
	return NewCommand().
		ID(id).
		Priority(pri).
		Extend(cmds).
		Directory(dir).
		Environment(env).
		Run(ctx)
}

// RunRemoteCommandGroup TODO.
func RunRemoteCommandGroup(ctx context.Context, id string, pri level.Priority, host string, cmds [][]string, dir string) error {
	return NewCommand().
		ID(id).
		Priority(pri).
		Host(host).
		Extend(cmds).
		Directory(dir).
		Run(ctx)
}

// RunParallelCommandGroup TODO.
func RunParallelCommandGroup(ctx context.Context, id string, pri level.Priority, cmds [][]string, dir string, env map[string]string) error {
	return NewCommand().
		ID(id).
		Priority(pri).
		Extend(cmds).
		Directory(dir).
		Environment(env).
		RunParallel(ctx)
}

// RunParallelRemoteCommandGroup TODO.
func RunParallelRemoteCommandGroup(ctx context.Context, id string, pri level.Priority, host string, cmds [][]string, dir string) error {
	return NewCommand().
		ID(id).
		Priority(pri).
		Host(host).
		Extend(cmds).
		Directory(dir).
		RunParallel(ctx)
}

// RunParallelCommandGroupContinueOnError TODO.
func RunParallelCommandGroupContinueOnError(ctx context.Context, id string, pri level.Priority, cmds [][]string, dir string, env map[string]string) error {
	return NewCommand().
		ID(id).
		Priority(pri).
		Extend(cmds).
		Directory(dir).
		Environment(env).
		ContinueOnError(true).
		RunParallel(ctx)
}

// RunParallelRemoteCommandGroupContinueOnError TODO.
func RunParallelRemoteCommandGroupContinueOnError(ctx context.Context, id string, pri level.Priority, host string, cmds [][]string, dir string) error {
	return NewCommand().
		ID(id).
		Priority(pri).
		Host(host).
		Extend(cmds).
		Directory(dir).
		ContinueOnError(true).
		RunParallel(ctx)
}
