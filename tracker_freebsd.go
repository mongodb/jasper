package jasper

// TODO: implement

type freebsdProcessTracker struct{}

func newProcessTracker(name string) (processTracker, error) {
	return &freebsdProcessTracker{}, nil
}

func (_ *freebsdProcessTracker) add(pid int) error {
	return nil
}

func (_ *freebsdProcessTracker) cleanup() error {
	return nil
}
