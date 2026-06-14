package action

import (
	"context"

	"Orch/internal/agent/model"
)

type Action interface {
	Name() string
	Description() string
	Run(ctx context.Context, args map[string]string) (<-chan model.Event, error)
}

type Info struct {
	Name        string
	Description string
}

type KnownActions interface {
	HasAction(name string) bool
}
