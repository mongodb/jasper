package options

import (
	"github.com/evergreen-ci/utility"
	"github.com/pkg/errors"
)

// ScriptingHarness defines the interface for all types that
// define a scripting environment.
type ScriptingHarness interface {
	// ID should return a unique hash of the implementation of
	// ScrptingEnvironment. This can be cached, and should change
	// if any of the dependencies change.
	ID() string
	// Type returns the name of the environment, and is useful to
	// identify the environment for users.
	Type() string
	// Interpreter should return a path to the binary that will be
	// used for running code.
	Interpreter() string
	// Validate checks the internal consistency of an
	// implementation and may set defaults.
	Validate() error
}

// ScriptingHarnessFactory represents a factory for producing scripting
// harnesses of a particular type.
type ScriptingHarnessFactory interface {
	// Name returns the preferred name for the scripting harness.
	Name() string
	// Names are all recognized names for the scripting harness.
	Names() []string
	// New returns a new unconfigured instance of the scripting harness type.
	New() ScriptingHarness
}

// AllScriptingHarnesses returns all supported scripting harnesses.
func AllScriptingHarnesses() []ScriptingHarnessFactory {
	return []ScriptingHarnessFactory{
		Golang(),
		Python2(),
		Python3(),
		Roswell(),
	}
}

// MatchesScriptingHarness returns whether or not the scripting harness factory
// matches the given named scripting environment.
func MatchesScriptingHarness(kind ScriptingHarnessFactory, name string) bool {
	return utility.StringSliceContains(kind.Names(), name)
}

// NewScriptingHarness provides a factory to generate concrete
// implementations of the ScriptingEnvironment interface for use in
// marshaling arbitrary values for a known environment.
func NewScriptingHarness(se string) (ScriptingHarness, error) {
	for _, harness := range AllScriptingHarnesses() {
		if MatchesScriptingHarness(harness, se) {
			return harness.New(), nil
		}
	}
	return nil, errors.Errorf("no supported scripting environment named '%s'", se)
}
