package service

import (
	"Orch/gen/go/presencepb"
	"Orch/pkg/logger"
	"context"
	"net"

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
		s.server.Stop()
		return nil
	case err := <-errCh:
		return err
	}
}
