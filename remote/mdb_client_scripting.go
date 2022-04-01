package remote

import (
	"context"

	"github.com/evergreen-ci/mrpc/mongowire"
	"github.com/evergreen-ci/mrpc/shell"
	"github.com/mongodb/jasper/scripting"
	"github.com/pkg/errors"
)

// mdbScriptingHarness is the client-side representation of a scripting.Harness
// for making requests to the remote MDB wire protocol service.
type mdbScriptingHarness struct {
	client *mdbClient
	id     string
}

func (s *mdbScriptingHarness) ID() string { return s.id }
func (s *mdbScriptingHarness) Setup(ctx context.Context) error {
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, &scriptingSetupRequest{ID: s.id})
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	msg, err := s.client.doRequest(ctx, req)
	if err != nil {
		return errors.Wrap(err, "making request")
	}

	resp := &shell.ErrorResponse{}
	if err = shell.MessageToResponse(msg, resp); err != nil {
		return errors.Wrap(err, "converting wire message to response")
	}

	return errors.Wrap(resp.SuccessOrError(), "response contained error")
}

func (s *mdbScriptingHarness) Cleanup(ctx context.Context) error {
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, &scriptingCleanupRequest{ID: s.id})
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	msg, err := s.client.doRequest(ctx, req)
	if err != nil {
		return errors.Wrap(err, "making request")
	}

	resp := &shell.ErrorResponse{}
	if err = shell.MessageToResponse(msg, resp); err != nil {
		return errors.Wrap(err, "converting wire message to response")
	}

	return errors.Wrap(resp.SuccessOrError(), "response contained error")
}

func (s *mdbScriptingHarness) Run(ctx context.Context, args []string) error {
	r := &scriptingRunRequest{}
	r.Params.ID = s.id
	r.Params.Args = args
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, r)
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	msg, err := s.client.doRequest(ctx, req)
	if err != nil {
		return errors.Wrap(err, "making request")
	}

	resp := &shell.ErrorResponse{}
	if err = shell.MessageToResponse(msg, resp); err != nil {
		return errors.Wrap(err, "converting wire message to response")
	}

	return errors.Wrap(resp.SuccessOrError(), "response contained error")
}

func (s *mdbScriptingHarness) RunScript(ctx context.Context, in string) error {
	r := &scriptingRunScriptRequest{}
	r.Params.ID = s.id
	r.Params.Script = in
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, r)
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	msg, err := s.client.doRequest(ctx, req)
	if err != nil {
		return errors.Wrap(err, "making request")
	}

	resp := &shell.ErrorResponse{}
	if err = shell.MessageToResponse(msg, resp); err != nil {
		return errors.Wrap(err, "converting wire message to response")
	}

	return errors.Wrap(resp.SuccessOrError(), "response contained error")
}

func (s *mdbScriptingHarness) Build(ctx context.Context, dir string, args []string) (string, error) {
	r := &scriptingBuildRequest{}
	r.Params.ID = s.id
	r.Params.Dir = dir
	r.Params.Args = args
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, r)
	if err != nil {
		return "", errors.Wrap(err, "creating request")
	}

	msg, err := s.client.doRequest(ctx, req)
	if err != nil {
		return "", errors.Wrap(err, "making request")
	}

	resp := &scriptingBuildResponse{}
	if err = shell.MessageToResponse(msg, resp); err != nil {
		return "", errors.Wrap(err, "converting wire message to response")
	}

	return resp.Path, errors.Wrap(resp.SuccessOrError(), "response contained error")
}

func (s *mdbScriptingHarness) Test(ctx context.Context, dir string, opts ...scripting.TestOptions) ([]scripting.TestResult, error) {
	r := &scriptingTestRequest{}
	r.Params.ID = s.id
	r.Params.Dir = dir
	r.Params.Options = opts
	req, err := shell.RequestToMessage(mongowire.OP_QUERY, r)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	msg, err := s.client.doRequest(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "making request")
	}

	resp := &scriptingTestResponse{}
	if err = shell.MessageToResponse(msg, resp); err != nil {
		return nil, errors.Wrap(err, "converting wire message to response")
	}

	return resp.Results, errors.Wrap(resp.SuccessOrError(), "response contained error")
}
