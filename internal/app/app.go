package app

import (
	"context"
	"sync"

	"Coordinator/internal/logger"
)

type App struct {
	services []Service
}

func NewApp(services ...Service) *App {
	return &App{
		services: services,
	}
}

func (a *App) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	logger.App("Starting app: services=%d", len(a.services))

	for _, service := range a.services {
		wg.Add(1)
		go func(srv Service) {
			defer wg.Done()
			if err := srv.Start(ctx); err != nil {
				logger.Error("Service error: %v", err)
			}
		}(service)
	}
	<-ctx.Done()
	logger.Info("App shutting down...")

	wg.Wait()
	logger.Info("All services shut down.")
	return nil
}
