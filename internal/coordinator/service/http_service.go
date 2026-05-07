package service

import (
	"Orch/pkg/logger"
	"context"
	"errors"
	"net/http"
	"time"
)

type HttpServer struct {
	server *http.Server
}

func NewHTTPServer(server *http.Server) *HttpServer {
	return &HttpServer{
		server: server,
	}
}

func (s *HttpServer) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		logger.Log("INFO", "NET", "http server listening on %s", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		logger.Log("INFO", "HTTP", "Shutting down HTTP server")
		return s.server.Shutdown(shutdownCtx)

	case err := <-errCh:
		logger.Log("ERROR", "HTTP", "Failed to start HTTP server: %v", err)
		return err

	}
}
