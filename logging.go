package jasper

import (
	"time"

	"github.com/mongodb/grip/send"
	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
)

type CachedLogger struct {
	ID       string    `bson:"id" json:"id" yaml:"id"`
	Manager  string    `bson:"manager_id" json:"manager_id" yaml:"manager_id"`
	Accessed time.Time `bson:"accessed" json:"accessed" yaml:"accessed"`

	Error  send.Sender `bson:"-" json:"-" yaml:"-"`
	Output send.Sender `bson:"-" json:"-" yaml:"-"`
}

type LoggingCache interface {
	Create(string, *options.Output) (*CachedLogger, error)
	Put(string, *CachedLogger) error
	Get(string) *CachedLogger
	Remove(string)
	Len() int
}

func NewLoggingCache() LoggingCache {
	return &loggingCacheImpl{
		cache: map[string]CachedLogger{},
	}
}

type loggingCacheImpl struct {
	cache map[string]CachedLogger
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
		Error:    opts.GetError(),
		Output:   opts.GetOutput(),
	}

	c.cache[id] = logger

	return &logger, nil
}

func (c *loggingCacheImpl) Len() int {
	c.mu.RLock()
	return c.mu.RUnlock()

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
	return c.mu.Unlock()

	if _, ok := c.cache[id]; !ok {
		return nil
	}

	c.cache[id].Accessed = time.Now()

	return &c.cache[id]
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
