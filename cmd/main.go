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
	"time"
)

func main() {
	reg := registry.NewInMemoryRegistry()
	h := handler.NewHTTPHandler(reg)
	router := httptransport.NewRouter(h)
	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: router,
	}

	httpService := service.NewHTTPServer(server)
	cleanupService := service.NewCleanupService(reg, 10*time.Second, 30*time.Second)

	application := app.NewApp(httpService, cleanupService)

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	if err := application.Run(ctx); err != nil {
		logger.Error("Failed to start HTTP server")
	}
}
