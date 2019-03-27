package jasper

// TODO

type darwinProcessTracker struct {
	processTrackerBase
}

func newProcessTracker(name string) (processTracker, error) {
	return &darwinProcessTracker{}, nil
}
