package options

import (
	"encoding/json"
	"sync"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

// LoggerRegistry is an interface that stores reusable logger factories.
type LoggerRegistry interface {
	Register(string, LoggerFactory)
	Check(string) bool
	Names() []string
	Resolve(string) (LoggerFactory, bool)
}

type basicRegistry struct {
	mu        sync.RWMutex
	factories map[string]LoggerFactory
}

func (r *basicRegistry) Register(name string, factory LoggerFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.factories[name] = factory
}

func (r *basicRegistry) Check(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.factories[name]
	return ok
}

func (r *basicRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := []string{}
	for name := range r.factories {
		names = append(names, name)
	}

	return names
}

func (r *basicRegistry) Resolve(name string) (LoggerFactory, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, ok := r.factories[name]
	return factory, ok
}

// LogConfigFormat describes the format of the log configuration.
type LogConfigFormat string

const (
	LogConfigFormatBSON LogConfigFormat = "BSON"
	LogConfigFormatJSON LogConfigFormat = "JSON"
)

// Unmarshal unmarshals the data into out based on the log config format.
func (f LogConfigFormat) Unmarshal(data []byte, out interface{}) error {
	switch f {
	case LogConfigFormatBSON:
		if err := bson.Unmarshal(data, out); err != nil {
			return errors.Wrapf(err, "could not render '%s' input into '%s'", data, out)

		}

		return nil
	case LogConfigFormatJSON:
		if err := json.Unmarshal(data, out); err != nil {
			return errors.Wrapf(err, "could not render '%s' input into '%s'", data, out)

		}

		return nil
	default:
		return errors.Errorf("unsupported format '%s'", f)
	}
}

// LoggerFactory produces a Logger interface backed by a grip logger.
type LoggerFactory func([]byte, LogConfigFormat) (Logger, error)
