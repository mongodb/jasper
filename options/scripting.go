package options

import "github.com/pkg/errors"

// ScriptingEnvironment defines the interface for all types that
// define a scripting environment.
type ScriptingEnvironment interface {
	ID() string
	Type() string
	Interpreter() string
	Validate() error
}

func NewScriptingEnvironment(se string) (ScriptingEnvironment, error) {
	switch se {
	case "python2":
		return &ScriptingPython{LegacyPython: true}, nil
	case "python", "python3":
		return &ScriptingPython{LegacyPython: false}, nil
	case "go", "golang":
		return &ScriptingGolang{}, nil
	case "roswell", "ros", "lisp", "cl":
		return &ScriptingRoswell{}, nil
	default:
		return nil, errors.Errorf("no supported scripting environment named '%s'", se)
	}
}
