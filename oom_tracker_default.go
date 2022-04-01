// +build !linux,!darwin

package jasper

import (
	"context"
)

// These are placeholder implementations for platforms that don't support the
// OOM killer.

func (o *oomTrackerImpl) Clear(ctx context.Context) error {
	return nil
}

func (o *oomTrackerImpl) Check(ctx context.Context) error {
	return nil
}
