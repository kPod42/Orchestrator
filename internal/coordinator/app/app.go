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

func (a *App) Run(parentCtx context.Context) error {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, len(a.services))
	logger.Log("INFO", "APP", "Starting app: services = %d", len(a.services))
	for _, service := range a.services {
		wg.Add(1)
		go func(srv Service) {
			defer wg.Done()

			if err := srv.Start(ctx); err != nil {
				logger.Log("ERROR", "APP", "Failed to start service: %v", err)
				select {
				case errCh <- err:
				default:
				}
				cancel()
			}
		}(service)
	}
	select {
	case <-parentCtx.Done():
		logger.Log("INFO", "APP", "Shutdown signal received")
		cancel()
	case err := <-errCh:
		logger.Log("ERROR", "APP", "Application stopping because service failed: %v", err)
		cancel()
		wg.Wait()
		return err
	}
	logger.Log("INFO", "APP", "Shutting down app")
	wg.Wait()
	logger.Log("INFO", "APP", "Application stopped successfully")
	return nil
}
