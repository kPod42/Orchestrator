package service

import (
	"context"
	"time"

	"Coordinator/internal/logger"
	"Coordinator/internal/registry"
)

type CleanupService struct {
	reg      registry.Registry
	interval time.Duration
	ttl      time.Duration
}

func NewCleanupService(reg registry.Registry, interval, ttl time.Duration) *CleanupService {
	return &CleanupService{
		reg:      reg,
		interval: interval,
		ttl:      ttl,
	}
}

func (s *CleanupService) Start(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	logger.Info("Cleanup service started")

	for {
		select {
		case <-ticker.C:
			err := s.reg.RemoveStale(s.ttl)
			if err != nil {
				logger.Error("Failed to remove stale cleanup service")
			}
		case <-ctx.Done():
			logger.Info("Cleanup service stopped")
			return nil
		}
	}
}
