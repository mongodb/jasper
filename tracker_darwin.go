package jasper

// TODO

type darwinProcessTracker struct {
	processTrackerBase
}

func NewProcessTracker(name string) (ProcessTracker, error) {
	return &darwinProcessTracker{processTrackerBase{Name: name}}, nil
}
