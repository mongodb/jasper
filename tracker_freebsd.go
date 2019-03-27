package jasper

// TODO: implement

type freebsdProcessTracker struct {
	processTrackerBase
}

func newProcessTracker(name string) (processTracker, error) {
	return &freebsdProcessTracker{}, nil
}
