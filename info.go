package jasper

type ProcessInfo struct {
	ID         string
	Host       string
	PID        int
	IsRunning  bool
	Successful bool
	Complete   bool
	Timeout    bool
	Options    CreateOptions
}
