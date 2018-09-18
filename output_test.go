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
		"RedirectCannotHaveCycle": func(t *testing.T, opts OutputOptions) {
			opts.Output = stdout
			opts.Error = stderr
			opts.SendOutputToError = true
			opts.SendErrorToOutput = true
			assert.Error(t, opts.Validate())
		},
		"ValidateFailsForInvalidLogTypes": func(t *testing.T, opts OutputOptions) {
			opts.Loggers = []Logger{Logger{LogType: LogType("")}}
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

func TestLogTypes(t *testing.T) {
	type testCase func(*testing.T, LogType, LogOptions)
	cases := map[string]testCase{
		"NonexistentLogTypeIsInvalid": func(t *testing.T, l LogType, opts LogOptions) {
			l = LogType("")
			assert.Error(t, l.Validate())
		},
		"ValidLogTypePasses": func(t *testing.T, l LogType, opts LogOptions) {
			assert.NoError(t, l.Validate())
		},
		"ConfigureFailsForInvalidLogType": func(t *testing.T, l LogType, opts LogOptions) {
			l = LogType("foo")
			sender, err := l.Configure(opts)
			assert.Error(t, err)
			assert.Nil(t, sender)
		},
		"ConfigurePassesWithLogDefault": func(t *testing.T, l LogType, opts LogOptions) {
			sender, err := l.Configure(opts)
			assert.NoError(t, err)
			assert.NotNil(t, sender)
		},
		"ConfigurePassesWithLogInherit": func(t *testing.T, l LogType, opts LogOptions) {
			l = LogInherit
			sender, err := l.Configure(opts)
			assert.NoError(t, err)
			assert.NotNil(t, sender)
		},
		"ConfigureFailsWithoutPopulatedSplunkOptions": func(t *testing.T, l LogType, opts LogOptions) {
			l = LogSplunk
			sender, err := l.Configure(opts)
			assert.Error(t, err)
			assert.Nil(t, sender)
		},
		"ConfigurePassesWithPopulatedSplunkOptions": func(t *testing.T, l LogType, opts LogOptions) {
			l = LogSplunk
			opts.SplunkOptions = send.SplunkConnectionInfo{ServerURL: "foo", Token: "bar"}
			sender, err := l.Configure(opts)
			assert.NoError(t, err)
			assert.NotNil(t, sender)
		},
		"ConfigureFailsWithoutLocalInBuildloggerOptions": func(t *testing.T, l LogType, opts LogOptions) {
			l = LogBuildloggerV2
			sender, err := l.Configure(opts)
			assert.Error(t, err)
			assert.Nil(t, sender)
		},
		"ConfigureFailsWithoutSetNameInBuildloggerOptions": func(t *testing.T, l LogType, opts LogOptions) {
			l = LogBuildloggerV2
			opts.BuildloggerOptions = send.BuildloggerConfig{Local: send.MakeNative()}
			sender, err := l.Configure(opts)
			assert.Error(t, err)
			assert.Nil(t, sender)
		},
		"ConfigureFailsWithoutPopulatedSumologicOptions": func(t *testing.T, l LogType, opts LogOptions) {
			l = LogSumologic
			sender, err := l.Configure(opts)
			assert.Error(t, err)
			assert.Nil(t, sender)
		},
		"ConfigureLogFilePasses": func(t *testing.T, l LogType, opts LogOptions) {
			l = LogFile
			file, err := ioutil.TempFile("build", "foo.txt")
			require.NoError(t, err)
			defer os.Remove(file.Name())
			opts.FileName = file.Name()

			sender, err := l.Configure(opts)
			assert.NoError(t, err)
			assert.NotNil(t, sender)
		},
		// "": func(t *testing.T, l LogType, opts LogOptions) {},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			test(t, LogDefault, LogOptions{})
		})
	}
}
