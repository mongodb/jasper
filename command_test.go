package jasper

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/mongodb/grip/level"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCommand(t *testing.T) {
	assert.NoError(
		t,
		RunCommand(
			context.Background(),
			"test",
			level.Info,
			[]string{"echo", "hello world"},
			"/Users/may/quick/",
			map[string]string{},
		),
	)
}

func TestCommandStdOutAndStdErr(t *testing.T) {
	myCmd := NewCommand()
	msg1, msg2 := "lalala", "second"
	testFile := "test.txt"
	myCmd.Add([]string{"echo", msg1})
	myCmd.Add([]string{"echo", msg2})
	myCmd.Add([]string{"ls", "DNE"})

	file, err := os.Create(testFile)
	require.NoError(t, err)
	myCmd.SetOutputWriter(file)

	assert.NoError(t, myCmd.Run(context.Background()))

	fileBytes, err := ioutil.ReadFile("test.txt")
	require.NoError(t, err)
	commandsOut := string(fileBytes)
	assert.True(t, strings.Contains(commandsOut, "lalala"))
	assert.True(t, strings.Contains(commandsOut, "second"))
	assert.True(t, strings.Contains(commandsOut, "No such file or directory"))
}

//
//func TestRemoteCommandForProcess(t *testing.T) {
//	opts := CreateOptions{
//		Args:   []string{"ssh", "root@ec2-54-198-151-1.compute-1.amazonaws.com", "cd", ".composer/", "&&", "ls"},
//		Output: OutputOptions{Output: os.Stdout},
//	}
//	ctx := context.Background()
//	remoteProc, err := newBasicProcess(ctx, &opts)
//	require.NoError(t, err)
//	remoteProc.Wait(ctx)
//}
