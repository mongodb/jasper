package jasper

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/evergreen-ci/gimlet"
	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/queue"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/mongodb/grip/send"
	"github.com/pkg/errors"
	"github.com/tychoish/bond"
	"github.com/tychoish/bond/recall"
	"github.com/tychoish/lru"
)

// Service defines a REST service that provides a remote manager, using
// gimlet to publish routes.
type Service struct {
	hostID     string
	manager    Manager
	cache      *lru.Cache
	cacheOpts  CacheOptions
	cacheMutex sync.RWMutex
}

// NewManagerService creates a service object around an existing
// manager. You must access the application and routes via the App()
// method separately. The constructor wraps basic managers with a
// manager implementation that does locking.
func NewManagerService(m Manager) *Service {
	if bpm, ok := m.(*basicProcessManager); ok {
		m = &localProcessManager{manager: bpm}
	}

	return &Service{
		manager: m,
	}
}

const (
	// DefaultCachePruneDelay is the duration between LRU cache prunes.
	DefaultCachePruneDelay = 10 * time.Second
	// DefaultMaxCacheSize is the maximum allowed size of the LRU cache.
	DefaultMaxCacheSize = 1024 * 1024 * 1024
)

// App constructs and returns a gimlet application for this
// service. It attaches no middleware and does not start the service.
func (s *Service) App() *gimlet.APIApp {
	s.hostID, _ = os.Hostname()
	s.cache = lru.NewCache()
	s.cacheMutex.Lock()
	s.cacheOpts.PruneDelay = DefaultCachePruneDelay
	s.cacheOpts.MaxSize = DefaultMaxCacheSize
	s.cacheOpts.Disabled = false
	s.cacheMutex.Unlock()

	app := gimlet.NewApp()

	app.AddRoute("/").Version(1).Get().Handler(s.rootRoute)
	app.AddRoute("/create").Version(1).Post().Handler(s.createProcess)
	app.AddRoute("/list/{filter}").Version(1).Get().Handler(s.listProcesses)
	app.AddRoute("/list/group/{name}").Version(1).Get().Handler(s.listGroupMembers)
	app.AddRoute("/process/{id}").Version(1).Get().Handler(s.getProcess)
	app.AddRoute("/process/{id}/tags").Version(1).Get().Handler(s.getProcessTags)
	app.AddRoute("/process/{id}/tags").Version(1).Delete().Handler(s.deleteProcessTags)
	app.AddRoute("/process/{id}/tags").Version(1).Post().Handler(s.addProcessTag)
	app.AddRoute("/process/{id}/wait").Version(1).Get().Handler(s.waitForProcess)
	app.AddRoute("/process/{id}/metrics").Version(1).Get().Handler(s.processMetrics)
	app.AddRoute("/process/{id}/signal/{signal}").Version(1).Patch().Handler(s.signalProcess)
	app.AddRoute("/close").Version(1).Delete().Handler(s.closeManager)
	app.AddRoute("/process/{id}/logs").Version(1).Get().Handler(s.getLogs)
	app.AddRoute("/configure-cache").Version(1).Post().Handler(s.configureCache)
	app.AddRoute("/download-mongodb").Version(1).Post().Handler(s.downloadMongoDB)

	go s.backgroundPrune()

	return app
}

func (s *Service) backgroundPrune() {
	s.cacheMutex.RLock()
	timer := time.NewTimer(s.cacheOpts.PruneDelay)
	s.cacheMutex.RUnlock()

	for {
		<-timer.C
		s.cacheMutex.RLock()
		if !s.cacheOpts.Disabled {
			if err := s.cache.Prune(s.cacheOpts.MaxSize, nil, false); err != nil {
				grip.Error(errors.Wrap(err, "error during cache pruning"))
			}
		}
		timer.Reset(s.cacheOpts.PruneDelay)
		s.cacheMutex.RUnlock()
	}
}

func writeError(rw http.ResponseWriter, err gimlet.ErrorResponse) {
	gimlet.WriteJSONResponse(rw, err.StatusCode, err)
}

func (s *Service) rootRoute(rw http.ResponseWriter, r *http.Request) {
	gimlet.WriteJSON(rw, struct {
		HostID string `json:"host_id"`
		Active bool   `json:"active"`
	}{
		HostID: s.hostID,
		Active: true,
	})

}

func (s *Service) createProcess(rw http.ResponseWriter, r *http.Request) {
	opts := &CreateOptions{}
	if err := gimlet.GetJSON(r.Body, opts); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "problem reading request").Error(),
		})
		return
	}

	if err := opts.Validate(); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "invalid creation options").Error(),
		})
		return
	}

	// If a logger is attached, it should be in-memory.
	if len(opts.Output.Loggers) > 1 {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.New("logger output only allows 1 logger").Error(),
		})
		return
	} else if len(opts.Output.Loggers) == 1 && opts.Output.Loggers[0].Type != LogInMemory {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.New("logger must be in-memory").Error(),
		})
		return
	}

	var ctx context.Context
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), opts.Timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}

	proc, err := s.manager.Create(ctx, opts)
	if err != nil {
		cancel()
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "problem submitting request").Error(),
		})
		return
	}

	if err := proc.RegisterTrigger(ctx, func(_ ProcessInfo) {
		cancel()
	}); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    errors.Wrap(err, "problem managing resources").Error(),
		})
		return
	}

	gimlet.WriteJSON(rw, proc.Info(r.Context()))
}

func (s *Service) listProcesses(rw http.ResponseWriter, r *http.Request) {
	filter := Filter(gimlet.GetVars(r)["filter"])
	if err := filter.Validate(); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "invalid input").Error(),
		})
		return
	}

	ctx := r.Context()

	procs, err := s.manager.List(ctx, filter)
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    err.Error(),
		})
		return
	}

	out := []ProcessInfo{}
	for _, proc := range procs {
		out = append(out, proc.Info(ctx))
	}

	gimlet.WriteJSON(rw, out)
}

func (s *Service) listGroupMembers(rw http.ResponseWriter, r *http.Request) {
	name := gimlet.GetVars(r)["name"]

	ctx := r.Context()

	procs, err := s.manager.Group(ctx, name)
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    err.Error(),
		})
		return
	}

	out := []ProcessInfo{}
	for _, proc := range procs {
		out = append(out, proc.Info(ctx))
	}

	gimlet.WriteJSON(rw, out)
}

func (s *Service) getProcess(rw http.ResponseWriter, r *http.Request) {
	id := gimlet.GetVars(r)["id"]
	ctx := r.Context()
	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    errors.Wrapf(err, "no process '%s' found", id).Error(),
		})
		return
	}

	info := proc.Info(ctx)
	if info.ID == "" {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    fmt.Sprintf("no process '%s' found", id),
		})
		return
	}

	gimlet.WriteJSON(rw, info)
}

func (s *Service) processMetrics(rw http.ResponseWriter, r *http.Request) {
	id := gimlet.GetVars(r)["id"]
	ctx := r.Context()
	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    errors.Wrapf(err, "no process '%s' found", id).Error(),
		})
		return
	}

	info := proc.Info(ctx)
	gimlet.WriteJSON(rw, message.CollectProcessInfoWithChildren(int32(info.PID)))
}

func (s *Service) getProcessTags(rw http.ResponseWriter, r *http.Request) {
	id := gimlet.GetVars(r)["id"]
	ctx := r.Context()
	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    errors.Wrapf(err, "no process '%s' found", id).Error(),
		})
		return
	}

	gimlet.WriteJSON(rw, proc.GetTags())
}

func (s *Service) deleteProcessTags(rw http.ResponseWriter, r *http.Request) {
	id := gimlet.GetVars(r)["id"]
	ctx := r.Context()
	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    errors.Wrapf(err, "no process '%s' found", id).Error(),
		})
		return
	}

	proc.ResetTags()
	gimlet.WriteJSON(rw, struct{}{})
}

func (s *Service) addProcessTag(rw http.ResponseWriter, r *http.Request) {
	id := gimlet.GetVars(r)["id"]
	ctx := r.Context()
	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    errors.Wrapf(err, "no process '%s' found", id).Error(),
		})
		return
	}

	newtags := r.URL.Query()["add"]
	if len(newtags) == 0 {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "no new tags specified",
		})
		return
	}

	for _, t := range newtags {
		proc.Tag(t)
	}

	gimlet.WriteJSON(rw, struct{}{})
}

func (s *Service) waitForProcess(rw http.ResponseWriter, r *http.Request) {
	id := gimlet.GetVars(r)["id"]
	ctx := r.Context()
	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    errors.Wrapf(err, "no process '%s' found", id).Error(),
		})
		return
	}

	if err := proc.Wait(ctx); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}

	gimlet.WriteJSON(rw, struct{}{})
}

func (s *Service) signalProcess(rw http.ResponseWriter, r *http.Request) {
	vars := gimlet.GetVars(r)
	id := vars["id"]
	sig, err := strconv.Atoi(vars["signal"])
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrapf(err, "problem finding signal '%s'", vars["signal"]).Error(),
		})
		return
	}

	ctx := r.Context()
	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    errors.Wrapf(err, "no process '%s' found", id).Error(),
		})
		return
	}

	if err := proc.Signal(ctx, syscall.Signal(sig)); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}

	gimlet.WriteJSON(rw, struct{}{})
}

func (s *Service) getLogs(rw http.ResponseWriter, r *http.Request) {
	vars := gimlet.GetVars(r)
	id := vars["id"]
	ctx := r.Context()

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    errors.Wrapf(err, "no process '%s' found", id).Error(),
		})
		return
	}

	if proc.Info(ctx).Options.Output.outputSender == nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    errors.Errorf("process '%s' is not being logged", id).Error(),
		})
		return
	}
	logger, ok := proc.Info(ctx).Options.Output.outputSender.Sender.(*send.InMemorySender)
	if !ok {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    errors.Errorf("cannot get log from process '%s'", id).Error(),
		})
		return
	}
	logs, err := logger.GetString()
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    err.Error(),
		})
		return
	}
	gimlet.WriteJSON(rw, logs)
}

func (s *Service) closeManager(rw http.ResponseWriter, r *http.Request) {
	if err := s.manager.Close(r.Context()); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}

	gimlet.WriteJSON(rw, struct{}{})
}

// MongoDBDownloadOptions represent one build variant of MongoDB.
type MongoDBDownloadOptions struct {
	BuildOpts bond.BuildOptions `json:"build_opts"`
	Path      string            `json:"path"`
	Releases  []string          `json:"releases"`
}

func (s *Service) downloadMongoDB(rw http.ResponseWriter, r *http.Request) {
	opts := MongoDBDownloadOptions{}
	if err := gimlet.GetJSON(r.Body, &opts); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "problem reading request").Error(),
		})
		return
	}

	if err := opts.BuildOpts.Validate(); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "invalid build options").Error(),
		})
		return
	}

	if err := SetupDownloadMongoDBReleases(r.Context(), s.cache, opts); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    errors.Wrap(err, "problem in download setup").Error(),
		})
		return
	}

	gimlet.WriteJSON(rw, struct{}{})
}

// SetupDownloadMongoDBReleases performs necessary setup to download MongoDB with the given options.
func SetupDownloadMongoDBReleases(ctx context.Context, cache *lru.Cache, opts MongoDBDownloadOptions) error {
	if err := opts.BuildOpts.Validate(); err != nil {
		return errors.Wrap(err, "invalid build options")
	}

	if err := MakeEnclosingDirectories(opts.Path); err != nil {
		return errors.Wrap(err, "problem creating enclosing directories")
	}

	absRootPath, err := filepath.Abs(opts.Path)
	if err != nil {
		return errors.Errorf("problem getting absolute path to %s: %s", opts.Path, err)
	}
	addToCache := addMongoDBFilesToCache(cache, absRootPath)

	feed, err := bond.GetArtifactsFeed(opts.Path)
	if err != nil {
		return errors.Wrap(err, "problem making artifacts feed")
	}

	catcher := grip.NewBasicCatcher()
	urls, errs1 := feed.GetArchives(opts.Releases, opts.BuildOpts)
	jobs, errs2 := createDownloadJobs(opts.Path, urls)

	if err := setupDownloadJobsAsync(ctx, jobs, processDownloadJobs(context.Background(), addToCache)); err != nil {
		catcher.Add(errors.Wrap(err, "problem starting download jobs"))
	}

	if err := aggregateErrors(errs1, errs2); err != nil {
		catcher.Add(errors.Wrap(err, "problem initializing download jobs"))
	}

	return catcher.Resolve()
}

func setupDownloadJobsAsync(ctx context.Context, jobs <-chan amboy.Job, processJobs func(amboy.Queue) error) error {
	q := queue.NewLocalUnordered(runtime.NumCPU())
	if err := q.Start(ctx); err != nil {
		return errors.Wrap(err, "problem starting download job queue")
	}

	if err := amboy.PopulateQueue(ctx, q, jobs); err != nil {
		return errors.Wrap(err, "problem adding download jobs to queue")
	}

	go func() {
		if err := processJobs(q); err != nil {
			grip.Errorf(errors.Wrap(err, "error occurred while adding jobs to cache").Error())
		}
	}()

	return nil
}

func addMongoDBFilesToCache(cache *lru.Cache, absRootPath string) func(string) error {
	return func(fileName string) error {
		catcher := grip.NewBasicCatcher()

		filePath := filepath.Join(absRootPath, fileName)
		if err := cache.AddFile(filePath); err != nil {
			catcher.Add(errors.Wrap(err, "problem adding file to cache"))
		}

		baseName := filepath.Base(fileName)
		ext := filepath.Ext(baseName)
		dirPath := filepath.Join(absRootPath, string(baseName[:len(baseName)-len(ext)]))

		err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Cache only handles individual files, not directories.
			if !info.IsDir() {
				if err := cache.AddStat(path, info); err != nil {
					catcher.Add(errors.Wrap(err, "problem adding file to cache"))
				}
			}

			return nil
		})
		if err != nil {
			catcher.Add(errors.Wrap(err, "problem walking download files"))
		}
		return catcher.Resolve()
	}
}

func processDownloadJobs(ctx context.Context, processFile func(string) error) func(amboy.Queue) error {
	return func(q amboy.Queue) error {
		grip.Infof("waiting for %d download jobs to complete", q.Stats().Total)
		_ = amboy.WaitCtxInterval(ctx, q, 1000*time.Millisecond)
		grip.Info("all download tasks complete, processing errors now")

		if err := amboy.ResolveErrors(ctx, q); err != nil {
			return errors.Wrap(err, "problem completing download jobs")
		}

		catcher := grip.NewBasicCatcher()
		for job := range q.Results(ctx) {
			downloadJob, ok := job.(*recall.DownloadFileJob)
			if !ok {
				catcher.Add(errors.New("problem retrieving download job from queue"))
				continue
			}
			if err := processFile(downloadJob.FileName); err != nil {
				catcher.Add(err)
			}
		}
		return catcher.Resolve()
	}
}

func createDownloadJobs(path string, urls <-chan string) (<-chan amboy.Job, <-chan error) {
	output := make(chan amboy.Job)
	errOut := make(chan error)

	go func() {
		catcher := grip.NewCatcher()
		for url := range urls {
			j, err := recall.NewDownloadJob(url, path, false)
			if err != nil {
				catcher.Add(errors.Wrapf(err, "problem generating task for %s", url))
				continue
			}

			output <- j
		}
		close(output)
		if catcher.HasErrors() {
			errOut <- catcher.Resolve()
		}
		close(errOut)
	}()

	return output, errOut
}

func aggregateErrors(groups ...<-chan error) error {
	catcher := grip.NewCatcher()

	for _, g := range groups {
		for err := range g {
			catcher.Add(err)
		}
	}

	return catcher.Resolve()
}

// CacheOptions represent the configuration options for the LRU cache.
type CacheOptions struct {
	Disabled   bool          `json:"disabled"`
	PruneDelay time.Duration `json:"prune_delay"`
	MaxSize    int           `json:"max_size"`
}

// Validate checks for valid cache options.
func (opts CacheOptions) Validate() error {
	catcher := grip.NewBasicCatcher()
	if opts.MaxSize < 0 {
		catcher.Add(errors.New("max size cannot be negative"))
	}
	if opts.PruneDelay < 0 {
		catcher.Add(errors.New("prune delay cannot be negative"))
	}
	return catcher.Resolve()
}

func (s *Service) configureCache(rw http.ResponseWriter, r *http.Request) {
	opts := CacheOptions{}
	if err := gimlet.GetJSON(r.Body, &opts); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "problem reading request").Error(),
		})
		return
	}
	if err := opts.Validate(); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "problem with validating cache options").Error(),
		})
		return
	}

	s.cacheMutex.Lock()
	if opts.MaxSize > 0 {
		s.cacheOpts.MaxSize = opts.MaxSize
	}
	if opts.PruneDelay > time.Duration(0) {
		s.cacheOpts.PruneDelay = opts.PruneDelay
	}
	s.cacheOpts.Disabled = opts.Disabled
	s.cacheMutex.Unlock()

	gimlet.WriteJSON(rw, struct{}{})
}
