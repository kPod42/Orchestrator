package service

import (
	"context"
	"net"

	pb "Coordinator/internal/transport/grpc/pb"

	"Coordinator/internal/logger"

	"google.golang.org/grpc"
)

type GRPCServer struct {
	server *grpc.Server
	addr   string
}

func NewGRPCServer(addr string, service pb.PresenceServiceServer) *GRPCServer {
	srv := grpc.NewServer()
	pb.RegisterPresenceServiceServer(srv, service)
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
		logger.Info("grpc server listening on %s", s.addr)
		errCh <- s.server.Serve(lis)
	}()
	select {
	case <-ctx.Done():
		logger.Info("grpc server shutting down")
		s.server.Stop()
		return nil
	case err := <-errCh:
		return err
	}
}
