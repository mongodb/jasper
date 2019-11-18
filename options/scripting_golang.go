package options

import (
	"crypto/sha1"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"time"
)

type ScriptingGolang struct {
	Gopath         string
	Goroot         string
	Packages       []string
	Context        string
	WithUpdate     bool
	CachedDuration time.Duration

	cachedAt   time.Time
	cachedHash string
}

func (opts *ScriptingGolang) Type() string        { return "go" }
func (opts *ScriptingGolang) Interpreter() string { return filepath.Join(opts.Goroot, "bin", "go") }

func (opts *ScriptingGolang) ID() string {
	if opts.cachedHash != "" && time.Since(opts.cachedAt) < opts.CachedDuration {
		return opts.cachedHash
	}
	hash := sha1.New()

	_, _ = io.WriteString(hash, opts.Goroot)
	_, _ = io.WriteString(hash, opts.Gopath)

	sort.Strings(opts.Packages)
	for _, str := range opts.Packages {
		_, _ = io.WriteString(hash, str)
	}

	_, _ = io.WriteString(hash, opts.Context)

	opts.cachedHash = fmt.Sprintf("%x", hash.Sum(nil))
	opts.cachedAt = time.Now()
	return opts.cachedHash
}
