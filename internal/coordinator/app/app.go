package app

import (
	"Orch/pkg/logger"
	"context"
	"sync"
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

	logger.Log("INFO", "APP", "Starting app: services = %d", len(a.services))

	for _, service := range a.services {
		wg.Add(1)
		go func(srv Service) {
			defer wg.Done()
			if err := srv.Start(ctx); err != nil {
				logger.Log("ERROR", "SERVICE", err.Error())
			}
		}(service)
	}
	<-ctx.Done()
	logger.Log("INFO", "APP", "Shutting down app")

	wg.Wait()
	logger.Log("INFO", "APP", "Shut down app successfully")
	return nil
}
