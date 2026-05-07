package main

import (
	"Orch/internal/coordinator/app"
	"Orch/internal/coordinator/handler"
	"Orch/internal/coordinator/model"
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
	clusterID := "home-lab"
	coordinatorID := "main"
	configVersion := 1

	coordinatorEndpoints := []model.Endpoint{
		{
			Name:     "local",
			Type:     "grpc",
			Address:  "127.0.0.1:9090",
			Scope:    "same-host",
			Priority: 10,
		},
		{
			Name:     "vmware",
			Type:     "grpc",
			Address:  "192.168.159.1:9090",
			Scope:    "vmware-vmnet",
			Priority: 20,
		},
		{
			Name:     "lan",
			Type:     "grpc",
			Address:  "192.168.50.29:9090",
			Scope:    "lan",
			Priority: 30,
		},
	}

	reg := registry.NewMemoryRegistry(
		clusterID,
		coordinatorID,
		configVersion,
		coordinatorEndpoints,
	)
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
