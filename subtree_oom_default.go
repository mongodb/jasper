// +build !linux,!darwin

package tracker

import (
	"context"
)

// placeholders for windows tests

func (o *oomTrackerImpl) Clear(ctx context.Context) error {
	return nil
}

func (o *oomTrackerImpl) Check(ctx context.Context) error {
	return nil
}
