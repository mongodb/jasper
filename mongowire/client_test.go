package mongowire

import (
	"context"
	"net"
	"testing"

	"github.com/mongodb/jasper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// kim: TODO: get this to pass
func TestID(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addr, err := net.ResolveTCPAddr("tcp", "localhost:12345")
	require.NoError(t, err)

	mgr, err := jasper.NewSynchronizedManager(false)
	require.NoError(t, err)

	svc, err := NewService(mgr, addr)
	require.NoError(t, err)
	go func() {
		svc.Run(ctx)
	}()

	c, err := NewClient(addr)
	require.NoError(t, err)
	defer c.CloseConnection()

	id := c.ID()
	assert.NotEmpty(t, id)
}
