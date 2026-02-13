package registry

import (
	"time"

	"Coordinator/internal/model"
)

type Registry interface {
	Register(node model.Node) error
	Heartbeat(nodeID string) error
	GetActive() []model.Node
	RemoveStale(timeout time.Duration) error
}
