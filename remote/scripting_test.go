package remote

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mongodb/jasper/scripting"
	"github.com/mongodb/jasper/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type scriptingTestCase struct {
	Name string
	Case func(ctx context.Context, t *testing.T, client Manager, tmpDir string)
}

func TestScripting(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpClient := testutil.GetHTTPClient()
	defer testutil.PutHTTPClient(httpClient)

	for managerName, makeManager := range remoteManagerTestCases(httpClient) {
		t.Run(managerName, func(t *testing.T) {
			for _, test := range []scriptingTestCase{
				{
					Name: "SetupSucceeds",
					Case: func(ctx context.Context, t *testing.T, client Manager, tmpDir string) {
						harness := createTestScriptingHarness(ctx, t, client, tmpDir)
						assert.NoError(t, harness.Setup(ctx))
					},
				},
				{
					Name: "SetupFails",
					Case: func(ctx context.Context, t *testing.T, client Manager, tmpDir string) {
						harness := createTestScriptingHarness(ctx, t, client, tmpDir)
						require.NoError(t, os.Chmod(tmpDir, 0111))
						assert.Error(t, harness.Setup(ctx))
						require.NoError(t, os.Chmod(tmpDir, 0777))
					},
				},
				{
					Name: "CleanupSucceeds",
					Case: func(ctx context.Context, t *testing.T, client Manager, tmpDir string) {
						harness := createTestScriptingHarness(ctx, t, client, tmpDir)
						assert.NoError(t, harness.Cleanup(ctx))
					},
				},
				{
					Name: "CleanupFails",
					Case: func(ctx context.Context, t *testing.T, client Manager, tmpDir string) {
						harness := createTestScriptingHarness(ctx, t, client, tmpDir)
						require.NoError(t, harness.Setup(ctx))
						require.NoError(t, os.Chmod(tmpDir, 0111))
						assert.Error(t, harness.Cleanup(ctx))
						require.NoError(t, os.Chmod(tmpDir, 0777))
					},
				},
				{
					Name: "RunSucceeds",
					Case: func(ctx context.Context, t *testing.T, client Manager, tmpDir string) {
						harness := createTestScriptingHarness(ctx, t, client, tmpDir)
						tmpFile := filepath.Join(tmpDir, "fake_script.go")
						require.NoError(t, ioutil.WriteFile(tmpFile, []byte(testutil.GolangMainSuccess()), 0755))
						assert.NoError(t, harness.Run(ctx, []string{tmpFile}))
					},
				},
				{
					Name: "RunFails",
					Case: func(ctx context.Context, t *testing.T, client Manager, tmpDir string) {
						harness := createTestScriptingHarness(ctx, t, client, tmpDir)

						tmpFile := filepath.Join(tmpDir, "fake_script.go")
						require.NoError(t, ioutil.WriteFile(tmpFile, []byte(testutil.GolangMainFail()), 0755))
						assert.Error(t, harness.Run(ctx, []string{tmpFile}))
					},
				},
				{
					Name: "RunScriptSucceeds",
					Case: func(ctx context.Context, t *testing.T, client Manager, tmpDir string) {
						harness := createTestScriptingHarness(ctx, t, client, tmpDir)
						assert.NoError(t, harness.RunScript(ctx, testutil.GolangMainSuccess()))
					},
				},
				{
					Name: "RunScriptFails",
					Case: func(ctx context.Context, t *testing.T, client Manager, tmpDir string) {

						harness := createTestScriptingHarness(ctx, t, client, tmpDir)
						require.Error(t, harness.RunScript(ctx, testutil.GolangMainFail()))
					},
				},
				{
					Name: "BuildSucceeds",
					Case: func(ctx context.Context, t *testing.T, client Manager, tmpDir string) {
						harness := createTestScriptingHarness(ctx, t, client, tmpDir)

						tmpFile := filepath.Join(tmpDir, "fake_script.go")
						require.NoError(t, ioutil.WriteFile(tmpFile, []byte(testutil.GolangMainSuccess()), 0755))
						buildFile := filepath.Join(tmpDir, "fake_script")
						_, err := harness.Build(ctx, tmpDir, []string{
							"-o",
							buildFile,
							tmpFile,
						})
						require.NoError(t, err)
						_, err = os.Stat(buildFile)
						require.NoError(t, err)
					},
				},
				{
					Name: "BuildFails",
					Case: func(ctx context.Context, t *testing.T, client Manager, tmpDir string) {
						harness := createTestScriptingHarness(ctx, t, client, tmpDir)

						tmpFile := filepath.Join(tmpDir, "fake_script.go")
						require.NoError(t, ioutil.WriteFile(tmpFile, []byte(`package main; func main() { "bad syntax" }`), 0755))
						buildFile := filepath.Join(tmpDir, "fake_script")
						_, err := harness.Build(ctx, tmpDir, []string{
							"-o",
							buildFile,
							tmpFile,
						})
						require.Error(t, err)
						_, err = os.Stat(buildFile)
						assert.Error(t, err)
					},
				},
				{
					Name: "TestSucceeds",
					Case: func(ctx context.Context, t *testing.T, client Manager, tmpDir string) {
						harness := createTestScriptingHarness(ctx, t, client, tmpDir)

						tmpFile := filepath.Join(tmpDir, "fake_script_test.go")
						require.NoError(t, ioutil.WriteFile(tmpFile, []byte(testutil.GolangTestSuccess()), 0755))
						results, err := harness.Test(ctx, tmpDir, scripting.TestOptions{Name: "dummy"})
						require.NoError(t, err)
						require.Len(t, results, 1)
						assert.Equal(t, scripting.TestOutcomeSuccess, results[0].Outcome)
					},
				},
				{
					Name: "TestFails",
					Case: func(ctx context.Context, t *testing.T, client Manager, tmpDir string) {
						harness := createTestScriptingHarness(ctx, t, client, tmpDir)

						tmpFile := filepath.Join(tmpDir, "fake_script_test.go")
						require.NoError(t, ioutil.WriteFile(tmpFile, []byte(testutil.GolangTestFail()), 0755))
						results, err := harness.Test(ctx, tmpDir, scripting.TestOptions{Name: "dummy"})
						assert.Error(t, err)
						require.Len(t, results, 1)
						assert.Equal(t, scripting.TestOutcomeFailure, results[0].Outcome)
					},
				},
			} {
				t.Run(test.Name, func(t *testing.T) {
					tctx, cancel := context.WithTimeout(ctx, testutil.RPCTestTimeout)
					defer cancel()
					client := makeManager(tctx, t)
					defer func() {
						assert.NoError(t, client.CloseConnection())
					}()
					tmpDir, err := ioutil.TempDir(testutil.BuildDirectory(), "scripting_tests")
					require.NoError(t, err)
					defer func() {
						assert.NoError(t, os.RemoveAll(tmpDir))
					}()
					test.Case(tctx, t, client, tmpDir)
				})

			}
		})
	}
}

func createTestScriptingHarness(ctx context.Context, t *testing.T, client Manager, dir string) scripting.Harness {
	opts := testutil.ValidGolangScriptingHarnessOptions(dir)
	sh, err := client.CreateScripting(ctx, opts)
	require.NoError(t, err)
	return sh
}
