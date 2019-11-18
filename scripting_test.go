package jasper

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mongodb/grip"
	"github.com/mongodb/jasper/options"
	"github.com/stretchr/testify/require"
)

func isInPath(binary string) bool {
	_, err := exec.LookPath(binary)
	if err != nil {
		return false
	}
	return true
}

func TestScriptingEnvironment(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	manager, err := NewSynchronizedManager(false)
	require.NoError(t, err)
	defer manager.Close(ctx)

	tmpdir, err := ioutil.TempDir("", "scripting_tests")
	require.NoError(t, err)
	defer func() {
		grip.Error(os.RemoveAll(tmpdir))
	}()

	for _, env := range []struct {
		Name           string
		Supported      bool
		DefaultOptions options.ScriptingEnvironment
		Tests          map[string]func(*testing.T, options.ScriptingEnvironment)
	}{
		{
			Name:      "Roswell",
			Supported: isInPath("ros"),
			DefaultOptions: &options.ScriptingRoswell{
				Path: filepath.Join(tmpdir, "roswell", "factory"),
				Lisp: "sbcl-bin",
			},
			Tests: map[string]func(*testing.T, options.ScriptingEnvironment){
				"Factory": func(t *testing.T, opts options.ScriptingEnvironment) {
					se, err := manager.CreateScripting(ctx, opts)
					require.NoError(t, err)
					require.NotNil(t, se)
				},
			},
		},
		{
			Name:      "Python3",
			Supported: isInPath("python3"),
		},
		{
			Name:      "Python2",
			Supported: isInPath("python"),
		},
		{
			Name:      "Golang",
			Supported: isInPath("go"),
		},
	} {
		t.Run(env.Name, func(t *testing.T) {
			if !env.Supported {
				t.Skipf("%s is not supported in the current system", env.Name)
				return
			}
			for name, test := range env.Tests {
				t.Run(name, func(t *testing.T) {
					test(t, env.DefaultOptions)
				})
			}
		})
	}
}
