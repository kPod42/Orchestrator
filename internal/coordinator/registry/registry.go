package registry

import (
	"Orch/internal/coordinator/model"
)

type Registry interface {
	Register(node model.Node) (model.RegisterResponse, error)
	GetActive() []model.Node
	GetCoordinatorInfo() model.CoordinatorInfo
	Attach(nodeID, sessionID string) error
	Detach(nodeID string, sessionID string)
	UpdateStatus(nodeID, sessionID string, busy bool) error
	UpdateEndpoints(nodeID, sessionID string, endpoints []model.Endpoint) error
}
