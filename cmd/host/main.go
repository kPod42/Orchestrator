package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Orch/internal/host/client"
	"Orch/internal/host/config"
	"Orch/internal/host/core"
	"Orch/internal/host/parser"
	consoleui "Orch/internal/host/ui/console"
	"Orch/pkg/logger"
)

func main() {
	configPath := flag.String("config", "configs/host.yaml", "path to host config")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Log("ERROR", "HOST", "failed to load config: %v", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	requestTimeout := time.Duration(cfg.Host.RequestTimeoutSec) * time.Second

	coordinatorClient := client.NewCoordinatorClient(
		cfg.Host.CoordinatorBaseUrl,
		requestTimeout,
	)

	hostCore := core.NewService(coordinatorClient)
	consoleParser := parser.NewParser()

	console := consoleui.New(
		os.Stdin,
		os.Stdout,
		os.Stderr,
		consoleParser,
		hostCore,
		requestTimeout,
	)

	if err := console.Run(ctx); err != nil {
		logger.Log("ERROR", "HOST", "application stopped with error: %v", err)
		os.Exit(1)
	}
}
