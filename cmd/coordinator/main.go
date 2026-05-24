package main

import (
	"Orch/internal/coordinator/app"
	coordconfig "Orch/internal/coordinator/config"
	"Orch/internal/coordinator/handler"
	"Orch/internal/coordinator/registry"
	service2 "Orch/internal/coordinator/service"
	httptransport "Orch/internal/coordinator/transport/http"
	"Orch/pkg/logger"
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configPath := flag.String("config", "configs/coordinator.yaml", "path to coordinator config file")
	flag.Parse()

	cfg, err := coordconfig.Load(*configPath)
	if err != nil {
		logger.Log("ERROR", "APP", "failed to load coordinator config: %v", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	reg := registry.NewMemoryRegistry(
		cfg.Cluster.ID,
		cfg.Coordinator.ID,
		cfg.Coordinator.ConfigVersion,
		cfg.Coordinator.Endpoints,
	)

	httpHandler := handler.NewHTTPHandler(reg)

	httpSrv := service2.NewHTTPServer(&http.Server{
		Addr:    cfg.Coordinator.HTTP.ListenAddr,
		Handler: httptransport.NewRouter(httpHandler),
	})

	presenceSvc := service2.NewPresenceService(reg)

	grpcSrv := service2.NewGRPCServer(
		cfg.Coordinator.GRPC.ListenAddr,
		presenceSvc,
	)

	logger.Log("INFO", "APP", "coordinator config loaded: clusterID = %s coordinatorID = %s configVersion = %d",
		cfg.Cluster.ID,
		cfg.Coordinator.ID,
		cfg.Coordinator.ConfigVersion,
	)

	for _, endpoint := range cfg.Coordinator.Endpoints {
		logger.Log("INFO", "APP", "configured coordinator endpoint: name = %s kind = %s address = %s scope = %s priority = %d",
			endpoint.Name,
			endpoint.Type,
			endpoint.Address,
			endpoint.Scope,
			endpoint.Priority,
		)
	}

	application := app.NewApp(
		httpSrv,
		grpcSrv,
	)

	if err := application.Run(ctx); err != nil {
		logger.Log("ERROR", "APP", "failed to start application: %v", err)
	}
}
