package jasper

import (
	"context"
)

type oomTrackerImpl struct {
	Lines []string `json:"lines"`
	PIDs  []int    `json:"pids"`
}

// OOMTracker provides a tool for detecting if there have been OOM
// events on the system. The Clear operation may affect the state the
// system logs and the data reported will reflect the entire system,
// not simply processes managed by Jasper tools.
type OOMTracker interface {
	Check(context.Context) error
	Clear(context.Context) error
	Report() ([]string, []int)
}

// NewOOMTracker returns an implementation of the OOMTracker interface
// for the current platform.
func NewOOMTracker() OOMTracker                     { return &oomTrackerImpl{} }
func (o *oomTrackerImpl) Report() ([]string, []int) { return o.Lines, o.PIDs }
