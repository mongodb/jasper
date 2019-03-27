// +build linux

package jasper

import (
	"context"
	"testing"

	"github.com/mongodb/grip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinuxProcessTracker(t *testing.T) {
	tracker, err := newProcessTracker("jasper")
	require.NoError(t, err)
	require.NotNil(t, tracker)
	for procName, makeProc := range map[string]ProcessConstructor{
		"Blocking": newBlockingProcess,
		"Basic":    newBasicProcess,
	} {
		t.Run(procName, func(t *testing.T) {

			for name, testCase := range map[string]func(context.Context, *testing.T, *linuxProcessTracker, Process){
				"AddNewProcessSucceeds": func(ctx context.Context, t *testing.T, tracker *linuxProcessTracker, proc Process) {
					pid := proc.Info(ctx).PID
					assert.NoError(t, tracker.add(pid))
					subsystems := tracker.cgroup.Subsystems()
					for _, subsystem := range subsystems {
						grip.Infof("kim: subsystem = %+v", subsystem.Name())
					}
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

					testCase(ctx, t, linuxTracker, proc)
				})
			}
		})
	}
}
