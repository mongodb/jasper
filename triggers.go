package jasper

type ProcessTrigger func(ProcessInfo)

type ProcessTriggerSequence []ProcessTrigger

func (s ProcessTriggerSequence) Run(info ProcessInfo) {
	for _, trigger := range s {
		trigger(info)
	}
}
