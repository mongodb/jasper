// +build darwin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPids(t *testing.T) {
	log := "2018-10-03 21:55:21.478932+0000 0x16b Default 0x0 0 kernel: low swap: killing largest compressed process with pid 29670 (mongod) and size 1 MB"
	pid, hasPid = getPidFromLog(log)
	assert.True(hasPid)
	assert.Equal(t, 29670, pid)
}
