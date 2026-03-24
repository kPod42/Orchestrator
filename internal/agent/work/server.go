package work

import (
	"context"
	"net"
	"sync"

	"Orch/gen/go/workpb"
	"Orch/internal/agent/executor"
	"Orch/internal/agent/model"
	"Orch/internal/agent/security"
	"Orch/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BusyReporter interface {
	SetBusy(bool)
}

type Server struct {
	address    string
	grpcServer *grpc.Server
	readyCh    chan struct{}
	readyOnce  sync.Once
}

type serviceImpl struct {
	workpb.UnimplementedAgentWorkServiceServer

	executor *executor.Executor
	policy   *security.Policy
	reporter BusyReporter

	mu   sync.Mutex
	busy bool
}

func NewServer(address string, exec *executor.Executor, policy *security.Policy, reporter BusyReporter) *Server {
	svc := &serviceImpl{
		executor: exec,
		policy:   policy,
		reporter: reporter,
	}

	grpcServer := grpc.NewServer()
	workpb.RegisterAgentWorkServiceServer(grpcServer, svc)

	return &Server{
		address:    address,
		grpcServer: grpcServer,
		readyCh:    make(chan struct{}),
	}
}

func (s *Server) Ready() <-chan struct{} {
	return s.readyCh
}

func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	s.readyOnce.Do(func() {
		close(s.readyCh)
	})

	errCh := make(chan error, 1)

	go func() {
		logger.Log("NET", "agent work grpc server listening on %s", s.address)
		errCh <- s.grpcServer.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		logger.Log("NET", "agent work grpc server shutting down", s.address)
		s.grpcServer.GracefulStop()
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *serviceImpl) Action(req *workpb.ActionRequest, stream workpb.AgentWorkService_ActionServer) error {
	logger.Log("WORK", "received Action RPC: action = %s", req.Action)

	if !s.occupy() {
		return status.Error(codes.FailedPrecondition, "agent is busy")
	}
	defer s.release()

	if err := s.policy.CheckActions(req.Action); err != nil {
		return status.Error(codes.PermissionDenied, err.Error())
	}

	s.reporter.SetBusy(true)
	defer s.reporter.SetBusy(false)

	events, err := s.executor.RunAction(stream.Context(), req.Action, req.Args)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	for ev := range events {
		if err := stream.Send(toActionEvent(ev)); err != nil {
			return err
		}
	}

	logger.Log("WORK", "finished Action RPC: action = %s", req.Action)
	return nil
}

func (s *serviceImpl) Exec(req *workpb.ExecRequest, stream workpb.AgentWorkService_ExecServer) error {
	logger.Log("WORK", "received Exec RPC: shell = %s command = %s", req.Shell, req.Command)

	if !s.occupy() {
		return status.Error(codes.FailedPrecondition, "agent is busy")
	}
	defer s.release()

	if err := s.policy.CheckExec(req.Shell); err != nil {
		return status.Error(codes.PermissionDenied, err.Error())
	}

	timeout := s.policy.EffectiveExecTimeoutSec(req.TimeoutSec)

	s.reporter.SetBusy(true)
	defer s.reporter.SetBusy(false)

	events, err := s.executor.RunExec(stream.Context(), req.Shell, req.Command, timeout)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	for ev := range events {
		if err := stream.Send(toExecEvent(ev)); err != nil {
			return err
		}
	}

	logger.Log("WORK", "finished Exec RPC", "shell = %s command = %s", req.Shell, req.Command)
	return nil
}

func (s *serviceImpl) occupy() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.busy {
		return false
	}

	s.busy = true
	return true
}

func (s *serviceImpl) release() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.busy = false
}

func toActionEvent(ev model.Event) *workpb.ActionEvent {
	if ev.Output != nil {
		return &workpb.ActionEvent{
			Payload: &workpb.ActionEvent_Output{
				Output: &workpb.OutputChunk{
					Stream: ev.Output.Stream,
					Chunk:  ev.Output.Chunk,
				},
			},
		}
	}

	return &workpb.ActionEvent{
		Payload: &workpb.ActionEvent_Result{
			Result: &workpb.CommandResult{
				Success:  ev.Result.Success,
				ExitCode: ev.Result.ExitCode,
				Message:  ev.Result.Message,
			},
		},
	}
}

func toExecEvent(ev model.Event) *workpb.ExecEvent {
	if ev.Output != nil {
		return &workpb.ExecEvent{
			Payload: &workpb.ExecEvent_Output{
				Output: &workpb.OutputChunk{
					Stream: ev.Output.Stream,
					Chunk:  ev.Output.Chunk,
				},
			},
		}
	}

	return &workpb.ExecEvent{
		Payload: &workpb.ExecEvent_Result{
			Result: &workpb.CommandResult{
				Success:  ev.Result.Success,
				ExitCode: ev.Result.ExitCode,
				Message:  ev.Result.Message,
			},
		},
	}
}
