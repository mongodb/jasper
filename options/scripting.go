package options

// ScriptingEnvironment defines the interface for all types that
// define a scripting environment.
type ScriptingEnvironment interface {
	ID() string
	Type() string
	Interpreter() string
}
