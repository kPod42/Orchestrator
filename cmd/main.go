package main

import (
	"Coordinator/internal/app"
	"Coordinator/internal/handler"
	"Coordinator/internal/logger"
	"Coordinator/internal/registry"
	"Coordinator/internal/service"
	httptransport "Coordinator/internal/transport/http"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpAddr := "0.0.0.0:8080"
	grpcListenAddr := "0.0.0.0:9090"
	grpcPublicAddr := "127.0.0.1:9090"

	reg := registry.NewMemoryRegistry(grpcPublicAddr)
	httpHandler := handler.NewHTTPHandler(reg)
	httpSrv := service.NewHTTPServer(&http.Server{
		Addr:    httpAddr,
		Handler: httptransport.NewRouter(httpHandler),
	})
	presenceSvc := service.NewPresenceService(reg)
	grpcSrv := service.NewGRPCServer(grpcListenAddr, presenceSvc)
	application := app.NewApp(
		httpSrv,
		grpcSrv,
	)
	if err := application.Run(ctx); err != nil {
		logger.Error("application.Run() error: %s", err)
	}

}
