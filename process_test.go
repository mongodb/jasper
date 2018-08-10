package jasper

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type processConstructor func(context.Context, *CreateOptions) (Process, error)

func makeLockingProcess(pmake processConstructor) processConstructor {
	return func(ctx context.Context, opts *CreateOptions) (Process, error) {
		proc, err := pmake(ctx, opts)
		if err != nil {
			return nil, err
		}
		return &localProcess{proc: proc}, nil
	}
}

func TestProcessImplementations(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for cname, makeProc := range map[string]processConstructor{
		"BlockingNoLock":   newBlockingProcess,
		"BlockingWithLock": makeLockingProcess(newBlockingProcess),
		"BasicNoLock":      newBasicProcess,
		"BasicWithLock":    makeLockingProcess(newBasicProcess),
	} {
		for name, testCase := range map[string]func(*testing.T, *CreateOptions, processConstructor){
			"WithPopulatedArgsCommandCreationPasses": func(t *testing.T, opts *CreateOptions, makep processConstructor) {
				assert.NotZero(t, opts.Args)
				proc, err := makep(ctx, opts)
				assert.NoError(t, err)
				assert.NotNil(t, proc)
			},
			"ErrorToCreateWithInvalidArgs": func(t *testing.T, opts *CreateOptions, makep processConstructor) {
				opts.Args = []string{}
				proc, err := makep(ctx, opts)
				assert.Error(t, err)
				assert.Nil(t, proc)
			},
			"WithCancledContextProcessCreationFailes": func(t *testing.T, opts *CreateOptions, makep processConstructor) {
				pctx, pcancel := context.WithCancel(ctx)
				pcancel()
				proc, err := makep(pctx, opts)
				assert.Error(t, err)
				assert.Nil(t, proc)
			},
			"CanceledContextTimesOutEarly": func(t *testing.T, opts *CreateOptions, makep processConstructor) {
				pctx, pcancel := context.WithTimeout(ctx, 200*time.Millisecond)
				defer pcancel()
				startAt := time.Now()
				opts.Args = []string{"sleep", "101"}
				proc, err := makep(pctx, opts)
				assert.NoError(t, err)

				time.Sleep(100 * time.Millisecond) // let time pass...
				require.NotNil(t, proc)
				assert.False(t, proc.Info(ctx).Successful)
				assert.True(t, time.Since(startAt) < 400*time.Millisecond)
			},
			"ProcessLacksTagsByDefault": func(t *testing.T, opts *CreateOptions, makep processConstructor) {
				proc, err := makep(ctx, opts)
				assert.NoError(t, err)
				tags := proc.GetTags()
				assert.Empty(t, tags)
			},
			"ProcessTagsPersist": func(t *testing.T, opts *CreateOptions, makep processConstructor) {
				opts.Tags = []string{"foo"}
				proc, err := makep(ctx, opts)
				assert.NoError(t, err)
				tags := proc.GetTags()
				assert.Contains(t, tags, "foo")
			},
			"InfoHasMatchingID": func(t *testing.T, opts *CreateOptions, makep processConstructor) {
				proc, err := makep(ctx, opts)
				assert.NoError(t, err)
				assert.Equal(t, proc.ID(), proc.Info(ctx).ID)
			},
			"ResetTags": func(t *testing.T, opts *CreateOptions, makep processConstructor) {
				proc, err := makep(ctx, opts)
				assert.NoError(t, err)
				proc.Tag("foo")
				assert.Contains(t, proc.GetTags(), "foo")
				proc.ResetTags()
				assert.Len(t, proc.GetTags(), 0)
			},
			"TagsAreSetLike": func(t *testing.T, opts *CreateOptions, makep processConstructor) {
				proc, err := makep(ctx, opts)
				assert.NoError(t, err)
				for i := 0; i < 100; i++ {
					proc.Tag("foo")
				}

				assert.Len(t, proc.GetTags(), 1)
				proc.Tag("bar")
				assert.Len(t, proc.GetTags(), 2)
			},
			"CompleteIsTrueAfterWait": func(t *testing.T, opts *CreateOptions, makep processConstructor) {
				proc, err := makep(ctx, opts)
				assert.NoError(t, err)
				time.Sleep(10 * time.Millisecond) // give the process time to start background machinery
				assert.NoError(t, proc.Wait(ctx))
				assert.True(t, proc.Complete(ctx))
			},
			"WaitReturnsWithCancledContext": func(t *testing.T, opts *CreateOptions, makep processConstructor) {

				opts.Args = []string{"sleep", "101"}
				pctx, pcancel := context.WithCancel(ctx)
				proc, err := makep(ctx, opts)
				assert.NoError(t, err)
				pcancel()
				assert.Error(t, proc.Wait(pctx))

			},
			// "": func(t *testing.T, opts *CreateOptions, makep processConstructor) {},
			// "": func(t *testing.T, opts *CreateOptions, makep processConstructor) {},
		} {
			t.Run(cname+name, func(t *testing.T) {
				opts := &CreateOptions{Args: []string{"ls"}}
				testCase(t, opts, makeProc)
			})
		}
	}

}
