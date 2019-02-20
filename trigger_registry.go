package jasper

import (
	"sync"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

type SignalTriggerFactory func() SignalTrigger

type signalTriggerRegistry struct {
	mu             *sync.RWMutex
	signalTriggers map[SignalTriggerID]SignalTriggerFactory
}

var jasperSignalTriggerRegistry *signalTriggerRegistry

func init() {
	jasperSignalTriggerRegistry = newSignalTriggerRegistry()

	signalTriggers := map[SignalTriggerID]SignalTriggerFactory{
		MongodShutdownSignalTrigger: makeMongodShutdownSignalTrigger,
	}

	for id, factory := range signalTriggers {
		grip.EmergencyPanic(RegisterSignalTriggerFactory(id, factory))
	}
}

func newSignalTriggerRegistry() *signalTriggerRegistry {
	return &signalTriggerRegistry{mu: &sync.RWMutex{}, signalTriggers: map[SignalTriggerID]SignalTriggerFactory{}}
}

func RegisterSignalTriggerFactory(id SignalTriggerID, factory SignalTriggerFactory) error {
	return errors.Wrap(jasperSignalTriggerRegistry.registerSignalTriggerFactory(id, factory), "problem registering signal trigger factory")
}

func GetSignalTriggerFactory(id SignalTriggerID) (SignalTriggerFactory, bool) {
	return jasperSignalTriggerRegistry.getSignalTriggerFactory(id)
}

func (r *signalTriggerRegistry) registerSignalTriggerFactory(id SignalTriggerID, factory SignalTriggerFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if string(id) == "" {
		return errors.New("cannot register an empty signal trigger id")
	}

	if _, ok := r.signalTriggers[id]; ok {
		return errors.Errorf("signal trigger '%s' is already registered", string(id))
	}

	if factory == nil {
		return errors.Errorf("cannot register a nil factory for signal trigger id '%s'", string(id))
	}

	r.signalTriggers[id] = factory
	return nil
}

func (r *signalTriggerRegistry) getSignalTriggerFactory(id SignalTriggerID) (SignalTriggerFactory, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	factory, ok := r.signalTriggers[id]
	grip.Debugf("kim: getting signal trigger factory for id %s was successful? %t", id, ok)
	return factory, ok
}
