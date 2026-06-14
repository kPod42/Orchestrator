package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"Orch/internal/agent/app"
	"Orch/internal/agent/config"
	"Orch/internal/agent/executor"
	"Orch/internal/agent/presence"
	"Orch/internal/agent/security"
	"Orch/internal/agent/state"
	"Orch/internal/agent/work"
	"Orch/pkg/logger"
)

func main() {
	configPath := flag.String("config", "configs/agent.dev.yaml", "path to agent config")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Log("ERROR", "AGENT", "failed to load config: %v", strings.ToTitle(err.Error()))
		os.Exit(1)
	}

	logger.Log("INFO", "AGENT", "starting agent: nodeID = %s", cfg.Agent.NodeID)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	policy := security.New(cfg.Security)
	exec := executor.New(policy)
	busyState := state.NewBusy()

	presenceClient := presence.New(
		cfg,
		exec,
		policy,
		busyState,
	)

	workServer := work.NewServer(
		cfg.Work.ListenAddress,
		exec,
		policy,
		presenceClient,
		busyState,
	)

	application := app.New(
		workServer,
		presenceClient,
	)

	if err := application.Run(ctx); err != nil {
		logger.Log("ERROR", "AGENT", "application stopped with error: %v", err)
		os.Exit(1)
	}

	logger.Log("INFO", "AGENT", "agent stopped")
}
