package remote

import (
	"context"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/evergreen-ci/gimlet"
	"github.com/evergreen-ci/lru"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/mongodb/grip/recovery"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
)

// Service defines a REST service that provides a remote manager, using
// gimlet to publish routes.
type Service struct {
	hostID     string
	manager    jasper.Manager
	cache      *lru.Cache
	cacheOpts  options.Cache
	cacheMutex sync.RWMutex
}

// NewRESTService creates a service object around an existing manager. You must
// access the application and routes via the App() method separately.
func NewRESTService(m jasper.Manager) *Service {
	return &Service{
		manager: m,
		cache:   lru.NewCache(),
	}
}

// App constructs and returns a gimlet application for this
// service. It attaches no middleware and does not start the service.
func (s *Service) App(ctx context.Context) *gimlet.APIApp {
	s.hostID, _ = os.Hostname()

	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	s.cacheOpts.PruneDelay = jasper.DefaultCachePruneDelay
	s.cacheOpts.MaxSize = jasper.DefaultMaxCacheSize
	s.cacheOpts.Disabled = false

	app := gimlet.NewApp()

	app.AddRoute("/").Version(1).Get().Handler(s.rootRoute)
	app.AddRoute("/id").Version(1).Get().Handler(s.id)
	app.AddRoute("/create").Version(1).Post().Handler(s.createProcess)
	app.AddRoute("/download").Version(1).Post().Handler(s.downloadFile)
	app.AddRoute("/download/cache").Version(1).Post().Handler(s.configureCache)
	app.AddRoute("/download/mongodb").Version(1).Post().Handler(s.downloadMongoDB)
	app.AddRoute("/list/oom").Version(1).Get().Handler(s.oomTrackerList)
	app.AddRoute("/list/oom").Version(1).Delete().Handler(s.oomTrackerClear)
	app.AddRoute("/list/{filter}").Version(1).Get().Handler(s.listProcesses)
	app.AddRoute("/list/group/{name}").Version(1).Get().Handler(s.listGroupMembers)
	app.AddRoute("/process/{id}").Version(1).Get().Handler(s.getProcess)
	app.AddRoute("/process/{id}/tags").Version(1).Get().Handler(s.getProcessTags)
	app.AddRoute("/process/{id}/tags").Version(1).Delete().Handler(s.deleteProcessTags)
	app.AddRoute("/process/{id}/tags").Version(1).Post().Handler(s.addProcessTag)
	app.AddRoute("/process/{id}/wait").Version(1).Get().Handler(s.waitForProcess)
	app.AddRoute("/process/{id}/respawn").Version(1).Get().Handler(s.respawnProcess)
	app.AddRoute("/process/{id}/metrics").Version(1).Get().Handler(s.processMetrics)
	app.AddRoute("/process/{id}/logs/{count}").Version(1).Get().Handler(s.getLogStream)
	app.AddRoute("/process/{id}/loginfo").Version(1).Get().Handler(s.getBuildloggerURLs)
	app.AddRoute("/process/{id}/signal/{signal}").Version(1).Patch().Handler(s.signalProcess)
	app.AddRoute("/process/{id}/trigger/signal/{trigger-id}").Version(1).Patch().Handler(s.registerSignalTriggerID)
	app.AddRoute("/signal/event/{name}").Version(1).Patch().Handler(s.signalEvent)
	app.AddRoute("/logging/id/{id}").Version(1).Post().Handler(s.loggingCacheCreate)
	app.AddRoute("/logging/id/{id}").Version(1).Get().Handler(s.loggingCacheGet)
	app.AddRoute("/logging/id/{id}").Version(1).Delete().Handler(s.loggingCacheRemove)
	app.AddRoute("/logging/id/{id}/close").Version(1).Delete().Handler(s.loggingCacheCloseAndRemove)
	app.AddRoute("/logging/clear").Version(1).Delete().Handler(s.loggingCacheClear)
	app.AddRoute("/logging/prune/{time}").Version(1).Delete().Handler(s.loggingCachePrune)
	app.AddRoute("/logging/len").Version(1).Get().Handler(s.loggingCacheLen)
	app.AddRoute("/logging/id/{id}/send").Version(1).Post().Handler(s.sendMessages)
	app.AddRoute("/file/write").Version(1).Put().Handler(s.writeFile)
	app.AddRoute("/clear").Version(1).Post().Handler(s.clearManager)
	app.AddRoute("/close").Version(1).Delete().Handler(s.closeManager)

	go s.pruneCache(ctx)

	return app
}

// SetDisableCachePruning toggles the underlying option for the
// services cache.
func (s *Service) SetDisableCachePruning(v bool) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.cacheOpts.Disabled = v
}

// SetCacheMaxSize sets the underlying option for the
// services cache.
func (s *Service) SetCacheMaxSize(size int) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.cacheOpts.MaxSize = size
}

// SetPruneDelay sets the underlying option for the
// services cache.
func (s *Service) SetPruneDelay(dur time.Duration) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.cacheOpts.PruneDelay = dur
}

func (s *Service) pruneCache(ctx context.Context) {
	defer func() {
		err := recovery.HandlePanicWithError(recover(), nil, "background pruning")
		if ctx.Err() != nil || err == nil {
			return
		}
		go s.pruneCache(ctx)
	}()

	s.cacheMutex.RLock()
	timer := time.NewTimer(s.cacheOpts.PruneDelay)
	s.cacheMutex.RUnlock()

	for {
		select {
		case <-timer.C:
			s.cacheMutex.RLock()
			if !s.cacheOpts.Disabled {
				if err := s.cache.Prune(s.cacheOpts.MaxSize, nil, false); err != nil {
					grip.Error(errors.Wrap(err, "pruning cache"))
				}
			}
			timer.Reset(s.cacheOpts.PruneDelay)
			s.cacheMutex.RUnlock()
		case <-ctx.Done():
			return
		}
	}
}

func getProcInfoNoHang(ctx context.Context, p jasper.Process) jasper.ProcessInfo {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	return p.Info(ctx)
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

func (s *Service) id(rw http.ResponseWriter, r *http.Request) {
	gimlet.WriteJSON(rw, s.manager.ID())
}

func (s *Service) createProcess(rw http.ResponseWriter, r *http.Request) {
	opts := &options.Create{}
	if err := gimlet.GetJSON(r.Body, opts); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "reading create options from request").Error(),
		})
		return
	}
	ctx := r.Context()

	if err := opts.Validate(); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "invalid creation options").Error(),
		})
		return
	}

	pctx, cancel := context.WithCancel(context.Background())

	proc, err := s.manager.CreateProcess(pctx, opts)
	if err != nil {
		cancel()
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "creating process").Error(),
		})
		return
	}

	if err := proc.RegisterTrigger(ctx, func(_ jasper.ProcessInfo) {
		cancel()
	}); err != nil {
		info := getProcInfoNoHang(ctx, proc)
		cancel()
		// If we get an error registering a trigger, then we should make sure
		// that the reason for it isn't just because the process has exited
		// already, since that should not be considered an error.
		if !info.Complete {
			writeError(rw, gimlet.ErrorResponse{
				StatusCode: http.StatusInternalServerError,
				Message:    errors.Wrap(err, "registering trigger").Error(),
			})
			return
		}
	}

	gimlet.WriteJSON(rw, getProcInfoNoHang(ctx, proc))
}

func (s *Service) getBuildloggerURLs(rw http.ResponseWriter, r *http.Request) {
	id := gimlet.GetVars(r)["id"]
	ctx := r.Context()

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrapf(err, "no process '%s' found", id).Error(),
		})
		return
	}

	info := getProcInfoNoHang(ctx, proc)
	urls := []string{}
	for _, logger := range info.Options.Output.Loggers {
		if logger.Type() == options.LogBuildloggerV2 {
			producer := logger.Producer()
			if producer == nil {
				continue
			}
			rawProducer, ok := producer.(*options.BuildloggerV2Options)
			if ok {
				urls = append(urls, rawProducer.Buildlogger.GetGlobalLogURL())
			}
		}
	}

	if len(urls) == 0 {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Errorf("process '%s' does not use buildlogger", id).Error(),
		})
		return
	}

	gimlet.WriteJSON(rw, urls)
}

func (s *Service) listProcesses(rw http.ResponseWriter, r *http.Request) {
	filter := options.Filter(gimlet.GetVars(r)["filter"])
	if err := filter.Validate(); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "invalid filter").Error(),
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

	out := []jasper.ProcessInfo{}
	for _, proc := range procs {
		out = append(out, getProcInfoNoHang(ctx, proc))
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

	out := []jasper.ProcessInfo{}
	for _, proc := range procs {
		out = append(out, getProcInfoNoHang(ctx, proc))
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

	info := getProcInfoNoHang(ctx, proc)
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

	info := getProcInfoNoHang(ctx, proc)
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

type restWaitResponse struct {
	Error    string `json:"error,omitempty"`
	ExitCode int    `json:"exit_code,omitempty"`
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

	exitCode, err := proc.Wait(ctx)
	if err != nil {
		gimlet.WriteJSON(rw, restWaitResponse{
			Error:    err.Error(),
			ExitCode: exitCode,
		})
		return
	}

	gimlet.WriteJSON(rw, restWaitResponse{ExitCode: exitCode})
}

func (s *Service) respawnProcess(rw http.ResponseWriter, r *http.Request) {
	id := gimlet.GetVars(r)["id"]
	ctx := r.Context()

	proc, err := s.manager.Get(r.Context(), id)
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    errors.Wrapf(err, "no process '%s' found", id).Error(),
		})
		return
	}

	// Spawn a new context so that the process' context is not potentially
	// canceled by the request's. See how createProcess() does this same thing.
	pctx, cancel := context.WithCancel(context.Background())
	newProc, err := proc.Respawn(pctx)
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		cancel()
		return
	}
	if err := s.manager.Register(ctx, newProc); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    errors.Wrap(err, "registering respawned process").Error(),
		})
		cancel()
		return
	}

	if err := newProc.RegisterTrigger(ctx, func(_ jasper.ProcessInfo) {
		cancel()
	}); err != nil {
		newProcInfo := getProcInfoNoHang(ctx, newProc)
		cancel()
		if !newProcInfo.Complete {
			writeError(rw, gimlet.ErrorResponse{
				StatusCode: http.StatusInternalServerError,
				Message:    errors.Wrap(err, "registering trigger on respawned process").Error(),
			})
			return
		}
	}

	info := getProcInfoNoHang(ctx, newProc)
	gimlet.WriteJSON(rw, info)
}

func (s *Service) signalProcess(rw http.ResponseWriter, r *http.Request) {
	vars := gimlet.GetVars(r)
	id := vars["id"]
	sig, err := strconv.Atoi(vars["signal"])
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrapf(err, "invalid signal '%s'", vars["signal"]).Error(),
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

func (s *Service) downloadFile(rw http.ResponseWriter, r *http.Request) {
	var opts options.Download
	if err := gimlet.GetJSON(r.Body, &opts); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "reading download options from request").Error(),
		})
		return
	}

	if err := opts.Validate(); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "invalid download options").Error(),
		})
		return
	}

	if err := opts.Download(); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    errors.Wrapf(err, "downloading file for URL '%s'", opts.URL).Error(),
		})
		return
	}

	gimlet.WriteJSON(rw, struct{}{})
}

func (s *Service) getLogStream(rw http.ResponseWriter, r *http.Request) {
	vars := gimlet.GetVars(r)
	id := vars["id"]
	count, err := strconv.Atoi(vars["count"])
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrapf(err, "invalid count '%s'", vars["count"]).Error(),
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

	stream := jasper.LogStream{}
	stream.Logs, err = jasper.GetInMemoryLogStream(ctx, proc, count)

	if err == io.EOF {
		stream.Done = true
	} else if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    errors.Wrapf(err, "getting logs for process '%s'", id).Error(),
		})
		return
	}

	gimlet.WriteJSON(rw, stream)
}

func (s *Service) signalEvent(rw http.ResponseWriter, r *http.Request) {
	vars := gimlet.GetVars(r)
	name := vars["name"]
	ctx := r.Context()

	if err := jasper.SignalEvent(ctx, name); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    errors.Wrapf(err, "signaling event '%s'", name).Error(),
		})
		return
	}

	gimlet.WriteJSON(rw, struct{}{})
}

func (s *Service) writeFile(rw http.ResponseWriter, r *http.Request) {
	var opts options.WriteFile
	if err := gimlet.GetJSON(r.Body, &opts); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "reading file write options from request").Error(),
		})
		return
	}

	if err := opts.Validate(); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "invalid file write options").Error(),
		})
		return
	}

	if err := opts.DoWrite(); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    errors.Wrapf(err, "writing file '%s'", opts.Path).Error(),
		})
		return
	}

	if err := opts.SetPerm(); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    errors.Wrapf(err, "setting permissions on file '%s'", opts.Path).Error(),
		})
		return
	}

	gimlet.WriteJSON(rw, struct{}{})
}

func (s *Service) clearManager(rw http.ResponseWriter, r *http.Request) {
	s.manager.Clear(r.Context())
	gimlet.WriteJSON(rw, struct{}{})
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

func (s *Service) configureCache(rw http.ResponseWriter, r *http.Request) {
	opts := options.Cache{}
	if err := gimlet.GetJSON(r.Body, &opts); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "reading cache options from request").Error(),
		})
		return
	}

	if err := opts.Validate(); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "invalid cache options").Error(),
		})
		return
	}

	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	if opts.MaxSize > 0 {
		s.cacheOpts.MaxSize = opts.MaxSize
	}
	if opts.PruneDelay > time.Duration(0) {
		s.cacheOpts.PruneDelay = opts.PruneDelay
	}
	s.cacheOpts.Disabled = opts.Disabled

	gimlet.WriteJSON(rw, struct{}{})
}

func (s *Service) downloadMongoDB(rw http.ResponseWriter, r *http.Request) {
	opts := options.MongoDBDownload{}
	if err := gimlet.GetJSON(r.Body, &opts); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "reading MongoDB download options from request").Error(),
		})
		return
	}

	if err := opts.Validate(); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "invalid MongoDB download options").Error(),
		})
		return
	}

	if err := jasper.SetupDownloadMongoDBReleases(r.Context(), s.cache, opts); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    errors.Wrap(err, "setting up download").Error(),
		})
		return
	}

	gimlet.WriteJSON(rw, struct{}{})
}

func (s *Service) registerSignalTriggerID(rw http.ResponseWriter, r *http.Request) {
	vars := gimlet.GetVars(r)
	id := vars["id"]
	triggerID := vars["trigger-id"]
	ctx := r.Context()

	proc, err := s.manager.Get(ctx, id)
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrapf(err, "no process '%s' found", id).Error(),
		})
		return
	}

	sigTriggerID := jasper.SignalTriggerID(triggerID)
	makeTrigger, ok := jasper.GetSignalTriggerFactory(sigTriggerID)
	if !ok {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Errorf("getting signal trigger '%s'", sigTriggerID).Error(),
		})
		return
	}

	if err := proc.RegisterSignalTrigger(ctx, makeTrigger()); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    errors.Wrapf(err, "registering signal trigger '%s'", sigTriggerID).Error(),
		})
		return
	}

	gimlet.WriteJSON(rw, struct{}{})
}

func (s *Service) sendMessages(rw http.ResponseWriter, r *http.Request) {
	id := gimlet.GetVars(r)["id"]
	lc := s.manager.LoggingCache(r.Context())
	if lc == nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    ErrLoggingCacheNotSupported.Error(),
		})
		return
	}
	logger, err := lc.Get(id)
	if err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    err.Error(),
		})
		return
	}

	payload := &options.LoggingPayload{}
	if err := gimlet.GetJSON(r.Body, payload); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    errors.Wrap(err, "parsing payload").Error(),
		})
		return
	}

	if err := logger.Send(payload); err != nil {
		writeError(rw, gimlet.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}

	gimlet.WriteJSON(rw, struct{}{})
}

func (s *Service) oomTrackerClear(rw http.ResponseWriter, r *http.Request) {
	resp := jasper.NewOOMTracker()

	if err := resp.Clear(r.Context()); err != nil {
		gimlet.WriteJSONInternalError(rw, err.Error())
		return
	}

	gimlet.WriteJSON(rw, resp)
}

func (s *Service) oomTrackerList(rw http.ResponseWriter, r *http.Request) {
	resp := jasper.NewOOMTracker()

	if err := resp.Check(r.Context()); err != nil {
		gimlet.WriteJSONInternalError(rw, err.Error())
		return
	}

	gimlet.WriteJSON(rw, resp)
}
