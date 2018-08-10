package jasper

import (
	"context"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestBlockingProcess(t *testing.T) {
	t.Parallel()
	for name, testCase := range map[string]func(context.Context, *testing.T, *blockingProcess){
		"VerifyTestCaseConfiguration": func(ctx context.Context, t *testing.T, proc *blockingProcess) {
			assert.NotNil(t, proc)
			assert.NotNil(t, ctx)
			assert.NotZero(t, proc.ID())
			assert.NotNil(t, makeDefaultTrigger(ctx, nil, &proc.opts, "foo"))
		},
		"InfoIDPopulatedInBasicCase": func(ctx context.Context, t *testing.T, proc *blockingProcess) {
			proc.info = &ProcessInfo{
				ID: proc.ID(),
			}

			info := proc.Info(ctx)
			assert.Equal(t, info.ID, proc.ID())
		},

		// "": func(ctx context.Context, t *testing.T, proc *basicProcess) {},
		// "": func(ctx context.Context, t *testing.T, proc *basicProcess) {},
		// "": func(ctx context.Context, t *testing.T, proc *basicProcess) {},
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			testCase(ctx, t, &blockingProcess{
				id: uuid.Must(uuid.NewV4()).String(),
			})
		})

	}

}
