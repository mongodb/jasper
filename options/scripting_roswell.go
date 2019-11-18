package options

import (
	"crypto/sha1"
	"fmt"
	"io"
	"sort"
	"time"
)

type ScriptingRoswell struct {
	Path           string
	Systems        []string
	Lisp           string
	CachedDuration time.Duration
	cachedAt       time.Time
	cachedHash     string
}

func (opts *ScriptingRoswell) Type() string        { return "roswell" }
func (opts *ScriptingRoswell) Interpreter() string { return "ros" }

func (opts *ScriptingRoswell) ID() string {
	if opts.cachedHash != "" && time.Since(opts.cachedAt) < opts.CachedDuration {
		return opts.cachedHash
	}

	hash := sha1.New()

	_, _ = io.WriteString(hash, opts.Lisp)
	_, _ = io.WriteString(hash, opts.Path)

	sort.Strings(opts.Systems)
	for _, sys := range opts.Systems {
		_, _ = io.WriteString(hash, sys)
	}

	opts.cachedHash = fmt.Sprintf("%x", hash.Sum(nil))
	opts.cachedAt = time.Now()

	return opts.cachedHash
}
