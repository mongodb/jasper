package jrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientPlaceholder(t *testing.T) {
	assert.NotPanics(t, func() {
		NewJRPCManager(nil)
	})
}
