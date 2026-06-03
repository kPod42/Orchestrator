package core

import (
	"context"

	"Orch/internal/host/model"
)

type CoordinatorGateway interface {
	GetNodes(ctx context.Context) ([]model.Node, error)
	Execute(ctx context.Context, request model.ExecuteRequest) (model.ExecuteResponse, error)
}
