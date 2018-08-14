package internal

import (
	"context"
	"fmt"
	"os"
	"strings"

	empty "github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/tychoish/jasper"
	grpc "google.golang.org/grpc"
)

func AttachService(manager jasper.Manager, s *grpc.Server) error {
	hn, err := os.Hostname()
	if err != nil {
		return errors.WithStack(err)
	}

	srv := &jasperService{
		hostID:  hn,
		manager: manager,
	}

	RegisterJasperProcessManagerServer(s, srv)

	return nil
}

type jasperService struct {
	hostID  string
	manager jasper.Manager
}

func (s *jasperService) Status(ctx context.Context, _ *empty.Empty) (*StatusResponse, error) {
	return &StatusResponse{
		HostId: s.hostID,
		Active: true,
	}, nil
}

func (s *jasperService) Create(ctx context.Context, opts *CreateOptions) (*ProcessInfo, error) {
	jopt := opts.Export()
	proc, err := s.manager.Create(ctx, jopt)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return ConvertProcessInfo(proc.Info(ctx)), nil
}

func (s *jasperService) List(f *Filter, stream JasperProcessManager_ListServer) error {
	ctx := stream.Context()
	procs, err := s.manager.List(ctx, jasper.Filter(strings.ToLower(f.GetName().String())))
	if err != nil {
		return errors.WithStack(err)
	}

	for _, p := range procs {
		if ctx.Err() != nil {
			return errors.New("operation canceled")
		}

		if err := stream.Send(ConvertProcessInfo(p.Info(ctx))); err != nil {
			return errors.Wrap(err, "problem sending process info")
		}
	}

	return nil
}

func (s *jasperService) Group(t *TagName, stream JasperProcessManager_GroupServer) error {
	ctx := stream.Context()
	procs, err := s.manager.Group(ctx, t.Value)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, p := range procs {
		if ctx.Err() != nil {
			return errors.New("operation canceled")
		}

		if err := stream.Send(ConvertProcessInfo(p.Info(ctx))); err != nil {
			return errors.Wrap(err, "problem sending process info")
		}
	}

	return nil
}

func (s *jasperService) Get(ctx context.Context, id *JasperProcessID) (*ProcessInfo, error) {
	proc, err := s.manager.Get(ctx, id.Value)
	if err != nil {
		return nil, errors.Wrapf(err, "problem fetching process '%s'", id.Value)
	}

	return ConvertProcessInfo(proc.Info(ctx)), nil
}

func (s *jasperService) Signal(ctx context.Context, sig *SignalProcess) (*OperationOutcome, error) {
	proc, err := s.manager.Get(ctx, sig.ProcessID.Value)
	if err != nil {
		err = errors.Wrapf(err, "couldn't find process with id '%s'", sig.ProcessID)
		return &OperationOutcome{
			Succuess: false,
			Text:     err.Error(),
		}, err
	}

	if err = proc.Signal(ctx, sig.Signal.Export()); err != nil {
		err = errors.Wrapf(err, "problem sending '%s' to '%s'", sig.Signal, sig.ProcessID)
		return &OperationOutcome{
			Succuess: false,
			Text:     err.Error(),
		}, err
	}

	out := &OperationOutcome{Succuess: true, Text: fmt.Sprintf("sending '%s' to '%s'", sig.Signal, sig.ProcessID)}
	return out, nil
}

func (s *jasperService) Wait(ctx context.Context, id *JasperProcessID) (*OperationOutcome, error) {
	proc, err := s.manager.Get(ctx, id.Value)
	if err != nil {
		err = errors.Wrapf(err, "problem finding process '%s'", id.Value)
		return &OperationOutcome{
			Succuess: false,
			Text:     err.Error(),
		}, err
	}

	if err = proc.Wait(ctx); err != nil {
		err = errors.Wrap(err, "problem encountered while waiting")
		return &OperationOutcome{
			Succuess: false,
			Text:     err.Error(),
		}, err
	}

	return &OperationOutcome{Succuess: true, Text: fmt.Sprintf("'%s' operation complete", id.Value)}, nil
}

func (s *jasperService) Close(ctx context.Context, _ *empty.Empty) (*OperationOutcome, error) {
	if err := s.manager.Close(ctx); err != nil {
		err = errors.Wrap(err, "problem encountered closing service")
		return &OperationOutcome{
			Succuess: false,
			Text:     err.Error(),
		}, err
	}

	return &OperationOutcome{Succuess: true, Text: "service closed"}, nil
}

func (s *jasperService) GetTags(ctx context.Context, id *JasperProcessID) (*ProcessTags, error) {
	proc, err := s.manager.Get(ctx, id.Value)
	if err != nil {
		return nil, errors.Wrapf(err, "problem finding process '%s'", id.Value)
	}

	return &ProcessTags{ProcessID: id.Value, Tags: proc.GetTags()}, nil
}

func (s *jasperService) TagProcess(ctx context.Context, tags *ProcessTags) (*OperationOutcome, error) {
	proc, err := s.manager.Get(ctx, tags.ProcessID)
	if err != nil {
		err = errors.Wrapf(err, "problem finding process '%s'", tags.ProcessID)
		return &OperationOutcome{
			Succuess: false,
			Text:     err.Error(),
		}, err
	}

	for _, t := range tags.Tags {
		proc.Tag(t)
	}

	return &OperationOutcome{
		Succuess: true,
		Text:     "added tags",
	}, nil
}

func (s *jasperService) ResetTags(ctx context.Context, id *JasperProcessID) (*OperationOutcome, error) {
	proc, err := s.manager.Get(ctx, id.Value)
	if err != nil {
		err = errors.Wrapf(err, "problem finding process '%s'", id.Value)
		return &OperationOutcome{
			Succuess: false,
			Text:     err.Error(),
		}, err
	}
	proc.ResetTags()
	return &OperationOutcome{Succuess: true, Text: "set tags"}, nil
}
