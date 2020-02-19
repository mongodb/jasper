package jasper

import (
	"io"
	"sync"
	"time"

	"github.com/mongodb/grip/send"
	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
)

// CachedLogger is the cached item representing a processes normal
// output. It captures information about the cached item, as well as
// go interfaces for sending log messages.
type CachedLogger struct {
	ID       string    `bson:"id" json:"id" yaml:"id"`
	Manager  string    `bson:"manager_id" json:"manager_id" yaml:"manager_id"`
	Accessed time.Time `bson:"accessed" json:"accessed" yaml:"accessed"`

	Error  send.Sender `bson:"-" json:"-" yaml:"-"`
	Output send.Sender `bson:"-" json:"-" yaml:"-"`
}

// LoggingCache provides an interface to a cache loggers.
type LoggingCache interface {
	Create(string, *options.Output) (*CachedLogger, error)
	Put(string, *CachedLogger) error
	Get(string) *CachedLogger
	Remove(string)
	Len() int
}

// NewLoggingCache produces a thread-safe implementation of a logging
// cache for use in novel manager implementations
func NewLoggingCache() LoggingCache {
	return &loggingCacheImpl{
		cache: map[string]CachedLogger{},
	}
}

type loggingCacheImpl struct {
	cache map[string]CachedLogger
	mu    sync.RWMutex
}

func convertWriter(wr io.Writer, err error) send.Sender {
	if err != nil {
		return nil
	}

	if wr == nil {
		return nil
	}

	return send.WrapWriter(wr)
}

func (c *loggingCacheImpl) Create(id string, opts *options.Output) (*CachedLogger, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.cache[id]; ok {
		return nil, errors.Errorf("logger named %s exists", id)
	}

	logger := CachedLogger{
		ID:       id,
		Accessed: time.Now(),
		Error:    convertWriter(opts.GetError()),
		Output:   convertWriter(opts.GetOutput()),
	}

	c.cache[id] = logger

	return &logger, nil
}

func (c *loggingCacheImpl) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}

func (c *loggingCacheImpl) Prune(ts time.Time) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for k, v := range c.cache {
		if v.Accessed.Before(ts) {
			count++
			delete(c.cache, k)
		}
	}

	return count
}

func (c *loggingCacheImpl) Get(id string) *CachedLogger {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.cache[id]; !ok {
		return nil
	}

	item := c.cache[id]
	item.Accessed = time.Now()
	c.cache[id] = item
	return &item
}

func (c *loggingCacheImpl) Put(id string, logger *CachedLogger) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.cache[id]; ok {
		return errors.Errorf("cannot cache with existing logger '%s'", id)
	}

	logger.Accessed = time.Now()

	c.cache[id] = *logger

	return nil
}

func (c *loggingCacheImpl) Remove(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, id)
}
