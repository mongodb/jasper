//go:build !unix

package executor

// SetGroupLeader is a noop on non-unix systems.
func (e *local) SetGroupLeader() {}
