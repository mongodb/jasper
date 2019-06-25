package jasper

import (
	"context"
	"testing"

	"github.com/mongodb/grip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockInterfaces(t *testing.T) {
	manager := &MockManager{}
	_, ok := interface{}(manager).(Manager)
	assert.True(t, ok)

	process := &MockProcess{}
	_, ok = interface{}(process).(Process)
	assert.True(t, ok)

	remoteClient := &MockRemoteClient{}
	_, ok = interface{}(remoteClient).(RemoteClient)
	assert.True(t, ok)
}

func TestFoo(t *testing.T) {
	manager := &MockManager{FailCreate: true}
	_, err := manager.CreateProcess(context.Background(), &CreateOptions{})
	grip.Infof("err = %s", err.Error())
	require.Error(t, err)
}
