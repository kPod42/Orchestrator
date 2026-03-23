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
	go func() {
		logger.Log("INFO", "NET", "http server listening on %s", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log("ERROR", "NET", "http server failed to start: %v", err)
		}
	}()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Log("INFO", "NET", "http server shutting down")
	return s.server.Shutdown(shutdownCtx)
}
