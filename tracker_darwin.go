package jasper

// TODO

type darwinProcessTracker struct {
	processTrackerBase
}

func NewProcessTracker(name string) (processTracker, error) {
	return &darwinProcessTracker{processTrackerBase{Name: name}}, nil
}
