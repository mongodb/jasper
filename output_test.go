package jasper

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/mongodb/grip/send"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputOptions(t *testing.T) {
	stdout := bytes.NewBuffer([]byte{})
	stderr := bytes.NewBuffer([]byte{})

	type testCase func(*testing.T, OutputOptions)

	cases := map[string]testCase{
		"NilOptionsDoNotValidate": func(t *testing.T, opts OutputOptions) {
			assert.Zero(t, opts)
			assert.Error(t, opts.Validate())
		},
		"ErrorOutputSpecified": func(t *testing.T, opts OutputOptions) {
			opts.Output = stdout
			opts.Error = stderr
			opts.LogType = LogDefault
			assert.NoError(t, opts.Validate())
		},
		"StreamsMustBeDifferent": func(t *testing.T, opts OutputOptions) {
			// invalid if both streams are the same
			opts.Error = stderr
			opts.Output = stderr
			opts.LogType = LogDefault
			assert.Error(t, opts.Validate())
		},
		"SuppressErrorWhenSpecified": func(t *testing.T, opts OutputOptions) {
			opts.Error = stderr
			opts.SuppressError = true
			opts.LogType = LogDefault
			assert.Error(t, opts.Validate())
		},
		"SuppressOutputWhenSpecified": func(t *testing.T, opts OutputOptions) {
			opts.Output = stdout
			opts.SuppressOutput = true
			opts.LogType = LogDefault
			assert.Error(t, opts.Validate())
		},
		"RedirectErrorToNillFails": func(t *testing.T, opts OutputOptions) {
			opts.SendOutputToError = true
			opts.LogType = LogDefault
			assert.Error(t, opts.Validate())
		},
		"RedirectOutputToError": func(t *testing.T, opts OutputOptions) {
			opts.SendOutputToError = true
			opts.LogType = LogDefault
			assert.Error(t, opts.Validate())
		},
		"SuppressAndRedirectOutputIsInvalid": func(t *testing.T, opts OutputOptions) {
			opts.SuppressOutput = true
			opts.SendOutputToError = true
			opts.LogType = LogDefault
			assert.Error(t, opts.Validate())
		},
		"SuppressAndRedirectErrorIsInvalid": func(t *testing.T, opts OutputOptions) {
			opts.SuppressError = true
			opts.SendErrorToOutput = true
			opts.LogType = LogDefault
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
		"RedirectCannotHaveCycle": func(t *testing.T, opts OutputOptions) {
			opts.Output = stdout
			opts.Error = stderr
			opts.SendOutputToError = true
			opts.SendErrorToOutput = true
			opts.LogType = LogDefault
			assert.Error(t, opts.Validate())
		},
		// "": func(t *testing.T, opts OutputOptions) {}
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
		{Output: buf, Error: buf, LogType: LogDefault},
		{Output: buf, SendOutputToError: true, LogType: LogDefault},
	}

	shouldPass := []OutputOptions{
		{SuppressError: true, SuppressOutput: true, LogType: LogDefault},
		{Output: buf, SendErrorToOutput: true, LogType: LogDefault},
	}

	for idx, opt := range shouldFail {
		assert.Error(t, opt.Validate(), "%d: %+v", idx, opt)
	}

	for idx, opt := range shouldPass {
		assert.NoError(t, opt.Validate(), "%d: %+v", idx, opt)
	}

}

func TestLogTypes(t *testing.T) {
	type testCase func(*testing.T, LogType, OutputOptions)
	cases := map[string]testCase{
		"NonexistentLogTypeIsInvalid": func(t *testing.T, l LogType, opts OutputOptions) {
			l = LogType("")
			assert.Error(t, l.Validate())
		},
		"ValidLogTypePasses": func(t *testing.T, l LogType, opts OutputOptions) {
			assert.NoError(t, l.Validate())
		},
		"ConfigureFailsForInvalidLogType": func(t *testing.T, l LogType, opts OutputOptions) {
			l = LogType("foo")
			err := l.Configure(&opts)
			assert.Error(t, err)
		},
		"ConfigurePassesWithLogDefault": func(t *testing.T, l LogType, opts OutputOptions) {
			err := l.Configure(&opts)
			assert.NoError(t, err)
		},
		"ConfigurePassesWithLogInherit": func(t *testing.T, l LogType, opts OutputOptions) {
			l = LogInherit
			err := l.Configure(&opts)
			assert.NoError(t, err)
		},
		"ConfigureErrorsWithoutPopulatedSplunkOptions": func(t *testing.T, l LogType, opts OutputOptions) {
			l = LogSplunk
			err := l.Configure(&opts)
			assert.Error(t, err)
		},
		"ConfigurePassesWithPopulatedSplunkOptions": func(t *testing.T, l LogType, opts OutputOptions) {
			l = LogSplunk
			opts.LogOptions.SplunkOptions = send.SplunkConnectionInfo{ServerURL: "foo", Token: "bar"}
			err := l.Configure(&opts)
			assert.NoError(t, err)
		},
		"ConfigureErrorsWithoutLocalInBuildloggerOptions": func(t *testing.T, l LogType, opts OutputOptions) {
			l = LogBuildloggerV2
			err := l.Configure(&opts)
			assert.Error(t, err)
		},
		"ConfigureErrorsWithoutSetNameInBuildloggerOptions": func(t *testing.T, l LogType, opts OutputOptions) {
			l = LogBuildloggerV2
			opts.LogOptions.BuildloggerOptions = send.BuildloggerConfig{Local: send.MakeNative()}
			err := l.Configure(&opts)
			assert.Error(t, err)
		},
		"ConfigureErrorsWithoutPopulatedSumologicOptions": func(t *testing.T, l LogType, opts OutputOptions) {
			l = LogSumologic
			err := l.Configure(&opts)
			assert.Error(t, err)
		},
		"ConfigureLogFilePasses": func(t *testing.T, l LogType, opts OutputOptions) {
			l = LogFile
			file, err := ioutil.TempFile("build", "foo.txt")
			require.NoError(t, err)
			defer os.Remove(file.Name())
			opts.LogOptions.FileName = file.Name()

			err = l.Configure(&opts)
			assert.NoError(t, err)
		},
		"ConfigureSetsSameOutputAndErrorLogger": func(t *testing.T, l LogType, opts OutputOptions) {
			l.Configure(&opts)
			assert.NotNil(t, opts.OutputLogger)
			assert.NotNil(t, opts.ErrorLogger)
			assert.Equal(t, opts.OutputLogger, opts.ErrorLogger)
		},
		"LoggerIgnoresStandardOutput": func(t *testing.T, l LogType, opts OutputOptions) {
			opts.LogOptions.IgnoreOutput = true
			l.Configure(&opts)
			assert.NotNil(t, opts.OutputLogger)
			assert.NotNil(t, opts.ErrorLogger)
			assert.NotEqual(t, opts.OutputLogger, opts.ErrorLogger)
		},
		"LoggerIgnoresStandardError": func(t *testing.T, l LogType, opts OutputOptions) {
			opts.LogOptions.IgnoreError = true
			l.Configure(&opts)
			assert.NotNil(t, opts.OutputLogger)
			assert.NotNil(t, opts.ErrorLogger)
			assert.NotEqual(t, opts.OutputLogger, opts.ErrorLogger)
		},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			test(t, LogDefault, OutputOptions{})
		})
	}
}
