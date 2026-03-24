package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"Orch/internal/agent/config"
	"Orch/internal/agent/executor"
	"Orch/internal/agent/presence"
	"Orch/internal/agent/security"
	"Orch/internal/agent/work"
	"Orch/pkg/logger"
)

func main() {
	configPath := flag.String("config", "configs/agent.dev.json", "path to agent config")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Log("ERROR", "AGENT", "failed to load config: %v", strings.ToTitle(err.Error()))
		os.Exit(1)
	}

	logger.Log("INFO", "AGENT", "starting agent: nodeID = %s", cfg.Agent.NodeID)

	baseCtx, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignal()

	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	exec := executor.New()
	policy := security.New(cfg.Security)
	presenceClient := presence.New(cfg)
	workServer := work.NewServer(cfg.Work.ListenAddress, exec, policy, presenceClient)

	errCh := make(chan error, 2)
	var wg sync.WaitGroup

	// 1. Start work server first.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := workServer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			errCh <- err
		}
	}()

	// Wait until work server is actually listening.
	select {
	case <-workServer.Ready():
		logger.Log("NET", "AGENT", "work server is ready, starting presence client")
	case err := <-errCh:
		logger.Log("ERROR", "AGENT", "startup failed: %v", err)
		cancel()
		wg.Wait()
		os.Exit(1)
	case <-ctx.Done():
		wg.Wait()
		return
	}

	// 2. Start presence client after work server is ready.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := presenceClient.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			errCh <- err
		}
	}()

	// 3. Wait for shutdown or service failure.
	select {
	case err := <-errCh:
		logger.Log("INFO", "AGENT", "service error: %v", err)
		cancel()
	case <-ctx.Done():
		logger.Log("INFO", "AGENT", "shutdown signal received")
	}

	wg.Wait()
	logger.Log("INFO", "AGENT", "agent stopped")
}
