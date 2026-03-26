package work

import (
	"context"
	"net"
	"sync"

	"Orch/gen/go/workpb"
	"Orch/internal/agent/model"
	"Orch/internal/agent/ports"
	"Orch/internal/agent/state"
	"Orch/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	address    string
	grpcServer *grpc.Server
	readyCh    chan struct{}
	readyOnce  sync.Once
}

type serviceImpl struct {
	workpb.UnimplementedAgentWorkServiceServer

	executor ports.Executor
	policy   ports.Policy
	reporter ports.PresenceReporter
	busy     *state.Busy
}

func NewServer(
	address string,
	exec ports.Executor,
	policy ports.Policy,
	reporter ports.PresenceReporter,
	busy *state.Busy,
) *Server {
	svc := &serviceImpl{
		executor: exec,
		policy:   policy,
		reporter: reporter,
		busy:     busy,
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
func (s *Server) Name() string {
	return "work"
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
		logger.Log("INFO", "NET", "agent work grpc server listening on %s", s.address)
		errCh <- s.grpcServer.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		logger.Log("INFO", "NET", "agent work grpc server shutting down")
		s.grpcServer.GracefulStop()
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *serviceImpl) Action(req *workpb.ActionRequest, stream workpb.AgentWorkService_ActionServer) error {
	logger.Log("INFO", "WORK", "received Action RPC: action = %s", req.Action)

	return s.run(
		stream.Context(),
		func() error {
			return s.policy.CheckAction(req.Action)
		},
		func(ctx context.Context) (<-chan model.Event, error) {
			return s.executor.RunAction(ctx, req.Action, req.Args)
		},
		func(ev model.Event) error {
			return stream.Send(toActionEvent(ev))
		},
		func() {
			logger.Log("INFO", "WORK", "finished Action RPC: action = %s", req.Action)
		},
	)
}

func (s *serviceImpl) Exec(req *workpb.ExecRequest, stream workpb.AgentWorkService_ExecServer) error {
	logger.Log("INFO", "WORK", "received Exec RPC: shell = %s command = %s", req.Shell, req.Command)

	timeout := s.policy.EffectiveExecTimeout(req.TimeoutSec)

	return s.run(
		stream.Context(),
		func() error {
			return s.policy.CheckExec(req.Shell)
		},
		func(ctx context.Context) (<-chan model.Event, error) {
			return s.executor.RunExec(ctx, req.Shell, req.Command, timeout)
		},
		func(ev model.Event) error {
			return stream.Send(toExecEvent(ev))
		},
		func() {
			logger.Log("INFO", "WORK", "finished Exec RPC: shell = %s command = %s", req.Shell, req.Command)
		},
	)
}

func (s *serviceImpl) run(
	ctx context.Context,
	prepare func() error,
	start func(context.Context) (<-chan model.Event, error),
	send func(model.Event) error,
	onDone func(),
) error {
	if !s.busy.Acquire() {
		return status.Error(codes.FailedPrecondition, "agent is busy")
	}
	defer s.busy.Release()

	if err := prepare(); err != nil {
		return status.Error(codes.PermissionDenied, err.Error())
	}

	s.reporter.SetBusy(true)
	defer s.reporter.SetBusy(false)

	events, err := start(ctx)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	for ev := range events {
		if err := send(ev); err != nil {
			return err
		}
	}

	onDone()
	return nil
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
