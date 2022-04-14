package jasper

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/evergreen-ci/bond"
	"github.com/evergreen-ci/bond/recall"
	"github.com/evergreen-ci/lru"
	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/queue"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/recovery"
	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
)

func makeEnclosingDirectories(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(path, os.ModeDir|os.ModePerm); err != nil {
			return err
		}
	} else if !info.IsDir() {
		return errors.Errorf("path '%s' already exists and is not a directory", path)
	}
	return nil
}

// SetupDownloadMongoDBReleases performs necessary setup to download MongoDB with the given options.
func SetupDownloadMongoDBReleases(ctx context.Context, cache *lru.Cache, opts options.MongoDBDownload) error {
	if err := makeEnclosingDirectories(opts.Path); err != nil {
		return errors.Wrap(err, "creating enclosing directories")
	}

	feed, err := bond.GetArtifactsFeed(ctx, opts.Path)
	if err != nil {
		return errors.Wrap(err, "making artifacts feed")
	}

	catcher := grip.NewBasicCatcher()
	urls, errs := feed.GetArchives(opts.Releases, opts.BuildOpts)
	jobs := createDownloadJobs(opts.Path, urls, catcher)

	if err = setupDownloadJobsAsync(ctx, jobs, processDownloadJobs(ctx, addMongoDBFilesToCache(cache, opts.Path))); err != nil {
		catcher.Wrap(err, "starting download jobs")
	}

	for err = range errs {
		catcher.Wrap(err, "initializing download jobs")
	}

	return catcher.Resolve()
}

func createDownloadJobs(path string, urls <-chan string, catcher grip.Catcher) <-chan amboy.Job {
	output := make(chan amboy.Job)

	go func() {
		defer recovery.LogStackTraceAndContinue("download generator")

		for url := range urls {
			j, err := recall.NewDownloadJob(url, path, true)
			if err != nil {
				catcher.Wrapf(err, "creating download job for URL '%s'", url)
				continue
			}

			output <- j
		}
		close(output)
	}()

	return output
}

func processDownloadJobs(ctx context.Context, processFile func(string) error) func(amboy.Queue) error {
	return func(q amboy.Queue) error {
		grip.Infof("waiting for %d download jobs to complete", q.Stats(ctx).Total)
		if !amboy.WaitInterval(ctx, q, time.Second) {
			return errors.New("download jobs queue timed out")
		}
		grip.Info("all download tasks complete, processing errors now")

		if err := amboy.ResolveErrors(ctx, q); err != nil {
			return errors.Wrap(err, "executing download jobs")
		}

		catcher := grip.NewBasicCatcher()
		for job := range q.Results(ctx) {
			catcher.Add(job.Error())
			downloadJob, ok := job.(*recall.DownloadFileJob)
			if !ok {
				catcher.Errorf("could not retrieve job '%s' from queue, expected download job but got %T instead", job.ID(), job)
				continue
			}
			file := filepath.Join(downloadJob.Directory, downloadJob.FileName)
			if err := processFile(file); err != nil {
				catcher.Wrapf(err, "processing file '%s'", file)
			}
		}
		return catcher.Resolve()
	}
}

func setupDownloadJobsAsync(ctx context.Context, jobs <-chan amboy.Job, processJobs func(amboy.Queue) error) error {
	q := queue.NewLocalLimitedSize(2, 1048)
	if err := q.Start(ctx); err != nil {
		return errors.Wrap(err, "starting download job queue")
	}

	if err := amboy.PopulateQueue(ctx, q, jobs); err != nil {
		return errors.Wrap(err, "enqueueing download jobs")
	}

	go func() {
		defer recovery.LogStackTraceAndContinue("download job generator")
		if err := processJobs(q); err != nil {
			grip.Error(errors.Wrap(err, "processing jobs"))
		}
	}()

	return nil
}

func addMongoDBFilesToCache(cache *lru.Cache, absRootPath string) func(string) error {
	return func(fileName string) error {
		filePath := filepath.Join(absRootPath, fileName)
		if err := cache.AddFile(filePath); err != nil {
			return errors.Wrapf(err, "adding file '%s' to LRU cache", filePath)
		}

		baseName := filepath.Base(fileName)
		ext := filepath.Ext(baseName)
		dirPath := filepath.Join(absRootPath, strings.TrimSuffix(baseName, ext))

		err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Cache only handles individual files, not directories.
			if !info.IsDir() {
				if err := cache.AddStat(path, info); err != nil {
					return errors.Wrapf(err, "adding file '%s' to LRU cache", path)
				}
			}

			return nil
		})
		return err
	}
}
