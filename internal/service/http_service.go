package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"Coordinator/internal/logger"
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
	go func() {
		logger.Net("Starting HTTP server on %s", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Error starting HTTP server: %s", err)
		}
	}()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Net("Shutting down HTTP server")
	return s.server.Shutdown(shutdownCtx)
}
