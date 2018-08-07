package jasper

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutputOptions(t *testing.T) {
	opts := OutputOptions{}
	assert.NoError(opts.Validate())

	stdout := bytes.NewBuffer([]byte{})
	stderr := bytes.NewBuffer([]byte{})

	opts.Output = stdout
	opts.Error = stderr
	assert.NoError(t, opts.Validate())

	// invalid if both streams are the same
	opts.Output = stderr
	assert.Error(t, opts.Validate())
	opts.Output = stdout
	assert.NoError(t, opts.Validate())

	// if the redirection and suppression options don't make
	// sense, validate should error, for stderr
	opts.SuppressError = true
	assert.Error(t, opts.Validate())
	opts.Error = nil
	assert.NoError(t, opts.Validate())
	opts.SendOutputToError = true
	assert.Error(t, opts.Validate())
	opts.SuppressError = false
	opts.Error = stderr
	assert.NoError(t, opts.Validate())
	opts.SendOutputToError = false
	assert.NoError(t, opts.Validate())

	// the same but for stdout
	opts.SuppressOutput = true
	assert.Error(t, opts.Validate())
	opts.Output = nil
	assert.NoError(t, opts.Validate())
	opts.SendErrorToOutput = true
	assert.Error(t, opts.Validate())
	opts.SuppressOutput = false
	opts.Output = stdout
	assert.NoError(t, opts.Validate())
	opts.SuppressOutput = false
	assert.NoError(t, opts.Validate())

	// but should be valid if you suppress both
	opts = OutputOptions{SuppressError: true, SuppressOutput: true}
	assert.NoError(t, opts.Validate())
}

func TestOutputOptionsIntegrationTableTest(t *testing.T) {
	buf := &bytes.Buffer{}
	shouldFail := []OutputOptions{
		{Output: buf, Error: buf},
		{Output: buf, SendOutputToError: true},
	}

	shouldPass := []OutputOptions{
		{SuppressError: true, SuppressOutput: true},
		{Output: buf, SendErrorToOutput: true},
	}

	for idx, opt := range shouldFail {
		assert.Error(t, opt.Validate(), "%d: %+v", idx, opt)
	}

	for idx, opt := range shouldPass {
		assert.NoError(t, opt.Validate(), "%d: %+v", idx, opt)
	}

}
