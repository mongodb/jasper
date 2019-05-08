package jasper

import (
	"context"

	"github.com/pkg/errors"
)

// SignalEvent signals the event object represented by the given name.
func SignalEvent(ctx context.Context, name string) error {
	event, err := GetEvent(name)
	if err != nil {
		return errors.Wrapf(err, "getting event '%s'", name)
	}
	defer event.Close()

	if err := event.Set(); err != nil {
		return errors.Wrapf(err, "signalling event '%s'", name)
	}

	return nil
}
