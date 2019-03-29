// +build linux

package jasper

import (
	"context"
	"os/exec"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetActivePIDs(t *testing.T) {
	for procName, makeProc := range map[string]ProcessConstructor{
		"Basic":    newBasicProcess,
		"Blocking": newBlockingProcess,
	} {
		t.Run(procName, func(t *testing.T) {
			for testName, testCase := range map[string]func(context.Context, *CreateOptions){
				"GetActivePIDsListsProcess": func(ctx context.Context, opts *CreateOptions) {
					proc, err := makeProc(ctx, opts)
					require.NoError(t, err)
					pid := proc.Info(ctx).PID

					pids, err := getActivePIDs()
					require.NoError(t, err)
					assert.Contains(t, pids, pid)
				},
				"GetActivePIDsDoesNotListFinishedProcesses": func(ctx context.Context, opts *CreateOptions) {
					opts.Args = []string{"ls"}
					proc, err := makeProc(ctx, opts)
					require.NoError(t, err)
					pid := proc.Info(ctx).PID
					exitCode, err := proc.Wait(ctx)
					require.NoError(t, err)
					assert.Zero(t, exitCode)

					pids, err := getActivePIDs()
					require.NoError(t, err)
					assert.NotZero(t, len(pids))
					assert.NotContains(t, pids, pid)
				},
			} {
				t.Run(testName, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
					defer cancel()
					opts := yesCreateOpts(taskTimeout)
					testCase(ctx, &opts)
				})
			}
		})
	}
}

func TestGetEnvironmentVariables(t *testing.T) {
	for procName, makeProc := range map[string]ProcessConstructor{
		"Basic":    newBasicProcess,
		"Blocking": newBlockingProcess,
	} {
		t.Run(procName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
			defer cancel()

			envVar := "foo"
			value := "bar"

			opts := yesCreateOpts(taskTimeout)
			opts.AddEnvVar("foo", "bar")

			proc, err := makeProc(ctx, &opts)
			require.NoError(t, err)
			pid := proc.Info(ctx).PID

			// Wait for the process to set up the environment variable and check that it is correct.
			for {
				select {
				case <-ctx.Done():
					assert.Fail(t, "context timed out before environment variables were set for process")
				default:
					if env, err := getEnvironmentVariables(pid); err == nil {
						if actualValue, ok := env[envVar]; ok {
							assert.Equal(t, value, actualValue)
							break
						}
					}
				}
			}
		})
	}
}

func TestCleanupProcess(t *testing.T) {
	for testName, testCase := range map[string]func(*exec.Cmd){
		"CleanupProcessSucceedsForRunningProcess": func(cmd *exec.Cmd) {
			require.NoError(t, cmd.Start())
			proc := cmd.Process
			require.True(t, proc.Pid > 0)
			assert.NoError(t, cleanupProcess(proc))

			state, err := proc.Wait()
			require.NoError(t, err)

			waitStatus, ok := state.Sys().(syscall.WaitStatus)
			require.True(t, ok)
			require.True(t, waitStatus.Signaled())
			assert.Equal(t, waitStatus.Signal(), syscall.SIGTERM)
		},
		"CleanupProcessFailsForFinishedProcess": func(cmd *exec.Cmd) {
			require.NoError(t, cmd.Start())
			proc := cmd.Process

			require.True(t, proc.Pid > 0)

			_, err := proc.Wait()
			require.NoError(t, err)

			assert.Error(t, cleanupProcess(proc))
		},
	} {
		t.Run(testName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
			defer cancel()

			cmd := exec.CommandContext(ctx, "yes")
			testCase(cmd)
		})
	}
}
