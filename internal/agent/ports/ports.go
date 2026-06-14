package ports

import (
	"context"
	"time"

	"Orch/internal/agent/config"
	"Orch/internal/agent/model"
)

type Executor interface {
	RunAction(ctx context.Context, action string, args map[string]string) (<-chan model.Event, error)
	RunExec(ctx context.Context, shell, command string, timeout time.Duration) (<-chan model.Event, error)
}

type Policy interface {
	CheckExec(shell string) error
	EffectiveExecTimeout(requestedSec int32) time.Duration
}

type PresenceReporter interface {
	SetBusy(bool)
	SetEndpoints([]config.Endpoint)
}
