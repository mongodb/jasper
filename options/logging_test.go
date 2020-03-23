package options

import (
	"testing"

	"github.com/mongodb/grip/level"
	"github.com/mongodb/grip/send"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogging(t *testing.T) {
	t.Run("LoggingSendErrors", func(t *testing.T) {
		lp := &LoggingPayload{}

		for _, err := range []error{
			lp.Send(nil),
			lp.Send(&CachedLogger{}),
		} {
			require.Error(t, err)
			require.Equal(t, "no output configured", err.Error())
		}
	})
	t.Run("OutputTargeting", func(t *testing.T) {
		output := send.MakeInternalLogger()
		error := send.MakeInternalLogger()
		lp := &LoggingPayload{Data: "hello world!", Priority: level.Info}
		t.Run("Output", func(t *testing.T) {
			assert.Equal(t, 0, output.Len())
			cl := &CachedLogger{Output: output}
			require.NoError(t, lp.Send(cl))
			require.Equal(t, 1, output.Len())
			msg := output.GetMessage()
			assert.Equal(t, "hello world!", msg.Message.String())
		})
		t.Run("Error", func(t *testing.T) {
			assert.Equal(t, 0, error.Len())
			cl := &CachedLogger{Error: error}
			require.NoError(t, lp.Send(cl))
			require.Equal(t, 1, error.Len())
			msg := error.GetMessage()
			assert.Equal(t, "hello world!", msg.Message.String())
		})
		t.Run("ErrorForce", func(t *testing.T) {
			lp.ForceSendToError = true
			assert.Equal(t, 0, error.Len())
			cl := &CachedLogger{Error: error, Output: output}
			require.NoError(t, lp.Send(cl))
			require.Equal(t, 1, error.Len())
			msg := error.GetMessage()
			assert.Equal(t, "hello world!", msg.Message.String())
		})

	})

}
