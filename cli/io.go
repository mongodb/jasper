package cli

import (
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/level"
	"github.com/mongodb/jasper"
	"github.com/pkg/errors"
)

// Validator represents an input that can be validated.
type Validator interface {
	Validate() error
}

// ErrorResponse represents CLI-specific output containing an error message.
type ErrorResponse struct {
	Error error `json:"error"`
}

// InfoResponse represents represents CLI-specific output containing process
// information and an error message.
type InfoResponse struct {
	Info jasper.ProcessInfo `json:"info"`
}

// InfosResponse represents CLI-specific output containing information for
// multiple processes and an error message.
type InfosResponse struct {
	Infos []jasper.ProcessInfo `json:"infos"`
}

// TagsResponse represents CLI-specific output containing a list of process tags
// and an error message.
type TagsResponse struct {
	Tags []string `json:"tags"`
}

// RunningResponse represents CLI-specific output for whether the process is
// running or not.
type RunningResponse struct {
	Running bool `json:"running"`
}

// CompleteResponse represents CLI-specific output for whether the process is
// complete or not.
type CompleteResponse struct {
	Complete bool `json:"complete"`
}

// WaitResponse represents CLI-specific output for the wait exit code.
type WaitResponse struct {
	ExitCode int `json:"exit_code"`
}

// IDInput represents CLI-specific input representing a Jasper process ID.
type IDInput struct {
	ID string `json:"id"`
}

// Validate checks that the Jasper process ID is non-empty.
func (id *IDInput) Validate() error {
	if len(id.ID) == 0 {
		return errors.New("Jasper process ID must not be empty")
	}
	return nil
}

// SignalInput represents CLI-specific input to signal a Jasper process.
type SignalInput struct {
	ID     string `json:"id"`
	Signal int    `json:"signal"`
}

// Validate checks that the SignalInput has a non-empty Jasper process ID and
// positive Signal.
func (sig *SignalInput) Validate() error {
	catcher := grip.NewBasicCatcher()
	if len(sig.ID) == 0 {
		catcher.Add(errors.New("Jasper process ID must not be empty"))
	}
	if sig.Signal <= 0 {
		catcher.Add(errors.New("signal must be greater than 0"))
	}
	return catcher.Resolve()
}

// SignalTriggerIDInput represents CLI-specific input to attach a signal trigger
// to a Jasper process.
type SignalTriggerIDInput struct {
	ID              string                 `json:"id"`
	SignalTriggerID jasper.SignalTriggerID `json:"signal_trigger_id"`
}

// Validate checks that the SignalTriggerIDInput has a non-empty Jasper process
// ID and a recognized signal trigger ID.
func (sig *SignalTriggerIDInput) Validate() error {
	catcher := grip.NewBasicCatcher()
	if len(sig.ID) == 0 {
		catcher.Add(errors.New("Jasper process ID must not be empty"))
	}
	_, ok := jasper.GetSignalTriggerFactory(sig.SignalTriggerID)
	if !ok {
		return errors.Errorf("could not find signal trigger with id '%s'", sig.SignalTriggerID)
	}
	return nil
}

// CommandInput represents CLI-specific input to create a jasper.Command.
type CommandInput struct {
	Background      bool                 `json:"background"`
	CreateOptions   jasper.CreateOptions `json:"create_options"`
	Priority        level.Priority       `json:"priority"`
	ContinueOnError bool                 `json:"continue_on_error"`
	IgnoreError     bool                 `json:"ignore_error"`
}

// Validate checks that the input to the jasper.Command is valid.
func (c *CommandInput) Validate() error {
	catcher := grip.NewBasicCatcher()
	catcher.Add(c.CreateOptions.Validate())
	if c.Priority != 0 && !level.IsValidPriority(c.Priority) {
		catcher.Add(errors.New("priority is not in the valid range of values"))
	}
	return catcher.Resolve()
}

// TagIDInput represents the CLI-specific input for a process with a given tag.
type TagIDInput struct {
	ID  string `json:"id"`
	Tag string `json:"tag"`
}

// Validate checks that the TagIDInput has a non-empty Jasper process ID and a
// non-empty tag.
func (t *TagIDInput) Validate() error {
	if len(t.ID) == 0 {
		return errors.New("Jasper process ID must not be empty")
	}
	if len(t.Tag) == 0 {
		return errors.New("tag must not be empty")
	}
	return nil
}

// TagInput represents the CLI-specific input for process tags.
type TagInput struct {
	Tag string `json:"tag"`
}

// Validate checks that the tag is non-empty.
func (t *TagInput) Validate() error {
	if len(t.Tag) == 0 {
		return errors.New("tag must not be empty")
	}
	return nil
}

// FilterInput represents the CLI-specific input to filter processes.
type FilterInput struct {
	Filter jasper.Filter
}

// Validate checks that the jasper.Filter is a recognized filter.
func (f *FilterInput) Validate() error {
	return f.Filter.Validate()
}
