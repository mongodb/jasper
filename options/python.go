package options

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ScriptingEnvironment defines the interface for all types that
// define a scripting environment.
type ScriptingEnvironment interface {
	ID() string
	Type() string
	Interpreter() string
}

// ScriptingPython defines the configuration of a python environment.
type ScriptingPython struct {
	VirtualEnvPath       string
	RequirementsFilePath string
	HostPythonInterpeter string
	Packages             []string
	CachedDuration       time.Duration
	LegacyPython         bool

	requirementsMTime time.Time
	cachedAt          time.Time
	requrementsHash   string
	cachedHash        string
}

func NewPythonScriptingEnvironmnet(path, reqtxt string, packages ...string) ScriptingEnvironment {
	return &ScriptingPython{
		CachedDuration:       time.Hour,
		HostPythonInterpeter: "python3",
		Packages:             packages,
		VirtualEnvPath:       path,
		RequirementsFilePath: reqtxt,
	}
}

func (opts *ScriptingPython) Interpreter() string {
	return filepath.Join(opts.VirtualEnvPath, "bin", "python")
}
func (opts *ScriptingPython) Type() string { return "python" }

func (opts *ScriptingPython) ID() string {
	if opts.cachedHash != "" && time.Since(opts.cachedAt) < opts.CachedDuration {
		return opts.cachedHash
	}
	hash := sha1.New()

	sort.Strings(opts.Packages)
	for _, str := range opts.Packages {
		_, _ = io.WriteString(hash, str)
	}

	_, _ = io.WriteString(hash, opts.VirtualEnvPath)
	_, _ = io.WriteString(hash, opts.HostPythonInterpeter)

	if opts.requrementsHash == "" {
		stat, err := os.Stat(opts.RequirementsFilePath)
		if !os.IsNotExist(err) && (stat.ModTime() != opts.requirementsMTime) {
			reqData, err := ioutil.ReadFile(opts.RequirementsFilePath)
			if err == nil {
				reqHash := sha1.New()
				_, _ = reqHash.Write(reqData)
				opts.requrementsHash = fmt.Sprintf("%x", reqHash.Sum(nil))
			}
		}
	}

	_, _ = io.WriteString(hash, opts.requrementsHash)
	opts.cachedHash = fmt.Sprintf("%x", hash.Sum(nil))
	opts.cachedAt = time.Now()
	return opts.cachedHash
}
