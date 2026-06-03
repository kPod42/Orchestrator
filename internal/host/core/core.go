package core

import (
	"context"
	"fmt"
	"strings"

	"Orch/internal/host/model"
)

type Service struct {
	coordinator CoordinatorGateway
}

func NewService(coordinator CoordinatorGateway) *Service {
	return &Service{coordinator: coordinator}
}

func (s *Service) GetNodes(ctx context.Context) ([]model.Node, error) {
	return s.coordinator.GetNodes(ctx)
}
func (s *Service) ExecuteAction(
	ctx context.Context, action string, targets []string, args map[string]string,
) (model.ExecuteResponse, error) {
	action = strings.TrimSpace(action)
	if action == "" {
		return model.ExecuteResponse{}, fmt.Errorf("action is empty")
	}
	if len(targets) == 0 {
		return model.ExecuteResponse{}, fmt.Errorf("targets are empty")
	}
	if args == nil {
		args = make(map[string]string)
	}
	request := model.ExecuteRequest{
		Mode:    "action",
		Action:  action,
		Targets: targets,
		Args:    args,
	}
	return s.coordinator.Execute(ctx, request)
}
