package cli

import (
	"context"

	"github.com/mongodb/jasper/scripting"
	"github.com/pkg/errors"
)

type sshClientScriptingHarness struct {
	id     string
	client *sshRunner
}

func newSSHClientScriptingHarness(client *sshRunner, id string) *sshClientScriptingHarness {
	return &sshClientScriptingHarness{
		id:     id,
		client: client,
	}
}

func (s *sshClientScriptingHarness) ID() string { return s.id }

// kim: TODO: implement

func (s *sshClientScriptingHarness) Setup(ctx context.Context) error {
	return errors.New("not implemented")
}

func (s *sshClientScriptingHarness) Run(ctx context.Context, args []string) error {
	return errors.New("not implemented")
}

func (s *sshClientScriptingHarness) RunScript(ctx context.Context, script string) error {
	return errors.New("not implemented")
}

func (s *sshClientScriptingHarness) Build(ctx context.Context, dir string, args []string) (string, error) {
	return "", errors.New("not implemented")
}

func (s *sshClientScriptingHarness) Test(ctx context.Context, dir string, opts ...scripting.TestOptions) ([]scripting.TestResult, error) {
	return nil, errors.New("not implemented")
}

func (s *sshClientScriptingHarness) Cleanup(ctx context.Context) error {
	return errors.New("not implemented")
}
