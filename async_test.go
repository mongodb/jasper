package jasper

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/queue"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tychoish/bond"
	"github.com/tychoish/lru"
)

func validMongoDBDownloadOptions() MongoDBDownloadOptions {
	target := runtime.GOOS
	if target == "darwin" {
		target = "osx"
	}
	edition := "enterprise"
	if target == "linux" {
		edition = "base"
	}
	return MongoDBDownloadOptions{
		BuildOpts: bond.BuildOptions{
			Target:  target,
			Arch:    bond.MongoDBArch("x86_64"),
			Edition: bond.MongoDBEdition(edition),
			Debug:   false,
		},
		Path:     "build",
		Releases: []string{"4.0-current"},
	}
}

func TestSetupDownloadMongoDBReleasesWithBadPath(t *testing.T) {
	opts := validMongoDBDownloadOptions()
	opts.Path = "async_test.go"
	assert.Error(t, SetupDownloadMongoDBReleases(context.Background(), lru.NewCache(), opts))
}

func TestSetupDownloadMongoDBReleasesWithBadArtifactsFeed(t *testing.T) {
	opts := validMongoDBDownloadOptions()
	opts.Path = filepath.Join("build", "full.json")
	assert.Error(t, SetupDownloadMongoDBReleases(context.Background(), lru.NewCache(), opts))
}

func TestCreateValidDownloadJobs(t *testing.T) {
	dir, err := ioutil.TempDir("build", "out")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	urls := make(chan string)

	go func() {
		urls <- "https://example.com/foo"
		close(urls)
	}()
	jobs, errs := createDownloadJobs(dir, urls)

	count := 0
	for job := range jobs {
		count++
		assert.Equal(t, 1, count)
		assert.NotNil(t, job)
	}
	assert.NoError(t, aggregateErrors(errs))
}

func TestCreateInvalidDownloadJobs(t *testing.T) {
	dir := "async_test.go"
	urls := make(chan string)

	go func() {
		urls <- "https://example.com"
		close(urls)
	}()
	jobs, errs := createDownloadJobs(dir, urls)

	for range jobs {
		assert.Fail(t, "should not create job for bad url")
	}
	assert.Error(t, aggregateErrors(errs))
}

func TestSetupDownloadJobsAsync(t *testing.T) {
	dir, err := ioutil.TempDir("build", "out")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	ctx := context.Background()
	urls := make(chan string)
	go func() {
		urls <- "https://example.com"
		close(urls)
	}()
	jobs, errs := createDownloadJobs(dir, urls)
	checkFileName := func(fileName string) error {
		if fileName != "example.com" {
			return errors.New("file name did not match expected")
		}
		return nil
	}

	assert.NoError(t, setupDownloadJobsAsync(ctx, jobs, processDownloadJobs(context.Background(), checkFileName)))
	assert.NoError(t, aggregateErrors(errs))
}

func TestSetupDownloadReleasesFailsForInvalidOptions(t *testing.T) {
	ctx := context.Background()
	opts := MongoDBDownloadOptions{}
	err := SetupDownloadMongoDBReleases(ctx, lru.NewCache(), opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid build options")
}

func TestProcessDownloadJobs(t *testing.T) {
	dir, err := ioutil.TempDir("build", "mongodb")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	absDir, err := filepath.Abs(dir)
	require.NoError(t, err)
	ctx := context.Background()
	cache := lru.NewCache()
	downloadOpts := validMongoDBDownloadOptions()
	opts := downloadOpts.BuildOpts
	releases := downloadOpts.Releases

	feed, err := bond.GetArtifactsFeed(dir)
	require.NoError(t, err)

	urls, errs1 := feed.GetArchives(releases, opts)
	jobs, errs2 := createDownloadJobs(dir, urls)

	q := queue.NewLocalUnordered(runtime.NumCPU())
	require.NoError(t, q.Start(ctx))
	require.NoError(t, amboy.PopulateQueue(ctx, q, jobs))
	require.NoError(t, aggregateErrors(errs1, errs2))

	_ = amboy.WaitCtxInterval(ctx, q, 100*time.Millisecond)
	require.NoError(t, amboy.ResolveErrors(ctx, q))

	assert.NoError(t, processDownloadJobs(ctx, addMongoDBFilesToCache(cache, absDir))(q))

	downloadedFiles := []string{}
	filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)
		if !info.IsDir() && info.Name() != "full.json" {
			downloadedFiles = append(downloadedFiles, path)
		}
		return nil
	})

	assert.NotEqual(t, 0, cache.Size())
	assert.Equal(t, len(downloadedFiles), cache.Count())

	for _, fileName := range downloadedFiles {
		fObj, err := cache.Get(fileName)
		assert.NoError(t, err)
		assert.NotNil(t, fObj)
	}
}

func TestAddMongoDBFilesToCacheWithBadPath(t *testing.T) {
	absPath, err := filepath.Abs("build")
	require.NoError(t, err)
	assert.Error(t, addMongoDBFilesToCache(lru.NewCache(), absPath)("foo.txt"))
}

func TestDoDownloadWithValidInfo(t *testing.T) {
	file, err := ioutil.TempFile("build", "out.txt")
	require.NoError(t, err)
	defer os.Remove(file.Name())

	info := DownloadInfo{
		URL:  "https://example.com",
		Path: file.Name(),
	}
	req, err := http.NewRequest(http.MethodGet, info.URL, nil)
	require.NoError(t, err)

	assert.NoError(t, DoDownload(req, info, http.Client{}))
	fileInfo, err := file.Stat()
	require.NoError(t, err)
	assert.NotZero(t, fileInfo.Size())
}

func TestDoDownloadWithNonexistentURL(t *testing.T) {
	file, err := ioutil.TempFile("build", "out.txt")
	require.NoError(t, err)
	defer os.Remove(file.Name())

	info := DownloadInfo{
		URL:  "https://example.com/foo",
		Path: file.Name(),
	}
	req, err := http.NewRequest(http.MethodGet, info.URL, nil)
	require.NoError(t, err)

	assert.Error(t, DoDownload(req, info, http.Client{}))
}
