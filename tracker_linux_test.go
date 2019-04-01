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

func TestLinuxProcessTrackerWithCgroups(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("cannot run Linux process tracker tests with cgroups without admin privileges")
	}
	for procName, makeProc := range map[string]ProcessConstructor{
		"Blocking": newBlockingProcess,
		"Basic":    newBasicProcess,
	} {
		t.Run(procName, func(t *testing.T) {

			for name, testCase := range map[string]func(context.Context, *testing.T, *linuxProcessTracker, Process){
				"ValidateInitialSetup": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					require.NotNil(t, tracker.cgroup)
					require.True(t, tracker.validCgroup())
					pids, err := tracker.listCgroupPIDs()
					require.NoError(t, err)
					require.Len(t, pids, 0)
				},
				"NilCgroupIsInvalid": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					tracker.cgroup = nil
					assert.False(t, tracker.validCgroup())
				},
				"DeletedCgroupIsInvalid": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					require.NoError(t, tracker.cgroup.Delete())
					assert.False(t, tracker.validCgroup())
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
					assert.NoError(t, tracker.Add(pid))

					pids, err := tracker.listCgroupPIDs()
					require.NoError(t, err)
					require.Len(t, pids, 1)
					assert.Equal(t, pid, pids[0])
				},
				"DoubleAddProcessSucceedsButDoesNotDuplicateProcess": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					pid := proc.Info(ctx).PID
					assert.NoError(t, tracker.Add(pid))
					assert.NoError(t, tracker.Add(pid))

					pids, err := tracker.listCgroupPIDs()
					require.NoError(t, err)
					require.Len(t, pids, 1)
					assert.Equal(t, pid, pids[0])
				},
				"ListPIDsDoesNotSeeTerminatedProcesses": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					pid := proc.Info(ctx).PID
					require.NoError(t, tracker.Add(pid))

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

					pids, err := tracker.listCgroupPIDs()
					require.NoError(t, err)
					assert.Len(t, pids, 0)
				},
				"ListPIDsErrorsIfCgroupDeleted": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					assert.NoError(t, tracker.cgroup.Delete())
					pids, err := tracker.listCgroupPIDs()
					assert.Error(t, err)
					assert.Len(t, pids, 0)
				},
				"CleanupNoProcsSucceeds": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					pids, err := tracker.listCgroupPIDs()
					require.NoError(t, err)
					assert.Len(t, pids, 0)
					assert.NoError(t, tracker.Cleanup())
				},
				"CleanupTerminatesProcessInCgroup": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					pid := proc.Info(ctx).PID
					assert.NoError(t, tracker.Add(pid))
					assert.NoError(t, tracker.Cleanup())

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
					assert.NoError(t, tracker.Add(pid))
					assert.NoError(t, tracker.Cleanup())
					assert.NoError(t, tracker.Cleanup())
				},
				"AddProcessAfterCleanupSucceeds": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					pid := proc.Info(ctx).PID
					require.NoError(t, tracker.Add(pid))
					require.NoError(t, tracker.Cleanup())

					opts := yesCreateOpts(0)
					newProc, err := makeProc(ctx, &opts)
					require.NoError(t, err)
					newPID := newProc.Info(ctx).PID

					require.NoError(t, tracker.Add(newPID))
					pids, err := tracker.listCgroupPIDs()
					require.NoError(t, err)
					assert.Len(t, pids, 1)
				},
				"UpdateLimitChecksForLinuxResources": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					assert.Error(t, tracker.SetLimits("foobar"))
					assert.NoError(t, tracker.SetLimits(&LinuxResources{}))
				},
				"UpdateLimitKillsProcessWithZeroMemoryLimit": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					pid := proc.Info(ctx).PID

					zero := int64(0)
					require.NoError(t, tracker.SetLimits(&LinuxResources{
						Memory: &LinuxMemory{Limit: &zero},
					}))

					require.NoError(t, tracker.Add(pid))

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

					opts := yesCreateOpts(taskTimeout)
					proc, err := makeProc(ctx, &opts)
					require.NoError(t, err)

					tracker, err := NewProcessTracker("test")
					require.NoError(t, err)
					require.NotNil(t, tracker)
					linuxTracker, ok := tracker.(*linuxProcessTracker)
					require.True(t, ok)
					defer func() {
						// Ensure that the cgroup is cleaned up.
						assert.NoError(t, tracker.Cleanup())
					}()

					testCase(ctx, t, linuxTracker, proc)
				})
			}
		})
	}
}

func TestLinuxProcessTrackerWithEnvironmentVariables(t *testing.T) {
	for procName, makeProc := range map[string]ProcessConstructor{
		"Blocking": newBlockingProcess,
		"Basic":    newBasicProcess,
	} {
		t.Run(procName, func(t *testing.T) {
			for testName, testCase := range map[string]func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, opts *CreateOptions, envVarName string, envVarValue string){
				"CleanupFindsProcessesByEnvironmentVariable": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, opts *CreateOptions, envVarName string, envVarValue string) {
					opts.AddEnvVar(envVarName, envVarValue)
					proc, err := makeProc(ctx, opts)
					require.NoError(t, err)
					pid := proc.Info(ctx).PID
					assert.NoError(t, tracker.Add(pid))
					assert.NoError(t, tracker.Cleanup())

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
				"CleanupFindsProcessesWithoutAdd": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, opts *CreateOptions, envVarName string, envVarValue string) {
					opts.AddEnvVar(envVarName, envVarValue)
					proc, err := makeProc(ctx, opts)
					require.NoError(t, err)
					assert.NoError(t, tracker.Cleanup())

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
				"CleanupIgnoresAddedProcessesWithoutEnvironmentVariable": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, opts *CreateOptions, envVarName string, envVarValue string) {
					proc, err := makeProc(ctx, opts)
					require.NoError(t, err)
					pid := proc.Info(ctx).PID
					assert.NoError(t, tracker.Add(pid))
					assert.NoError(t, tracker.Cleanup())
					assert.True(t, proc.Running(ctx))
				},
				// "": func(ctx, context.Context, t *testing.T, tracker *linuxProcessTracker, envVarName string, envVarValue string) {},
			} {
				t.Run(testName, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), taskTimeout)
					defer cancel()

					envVarValue := "bar"

					opts := yesCreateOpts(taskTimeout)

					tracker, err := NewProcessTracker(envVarValue)
					require.NoError(t, err)
					require.NotNil(t, tracker)
					linuxTracker, ok := tracker.(*linuxProcessTracker)
					require.True(t, ok)
					defer func() {
						// Ensure that the cgroup is cleaned up.
						assert.NoError(t, tracker.Cleanup())
					}()
					// Ignore cgroup behavior.
					linuxTracker.cgroup = nil

					testCase(ctx, t, linuxTracker, &opts, ManagerEnvironID, envVarValue)
				})
			}
		})
	}
}

func TestManagerSetsEnvironmentVariables(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for managerName, makeManager := range map[string]func() *basicProcessManager{
		"Basic/NoLock/BasicProcs": func() *basicProcessManager {
			return &basicProcessManager{
				procs:    map[string]Process{},
				blocking: false,
				tracker:  newMockProcessTracker(),
			}
		},
		"Basic/NoLock/BlockingProcs": func() *basicProcessManager {
			return &basicProcessManager{
				procs:    map[string]Process{},
				blocking: true,
				tracker:  newMockProcessTracker(),
			}
		},
	} {
		t.Run(managerName, func(t *testing.T) {
			for testName, testCase := range map[string]func(context.Context, *testing.T, *basicProcessManager){
				"CreateProcessSetsManagerEnvironmentVariables": func(ctx context.Context, t *testing.T, manager *basicProcessManager) {
					opts := yesCreateOpts(managerTestTimeout)
					proc, err := manager.CreateProcess(ctx, &opts)
					require.NoError(t, err)
					pid := proc.Info(ctx).PID

					env, err := getEnvironmentVariables(pid)
					require.NoError(t, err)
					value, ok := env[ManagerEnvironID]
					require.True(t, ok)
					assert.Equal(t, value, manager.id, "process should have manager environment variable set")
				},
				"CreateCommandAddsEnvironmentVariables": func(ctx context.Context, t *testing.T, manager *basicProcessManager) {
					envVar := ManagerEnvironID
					value := manager.id

					cmdArgs := []string{"yes"}
					cmd := manager.CreateCommand(ctx).AddEnv(ManagerEnvironID, manager.id).Add(cmdArgs).Background(true)
					require.NoError(t, cmd.Run(ctx))

					ids := cmd.GetProcIDs()
					require.Len(t, ids, 1)
					proc, err := manager.Get(ctx, ids[0])
					require.NoError(t, err)
					pid := proc.Info(ctx).PID

					for {
						select {
						case <-ctx.Done():
							assert.Fail(t, "context timed out before environment variables were set for process")
							return
						default:
							if env, err := getEnvironmentVariables(pid); err == nil {
								if actualValue, ok := env[envVar]; ok {
									assert.Equal(t, value, actualValue)
									return
								}
							}
						}
					}
				},
			} {
				t.Run(testName, func(t *testing.T) {
					tctx, cancel := context.WithTimeout(ctx, managerTestTimeout)
					defer cancel()
					testCase(tctx, t, makeManager())
				})
			}
		})
	}
}
