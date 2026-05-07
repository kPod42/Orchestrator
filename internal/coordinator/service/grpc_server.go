package service

import (
	"Orch/gen/go/presencepb"
	"Orch/pkg/logger"
	"context"
	"net"
	"time"

	"google.golang.org/grpc"
)

type GRPCServer struct {
	server *grpc.Server
	addr   string
}

func NewGRPCServer(addr string, service presencepb.PresenceServiceServer) *GRPCServer {
	srv := grpc.NewServer()
	presencepb.RegisterPresenceServiceServer(srv, service)
	return &GRPCServer{
		server: srv,
		addr:   addr,
	}
}
func (s *GRPCServer) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	errCh := make(chan error, 1)

	go func() {
		logger.Log("INFO", "NET", "grpc server listening on %s", s.addr)
		errCh <- s.server.Serve(lis)
	}()
	select {
	case <-ctx.Done():
		logger.Log("INFO", "NET", "grpc server shutting down")
		stopped := make(chan struct{})
		go func() {
			s.server.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
			logger.Log("INFO", "NET", "grpc server stopped gracefully")
		case <-time.After(5 * time.Second):
			logger.Log("WARNING", "NET", "grpc graceful stop timeout, forcing stop")
			s.server.Stop()
		}
		return nil
	case err := <-errCh:
		return err
	}
}
