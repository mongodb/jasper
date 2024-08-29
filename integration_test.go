package jasper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/evergreen-ci/bond"
	"github.com/evergreen-ci/bond/recall"
	"github.com/mongodb/grip"
	"github.com/mongodb/jasper/options"
	"github.com/mongodb/jasper/testutil"
	testutiloptions "github.com/mongodb/jasper/testutil/options"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// downloadMongoDB downloads MongoDB for testing. It returns the directory
// containing the downloaded MongoDB files and the mongod executable itself.
func downloadMongoDB(t *testing.T) (string, string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := testutiloptions.ValidMongoDBBuildOptions()
	releases := []string{"7.0-stable"}
	dir, err := os.MkdirTemp("", "mongodb")
	require.NoError(t, err)
	require.NoError(t, recall.DownloadReleases(releases, dir, opts))

	catalog, err := bond.NewCatalog(ctx, dir)
	require.NoError(t, err)

	path, err := catalog.Get("7.0-current", string(opts.Edition), opts.Target, string(opts.Arch), false)
	require.NoError(t, err)

	var mongodPath string
	if runtime.GOOS == "windows" {
		mongodPath = filepath.Join(path, "bin", "mongod.exe")
	} else {
		mongodPath = filepath.Join(path, "bin", "mongod")
	}

	_, err = os.Stat(mongodPath)
	require.NoError(t, err)

	return dir, mongodPath
}

func setupMongods(numProcs int, mongodPath string) ([]options.Create, []string, error) {
	dbPaths := make([]string, numProcs)
	optslist := make([]options.Create, numProcs)
	for i := 0; i < numProcs; i++ {
		procName := "mongod"
		port := testutil.GetPortNumber()

		dbPath, err := os.MkdirTemp(testutil.BuildDirectory(), procName)
		if err != nil {
			return nil, nil, err
		}
		dbPaths[i] = dbPath

		opts := options.Create{
			Args: []string{mongodPath, "--port", fmt.Sprintf("%d", port), "--dbpath", dbPath},
		}
		optslist[i] = opts
	}

	return optslist, dbPaths, nil
}

func removeDBPaths(dbPaths []string) {
	for _, dbPath := range dbPaths {
		grip.Error(os.RemoveAll(dbPath))
	}
}

func TestMongod(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping mongod tests in short mode")
	}
	if runtime.GOOS != "linux" {
		t.Skip("skipping mongod tests on non-Linux platforms because they're not important and mongod is heavily platform-dependent")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dir, mongodPath := downloadMongoDB(t)
	defer os.RemoveAll(dir)

	for name, makeProc := range map[string]ProcessConstructor{
		"BasicProcess":    newBasicProcess,
		"BlockingProcess": newBlockingProcess,
	} {
		t.Run(name, func(t *testing.T) {
			for _, test := range []struct {
				id          string
				numProcs    int
				signal      syscall.Signal
				sleep       time.Duration
				expectError bool
			}{
				{
					id:          "WithSingleMongodAndSIGKILL",
					numProcs:    1,
					signal:      syscall.SIGKILL,
					sleep:       0,
					expectError: true,
				},
				{
					id:          "With10MongodsAndSIGKILL",
					numProcs:    10,
					signal:      syscall.SIGKILL,
					sleep:       2 * time.Second,
					expectError: true,
				},
				{
					id:          "With30MongodsAndSIGKILL",
					numProcs:    30,
					signal:      syscall.SIGKILL,
					sleep:       3 * time.Second,
					expectError: true,
				},
			} {
				t.Run(test.id, func(t *testing.T) {
					optslist, dbPaths, err := setupMongods(test.numProcs, mongodPath)
					defer removeDBPaths(dbPaths)
					require.NoError(t, err)

					// Spawn concurrent mongods
					procs := make([]Process, test.numProcs)
					for i, opts := range optslist {
						proc, err := makeProc(ctx, &opts)
						require.NoError(t, err)
						assert.True(t, proc.Running(ctx))
						procs[i] = proc
					}

					waitError := make(chan error, test.numProcs)
					for _, proc := range procs {
						go func(proc Process) {
							_, err := proc.Wait(ctx)
							select {
							case waitError <- err:
							case <-ctx.Done():
							}
						}(proc)
					}

					// Signal to stop mongods
					time.Sleep(test.sleep)
					for _, proc := range procs {
						assert.NoError(t, proc.Signal(ctx, test.signal))
					}

					// Wait for all processes to exit after receiving the
					// signal.
					for i := 0; i < test.numProcs; i++ {
						select {
						case err := <-waitError:
							if test.expectError {
								assert.Error(t, err)
							} else {
								assert.NoError(t, err)
							}
						case <-ctx.Done():
							require.FailNow(t, "context errored before processes exited: %s", ctx.Err())
						}
					}

					// Check that the processes have all noted a unsuccessful run
					for _, proc := range procs {
						assert.Equal(t, !test.expectError, proc.Info(ctx).Successful)
					}
				})
			}
		})
	}
}
