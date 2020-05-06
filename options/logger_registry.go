package options

import (
	"encoding/json"
	"sync"

	"github.com/mongodb/grip/send"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

// LoggerRegistry is an interface that stores reusable logger factories.
type LoggerRegistry interface {
	Register(LoggerProducerFactory)
	Check(string) bool
	Names() []string
	Resolve(string) (LoggerProducerFactory, bool)
}

type basicLoggerRegistry struct {
	mu        sync.RWMutex
	factories map[string]LoggerProducerFactory
}

// NewBasicLoggerRegsitry returns a new LoggerRegistry backed by the
// basicLoggerRegistry implementation.
func NewBasicLoggerRegistry() LoggerRegistry {
	return &basicLoggerRegistry{
		factories: map[string]LoggerProducerFactory{},
	}
}

func (r *basicLoggerRegistry) Register(factory LoggerProducerFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.factories[factory().Type()] = factory
}

func (r *basicLoggerRegistry) Check(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.factories[name]
	return ok
}

func (r *basicLoggerRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := []string{}
	for name := range r.factories {
		names = append(names, name)
	}

	return names
}

func (r *basicLoggerRegistry) Resolve(name string) (LoggerProducerFactory, bool) {
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

// Validate ensures that LogConfigFormat is valid.
func (f LogConfigFormat) Validate() error {
	switch f {
	case LogConfigFormatBSON, LogConfigFormatJSON:
		return nil
	default:
		return errors.New("unknown log config format")
	}
}

func (f LogConfigFormat) unmarshal(data []byte, out LoggerProducer) error {
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

// LoggerProducer produces a Logger interface backed by a grip logger.
type LoggerProducer interface {
	Type() string
	Configure() (send.Sender, error)
}

// LoggerProducerFactory creates a new instance of a LoggerProducer implementation.
type LoggerProducerFactory func() LoggerProducer
