package registry

import (
	"Coordinator/internal/logger"
	"errors"
	"sync"
	"time"

	"Coordinator/internal/model"
)

type InMemoryRegistry struct {
	mu    sync.RWMutex
	nodes map[string]*model.Node
}

func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		nodes: make(map[string]*model.Node),
	}
}

func (r *InMemoryRegistry) Register(node model.Node) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	node.LastSeen = time.Now()
	r.nodes[node.ID] = &node
	logger.Info("Node registered: %s (%s:%d)",
		node.ID, node.IP, node.Port)
	return nil
}

func (r *InMemoryRegistry) Heartbeat(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	node, exists := r.nodes[id]
	if !exists {
		return errors.New("node does not exist")
	}
	node.LastSeen = time.Now()
	return nil
}

func (r *InMemoryRegistry) GetActive() []model.Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []model.Node
	for _, node := range r.nodes {
		result = append(result, *node)
	}
	return result
}

func (r *InMemoryRegistry) RemoveStale(timeout time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	removed := 0
	now := time.Now()

	for _, node := range r.nodes {
		if now.Sub(node.LastSeen) > timeout {
			delete(r.nodes, node.ID)
			removed++
		}
	}

	if removed > 0 {
		logger.Info("Removed %d stale nodes", removed)
	}
	return nil
}
