package options

import (
	"crypto/sha1"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"time"

	uuid "github.com/satori/go.uuid"
)

type ScriptingRoswell struct {
	Path           string
	Systems        []string
	Lisp           string
	CachedDuration time.Duration
	Environment    map[string]string
	Output         Output

	cachedAt   time.Time
	cachedHash string
}

func (opts *ScriptingRoswell) Validate() error {
	if opts.CachedDuration == 0 {
		opts.CachedDuration = 10 * time.Minute
	}

	if opts.Path == "" {
		opts.Path = filepath.Join("roswell", uuid.Must(uuid.NewV4()).String())
	}

	if opts.Lisp == "" {
		opts.Lisp = "sbcl-bin"
	}

	return nil
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
