//go:build darwin || linux

package jasper

import (
	"bufio"
	"context"
	"os/exec"

	"github.com/pkg/errors"
)

func isSudo(ctx context.Context) (bool, error) {
	if err := exec.CommandContext(ctx, "sudo", "-n", "date").Run(); err != nil {
		switch err.(type) {
		case *exec.ExitError:
			return false, nil
		default:
			return false, errors.Wrap(err, "executing sudo permissions check")
		}
	}

	return true, nil
}

type logAnalyzer struct {
	cmdArgs        []string
	lineHasOOMKill func(string) bool
	extractPID     func(string) (int, bool)
}

func (a *logAnalyzer) analyzeKernelLog(ctx context.Context) ([]string, []int, error) {
	sudo, err := isSudo(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "checking sudo")
	}

	var cmd *exec.Cmd
	if sudo {
		cmd = exec.CommandContext(ctx, "sudo", a.cmdArgs...)
	} else {
		cmd = exec.CommandContext(ctx, a.cmdArgs[0], a.cmdArgs[1:]...)
	}
	logPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, errors.Wrap(err, "creating standard output pipe for log command")
	}
	scanner := bufio.NewScanner(logPipe)
	if err := cmd.Start(); err != nil {
		return nil, nil, errors.Wrap(err, "starting log command")
	}

	lines := []string{}
	pids := []int{}
	for scanner.Scan() {
		line := scanner.Text()
		if a.lineHasOOMKill(line) {
			lines = append(lines, line)
			if pid, hasPID := a.extractPID(line); hasPID {
				pids = append(pids, pid)
			}
		}
	}

	errs := make(chan error, 1)
	select {
	case <-ctx.Done():
		return nil, nil, errors.New("request cancelled")
	case errs <- cmd.Wait():
		err = <-errs
		return lines, pids, errors.Wrap(err, "waiting for log command")
	}
}
