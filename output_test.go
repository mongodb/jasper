package jasper

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutputOptions(t *testing.T) {
	stdout := bytes.NewBuffer([]byte{})
	stderr := bytes.NewBuffer([]byte{})

	type testCase func(*testing.T, OutputOptions)

	cases := map[string]testCase{
		"NilOptionsValidate": func(t *testing.T, opts OutputOptions) {
			assert.Zero(t, opts)
			assert.NoError(t, opts.Validate())
		},
		"ErrorOutputSpecified": func(t *testing.T, opts OutputOptions) {
			opts.Output = stdout
			opts.Error = stderr
			assert.NoError(t, opts.Validate())
		},
		"StreamsMustBeDifferent": func(t *testing.T, opts OutputOptions) {
			// invalid if both streams are the same
			opts.Error = stderr
			opts.Output = stderr
			assert.Error(t, opts.Validate())
		},
		"SuppressErrorWhenSpecified": func(t *testing.T, opts OutputOptions) {
			opts.Error = stderr
			opts.SuppressError = true
			assert.Error(t, opts.Validate())
		},
		"SuppressOutputWhenSpecified": func(t *testing.T, opts OutputOptions) {
			opts.Output = stdout
			opts.SuppressOutput = true
			assert.Error(t, opts.Validate())
		},
		"RedirectErrorToNillFails": func(t *testing.T, opts OutputOptions) {
			opts.SendOutputToError = true
			assert.Error(t, opts.Validate())
		},
		"RedirectOutputToError": func(t *testing.T, opts OutputOptions) {
			opts.SendOutputToError = true
			assert.Error(t, opts.Validate())
		},
		"SuppressAndRedirectOutputIsInvalid": func(t *testing.T, opts OutputOptions) {
			opts.SuppressOutput = true
			opts.SendOutputToError = true
			assert.Error(t, opts.Validate())
		},
		"SuppressAndRedirectErrorIsInvalid": func(t *testing.T, opts OutputOptions) {
			opts.SuppressError = true
			opts.SendErrorToOutput = true
			assert.Error(t, opts.Validate())
		},
		"DiscardIsNilForOutput": func(t *testing.T, opts OutputOptions) {
			opts.Error = stderr
			opts.Output = ioutil.Discard

			assert.True(t, opts.outputIsNull())
			assert.False(t, opts.errorIsNull())
		},
		"NilForOutputIsValid": func(t *testing.T, opts OutputOptions) {
			opts.Error = stderr
			assert.True(t, opts.outputIsNull())
			assert.False(t, opts.errorIsNull())
		},
		"DiscardIsNilForError": func(t *testing.T, opts OutputOptions) {
			opts.Error = ioutil.Discard
			opts.Output = stdout
			assert.True(t, opts.errorIsNull())
			assert.False(t, opts.outputIsNull())
		},
		"NilForErrorIsValid": func(t *testing.T, opts OutputOptions) {
			opts.Output = stdout
			assert.True(t, opts.errorIsNull())
			assert.False(t, opts.outputIsNull())
		},
		"OutputGetterNilIsIoDiscard": func(t *testing.T, opts OutputOptions) {
			assert.Equal(t, ioutil.Discard, opts.GetOutput())
		},
		"OutputGetterWhenPopulatedIsCorrect": func(t *testing.T, opts OutputOptions) {
			opts.Output = stdout
			assert.Equal(t, stdout, opts.GetOutput())
		},
		"ErrorGetterNilIsIoDiscard": func(t *testing.T, opts OutputOptions) {
			assert.Equal(t, ioutil.Discard, opts.GetError())
		},
		"ErrorGetterWhenPopulatedIsCorrect": func(t *testing.T, opts OutputOptions) {
			opts.Error = stderr
			assert.Equal(t, stderr, opts.GetError())
		},
		"RedirectErrorHasCorrectSemantics": func(t *testing.T, opts OutputOptions) {
			opts.Output = stdout
			opts.Error = stderr
			opts.SendErrorToOutput = true
			assert.Equal(t, stdout, opts.GetError())

		},
		"RedirectOutputHasCorrectSemantics": func(t *testing.T, opts OutputOptions) {
			opts.Output = stdout
			opts.Error = stderr
			opts.SendOutputToError = true
			assert.Equal(t, stderr, opts.GetOutput())
		},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			test(t, OutputOptions{})
		})
	}
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
