package cli

import (
	"testing"

	"github.com/mongodb/jasper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSSHManager(t *testing.T) {
	for testName, testCase := range map[string]func(t *testing.T, remoteOpts jasper.RemoteOptions, clientOpts ClientOptions){
		"NewSSHManagerFailsWithEmptyRemoteOptions": func(t *testing.T, remoteOpts jasper.RemoteOptions, clientOpts ClientOptions) {
			remoteOpts = jasper.RemoteOptions{}
			_, err := NewSSHManager(remoteOpts, clientOpts)
			assert.Error(t, err)
		},
		"NewSSHManagerFailsWithEmptyClientOptions": func(t *testing.T, remoteOpts jasper.RemoteOptions, clientOpts ClientOptions) {
			clientOpts = ClientOptions{}
			_, err := NewSSHManager(remoteOpts, clientOpts)
			assert.Error(t, err)
		},
		"NewSSHManagerSucceedsWithPopulatedOptions": func(t *testing.T, remoteOpts jasper.RemoteOptions, clientOpts ClientOptions) {
			manager, err := NewSSHManager(remoteOpts, clientOpts)
			require.NoError(t, err)
			assert.NotNil(t, manager)
		},
	} {
		t.Run(testName, func(t *testing.T) {
			remoteOpts := jasper.RemoteOptions{
				User: "kim",
				Host: "localhost",
			}

			clientOpts := ClientOptions{
				BinaryPath: "binary",
				Type:       RPCService,
			}

			testCase(t, remoteOpts, clientOpts)
		})
	}
}
