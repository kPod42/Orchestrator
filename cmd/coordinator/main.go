package main

import (
	"Orch/internal/coordinator/app"
	"Orch/internal/coordinator/handler"
	"Orch/internal/coordinator/registry"
	service2 "Orch/internal/coordinator/service"
	httptransport "Orch/internal/coordinator/transport/http"
	"Orch/pkg/logger"
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
	httpSrv := service2.NewHTTPServer(&http.Server{
		Addr:    httpAddr,
		Handler: httptransport.NewRouter(httpHandler),
	})
	presenceSvc := service2.NewPresenceService(reg)
	grpcSrv := service2.NewGRPCServer(grpcListenAddr, presenceSvc)
	application := app.NewApp(
		httpSrv,
		grpcSrv,
	)
	if err := application.Run(ctx); err != nil {
		logger.Log("ERROR", "failed to start application", "err", err)
	}

}
