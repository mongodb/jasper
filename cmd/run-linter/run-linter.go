package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/shlex"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

type result struct {
	name     string
	cmd      string
	passed   bool
	duration time.Duration
	output   []string
}

// String prints the results of a linter run in gotest format.
func (r *result) String() string {
	buf := &bytes.Buffer{}

	fmt.Fprintln(buf, "=== RUN", r.name)
	if r.passed {
		fmt.Fprintf(buf, "--- PASS: %s (%s)", r.name, r.duration)
	} else {
		fmt.Fprintf(buf, strings.Join(r.output, "\n"))
		fmt.Fprintf(buf, "--- FAIL: %s (%s)", r.name, r.duration)
	}

	return buf.String()
}

// fixup goes through the output and improves the output generated by
// specific linters so that all output includes the relative path to the
// error, instead of mixing relative and absolute paths.
func (r *result) fixup(dirname string) {
	for idx, ln := range r.output {
		if strings.HasPrefix(ln, dirname) {
			r.output[idx] = ln[len(dirname)+1:]
		}
	}
}

// runs the golangci-lint on a list of packages; integrating with the "make lint" target.
func main() {
	var (
		lintArgs          string
		lintBin           string
		customLintersFlag string
		customLinters     []string
		packageList       string
		output            string
		packages          []string
		results           []*result
		hasFailingTest    bool

		gopath = os.Getenv("GOPATH")
	)

	gopath, _ = filepath.Abs(gopath)

	flag.StringVar(&lintArgs, "lintArgs", "", "args to pass to golangci-lint")
	flag.StringVar(&lintBin, "lintBin", filepath.Join(gopath, "bin", "golangci-lint"), "path to golangci-lint")
	flag.StringVar(&packageList, "packages", "", "list of space separated packages")
	flag.StringVar(&customLintersFlag, "customLinters", "", "list of comma-separated custom linter commands")
	flag.StringVar(&output, "output", "", "output file for to write results.")
	flag.Parse()

	if len(customLintersFlag) != 0 {
		customLinters = strings.Split(customLintersFlag, ",")
	}
	packages = strings.Split(strings.Replace(packageList, "-", "/", -1), " ")
	dirname, _ := os.Getwd()
	cwd := filepath.Base(dirname)

	for _, pkg := range packages {
		pkgDir := "./"
		if cwd != pkg {
			pkgDir += pkg
		}
		splitLintArgs, err := shlex.Split(lintArgs)
		if err != nil {
			grip.Error(errors.Wrapf(err, "splitting lint args '%s'", lintArgs))
			os.Exit(1)
		}
		args := []string{lintBin, "run"}
		args = append(args, splitLintArgs...)
		args = append(args, pkgDir)

		startAt := time.Now()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Env = os.Environ()
		cmd.Dir = dirname
		out, err := cmd.CombinedOutput()
		r := &result{
			cmd:      strings.Join(args, " "),
			name:     "lint-" + strings.Replace(pkg, "/", "-", -1),
			passed:   err == nil,
			duration: time.Since(startAt),
			output:   strings.Split(string(out), "\n"),
		}

		for _, linter := range customLinters {
			customLinterStart := time.Now()
			cmd := exec.Command(linter, pkgDir)
			cmd.Env = os.Environ()
			cmd.Dir = dirname
			out, err := cmd.CombinedOutput()
			r.passed = r.passed && err == nil
			r.duration += time.Since(customLinterStart)
			r.output = append(r.output, strings.Split(string(out), "\n")...)
		}
		r.fixup(dirname)

		if !r.passed {
			hasFailingTest = true
		}

		results = append(results, r)
		fmt.Println(r)
	}

	if output != "" {
		f, err := os.Create(output)
		if err != nil {
			os.Exit(1)
		}
		defer func() {
			if err != f.Close() {
				panic(err)
			}
		}()

		for _, r := range results {
			if _, err = f.WriteString(r.String() + "\n"); err != nil {
				grip.Error(err)
				os.Exit(1)
			}
		}
	}

	if hasFailingTest {
		os.Exit(1)
	}
}
