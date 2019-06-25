package jasper

import (
	"testing"

	"github.com/mongodb/jasper"
	"github.com/stretchr/testify/assert"
)

func TestMockInterfaces(t *testing.T) {
	manager := &MockManager{}
	_, ok := interface{}(manager).(jasper.Manager)
	assert.True(t, ok)

	process := &MockProcess{}
	_, ok = interface{}(process).(jasper.Process)
	assert.True(t, ok)

	remoteClient := &MockRemoteClient{}
	_, ok = interface{}(remoteClient).(jasper.RemoteClient)
	assert.True(t, ok)
}
