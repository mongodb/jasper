package executor

import (
	"context"
	"runtime"
	"testing"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDocker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := client.NewClientWithOpts()
	require.NoError(t, err)

	t.Run("SucceedsWithAllOptions", func(t *testing.T) {
		opts := DockerOptions{
			Client:  client,
			Image:   "image",
			Command: []string{"echo", "foo"},
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
		}
		exec, err := NewDocker(ctx, opts)
		require.NoError(t, err)
		assert.NotZero(t, exec)
	})
	t.Run("SucceedsWithMinimalOptions", func(t *testing.T) {
		opts := DockerOptions{
			Client:  client,
			Image:   "image",
			Command: []string{"echo", "foo"},
		}
		exec, err := NewDocker(ctx, opts)
		require.NoError(t, err)
		assert.NotZero(t, exec)
	})
	t.Run("FailsWithZeroOptions", func(t *testing.T) {
		exec, err := NewDocker(ctx, DockerOptions{})
		assert.Error(t, err)
		assert.Zero(t, exec)
	})
}

func TestDockerOptions(t *testing.T) {
	client, err := client.NewClientWithOpts()
	require.NoError(t, err)

	getAllPopulatedOpts := func() DockerOptions {
		return DockerOptions{
			Client:  client,
			Image:   "image",
			Command: []string{"echo", "foo"},
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
		}
	}

	t.Run("Validate", func(t *testing.T) {
		t.Run("SucceedsWithAllOptions", func(t *testing.T) {
			opts := getAllPopulatedOpts()
			assert.NoError(t, opts.Validate())
		})
		t.Run("SucceedsWithoutOS", func(t *testing.T) {
			opts := getAllPopulatedOpts()
			opts.OS = ""
			assert.NoError(t, opts.Validate())
		})
		t.Run("SucceedsWithoutArch", func(t *testing.T) {
			opts := getAllPopulatedOpts()
			opts.Arch = ""
			assert.NoError(t, opts.Validate())
		})
		t.Run("FailsWithZeroOptions", func(t *testing.T) {
			opts := DockerOptions{}
			assert.Error(t, opts.Validate())
		})
		t.Run("FailsWithoutClient", func(t *testing.T) {
			opts := getAllPopulatedOpts()
			opts.Client = nil
			assert.Error(t, opts.Validate())
		})
		t.Run("FailsWithoutImage", func(t *testing.T) {
			opts := getAllPopulatedOpts()
			opts.Image = ""
			assert.Error(t, opts.Validate())
		})
		t.Run("FailsWithoutCommand", func(t *testing.T) {
			opts := getAllPopulatedOpts()
			opts.Command = nil
			assert.Error(t, opts.Validate())
		})
	})
}
