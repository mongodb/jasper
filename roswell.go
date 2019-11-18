package jasper

import (
	"context"
	"errors"

	"github.com/mongodb/jasper/options"
)

type roswellEnvironment struct {
	opts *options.ScriptingRoswell

	isConfigured bool
	cachedHash   string
	manager      Manager
}

func (e *roswellEnvironment) ID() string { e.cachedHash = e.opts.ID(); return e.cachedHash }
func (e *roswellEnvironment) Setup(ctx context.Context) error {
	if e.isConfigured && e.cachedHash == e.opts.ID() {
		return nil
	}

	return errors.New("not implemented")
}
