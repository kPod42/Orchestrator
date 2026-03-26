package app

import (
	"context"
	"errors"
	"sync"

	"Orch/internal/agent/ports"
	"Orch/pkg/logger"
)

type App struct {
	services []ports.Service
}

func New(services ...ports.Service) *App {
	return &App{
		services: services,
	}
}

func (a *App) Run(ctx context.Context) error {
	logger.Log("INFO", "APP", "Starting app: services = %d", len(a.services))

	errCh := make(chan error, len(a.services))
	var wg sync.WaitGroup

	for _, svc := range a.services {
		wg.Add(1)

		go func(service ports.Service) {
			defer wg.Done()

			if err := service.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
				errCh <- err
			}
		}(svc)

		select {
		case <-svc.Ready():
			logger.Log("INFO", "APP", "Service is ready: %s", svc.Name())

		case err := <-errCh:
			return err

		case <-ctx.Done():
			wg.Wait()
			return nil
		}
	}

	select {
	case err := <-errCh:
		return err

	case <-ctx.Done():
		logger.Log("INFO", "APP", "Shutdown signal received")
		wg.Wait()
		return nil
	}
}
