// +build linux

package jasper

import (
	"context"
	"os"
	"runtime"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinuxProcessTracker(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("cannot run Linux process tracker tests without admin privileges")
	}
	for procName, makeProc := range map[string]ProcessConstructor{
		"Blocking": newBlockingProcess,
		"Basic":    newBasicProcess,
	} {
		t.Run(procName, func(t *testing.T) {

			for name, testCase := range map[string]func(context.Context, *testing.T, *linuxProcessTracker, Process){
				"ValidateInitialSetup": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					require.NotNil(t, tracker.cgroup)
					require.True(t, tracker.valid())
					pids, err := tracker.listPIDs()
					require.NoError(t, err)
					require.Len(t, pids, 0)
				},
				"NilCgroupIsInvalid": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					tracker.cgroup = nil
					assert.False(t, tracker.valid())
				},
				"DeletedCgroupIsInvalid": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					require.NoError(t, tracker.cgroup.Delete())
					assert.False(t, tracker.valid())
				},
				"SetDefaultCgroupIfInvalidNoopsIfCgroupIsValid": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					cgroup := tracker.cgroup
					assert.NotNil(t, cgroup)
					assert.NoError(t, tracker.setDefaultCgroupIfInvalid())
					assert.Equal(t, cgroup, tracker.cgroup)
				},
				"SetDefaultCgroupIfNilSetsIfCgroupIsInvalid": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					tracker.cgroup = nil
					assert.NoError(t, tracker.setDefaultCgroupIfInvalid())
					assert.NotNil(t, tracker.cgroup)
				},
				"AddNewProcessSucceeds": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					pid := proc.Info(ctx).PID
					assert.NoError(t, tracker.add(pid))

					pids, err := tracker.listPIDs()
					require.NoError(t, err)
					require.Len(t, pids, 1)
					assert.Equal(t, pid, pids[0])
				},
				"DoubleAddProcessSucceedsButDoesNotDuplicateProcess": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					pid := proc.Info(ctx).PID
					assert.NoError(t, tracker.add(pid))
					assert.NoError(t, tracker.add(pid))

					pids, err := tracker.listPIDs()
					require.NoError(t, err)
					require.Len(t, pids, 1)
					assert.Equal(t, pid, pids[0])
				},
				"ListPIDsDoesNotSeeTerminatedProcesses": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					pid := proc.Info(ctx).PID
					require.NoError(t, tracker.add(pid))

					assert.NoError(t, proc.RegisterSignalTriggerID(ctx, CleanTerminationSignalTrigger))
					err := proc.Signal(ctx, syscall.SIGTERM)
					assert.NoError(t, err)
					exitCode, err := proc.Wait(ctx)
					assert.Error(t, err)
					if runtime.GOOS == "windows" {
						assert.Zero(t, exitCode)
					} else {
						assert.Equal(t, exitCode, int(syscall.SIGTERM))
					}

					pids, err := tracker.listPIDs()
					require.NoError(t, err)
					assert.Len(t, pids, 0)
				},
				"ListPIDsErrorsIfCgroupDeleted": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					assert.NoError(t, tracker.cgroup.Delete())
					pids, err := tracker.listPIDs()
					assert.Error(t, err)
					assert.Len(t, pids, 0)
				},
				"CleanupNoProcsSucceeds": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					pids, err := tracker.listPIDs()
					require.NoError(t, err)
					assert.Len(t, pids, 0)
					assert.NoError(t, tracker.cleanup())
				},
				"CleanupTerminatesProcessInCgroup": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					pid := proc.Info(ctx).PID
					assert.NoError(t, tracker.add(pid))
					assert.NoError(t, tracker.cleanup())

					procTerminated := make(chan struct{})
					go func() {
						defer close(procTerminated)
						_, _ = proc.Wait(ctx)
					}()

					select {
					case <-procTerminated:
					case <-ctx.Done():
						assert.Fail(t, "context timed out before process was complete")
					}
				},
				"DoubleCleanupDoesNotError": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					pid := proc.Info(ctx).PID
					assert.NoError(t, tracker.add(pid))
					assert.NoError(t, tracker.cleanup())
					assert.NoError(t, tracker.cleanup())
				},
				"AddProcessAfterCleanupSucceeds": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					pid := proc.Info(ctx).PID
					require.NoError(t, tracker.add(pid))
					require.NoError(t, tracker.cleanup())

					opts := yesCreateOpts(0)
					newProc, err := makeProc(ctx, &opts)
					require.NoError(t, err)
					newPID := newProc.Info(ctx).PID

					require.NoError(t, tracker.add(newPID))
					pids, err := tracker.listPIDs()
					require.NoError(t, err)
					assert.Len(t, pids, 1)
				},
				"UpdateLimitChecksForLinuxResources": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					assert.Error(t, tracker.setLimits("foobar"))
					assert.NoError(t, tracker.setLimits(&LinuxResources{}))
				},
				"UpdateLimitKillsProcessWithZeroMemoryLimit": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					pid := proc.Info(ctx).PID

					zero := int64(0)
					require.NoError(t, tracker.setLimits(&LinuxResources{
						Memory: &LinuxMemory{Limit: &zero},
					}))

					require.NoError(t, tracker.add(pid))

					procTerminated := make(chan struct{})
					go func() {
						defer close(procTerminated)
						_, _ = proc.Wait(ctx)
					}()

					select {
					case <-procTerminated:
					case <-ctx.Done():
						require.Fail(t, "context timed out before process was killed")
					}

					oomTracker := NewOOMTracker()
					assert.NoError(t, oomTracker.Check(ctx))
					oomKilled, pids := oomTracker.Report()
					assert.True(t, oomKilled)
					assert.NotZero(t, len(pids))

					oomKilledProc := false
					for _, oomKilledPID := range pids {
						if oomKilledPID == pid {
							oomKilledProc = true
							break
						}
					}
					assert.True(t, oomKilledProc)
				},
			} {
				t.Run(name, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
					defer cancel()

					opts := yesCreateOpts(0)
					proc, err := makeProc(ctx, &opts)
					require.NoError(t, err)

					tracker, err := newProcessTracker("test")
					require.NoError(t, err)
					require.NotNil(t, tracker)
					linuxTracker, ok := tracker.(*linuxProcessTracker)
					require.True(t, ok)
					defer func() {
						// Ensure that the cgroup is cleaned up.
						assert.NoError(t, tracker.cleanup())
					}()

					testCase(ctx, t, linuxTracker, proc)
				})
			}
		})
	}
}
