package jasper

import (
	"context"
	"net/http"
	"strconv"
	"syscall"

	"github.com/evergreen-ci/gimlet"
	"github.com/pkg/errors"
)

// Service defines a REST service that provides a remote manager, using
// gimlet to publish routes.
type Service struct {
	manager Manager
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

// App constructs and returns a gimlet application for this
// service. It attaches no middleware and does not start the service.
func (s *Service) App() *gimlet.APIApp {
	app := gimlet.NewApp()

	app.AddRoute("/create").Version(1).Post().Handler(s.createProcess)
	app.AddRoute("/list/{filter}").Version(1).Get().Handler(s.listProcesses)
	app.AddRoute("/list/group/{name}").Version(1).Get().Handler(s.listGroupMembers)
	app.AddRoute("/process/{id}").Version(1).Get().Handler(s.getProcess)
	app.AddRoute("/process/{id}/tags").Version(1).Get().Handler(s.getProcessTags)
	app.AddRoute("/process/{id}/tags").Version(1).Delete().Handler(s.deleteProcessTags)
	app.AddRoute("/process/{id}/tags").Version(1).Post().Handler(s.addProcessTag)
	app.AddRoute("/process/{id}/wait").Version(1).Get().Handler(s.waitForProcess)
	app.AddRoute("/process/{id}/signal/{signal}").Version(1).Patch().Handler(s.signalProcess)
	app.AddRoute("/close").Version(1).Delete().Handler(s.closeManager)

	return app
}

func writeError(rw http.ResponseWriter, err gimlet.ErrorResponse) {
	gimlet.WriteJSONResponse(rw, err.StatusCode, err)
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
			Message:    errors.Wrapf(err, "no process '%s' found", id).Error(),
		})
		return
	}

	gimlet.WriteJSON(rw, info)
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
