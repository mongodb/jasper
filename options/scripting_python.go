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

// ScriptingPython defines the configuration of a python environment.
type ScriptingPython struct {
	VirtualEnvPath       string        `bson:"virtual_env_path" json:"virtual_env_path" yaml:"virtual_env_path"`
	RequirementsFilePath string        `bson:"requirements_path" json:"requirements_path" yaml:"requirements_path"`
	HostPythonInterpeter string        `bson:"host_python" json:"host_python" yaml:"host_python"`
	Packages             []string      `bson:"packages" json:"packages" yaml:"packages"`
	CachedDuration       time.Duration `bson:"cache_duration" json:"cache_duration" yaml:"cache_duration"`
	LegacyPython         bool          `bson:"legacy_python" json:"legacy_python" yaml:"legacy_python"`

	requirementsMTime time.Time
	cachedAt          time.Time
	requrementsHash   string
	cachedHash        string
}

// NewPythonScriptingEnvironmnet generates a ScriptingEnvironment
// taking the arguments given for later use. Use this function for
// simple cases when you do not need or want to set as many aspects of
// the environment configuration.
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

	_, _ = io.WriteString(hash, opts.HostPythonInterpeter)
	_, _ = io.WriteString(hash, opts.VirtualEnvPath)

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

	sort.Strings(opts.Packages)
	for _, str := range opts.Packages {
		_, _ = io.WriteString(hash, str)
	}

	opts.cachedHash = fmt.Sprintf("%x", hash.Sum(nil))
	opts.cachedAt = time.Now()
	return opts.cachedHash
}
