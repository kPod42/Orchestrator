package registry

import (
	"Orch/internal/coordinator/model"
	"Orch/pkg/logger"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
)

type nodeRecord struct {
	node      model.Node
	sessionID string
	active    bool
}

type MemoryRegistry struct {
	mutex       sync.RWMutex
	nodes       map[string]*nodeRecord
	grpcAddress string
}

func NewMemoryRegistry(grpcAddress string) *MemoryRegistry {
	return &MemoryRegistry{
		nodes:       make(map[string]*nodeRecord),
		grpcAddress: grpcAddress,
	}
}

func (m *MemoryRegistry) Register(node model.Node) (model.RegisterResponse, error) {
	if node.ID == "" {
		return model.RegisterResponse{}, errors.New("node ID can`t be empty")
	}

	sessionID, err := newSessionID()
	if err != nil {
		return model.RegisterResponse{}, err
	}
	node.Busy = false

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.nodes[node.ID] = &nodeRecord{
		node:      node,
		sessionID: sessionID,
		active:    false,
	}

	logger.Log("INFO", "MEMORY", "Registered node: nodeID = %s sessionID = %s", node.ID, sessionID)

	return model.RegisterResponse{
		NodeID:      node.ID,
		SessionID:   sessionID,
		GRPCAddress: m.grpcAddress,
	}, nil
}

func (m *MemoryRegistry) Attach(nodeID, sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	rec, ok := m.nodes[nodeID]
	if !ok {
		return errors.New("node not registered")
	}
	if rec.sessionID != sessionID {
		return errors.New("invalid sessionID")
	}

	rec.active = true
	logger.Log("INFO", "MEMORY", "Attached node: nodeID = %s sessionID = %s", nodeID, sessionID)
	return nil
}

func (m *MemoryRegistry) Detach(nodeID string, sessionID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	rec, ok := m.nodes[nodeID]
	if !ok {
		return
	}
	if rec.sessionID != sessionID {
		return
	}
	rec.active = false
	rec.node.Busy = false
	logger.Log("INFO", "MEMORY", "Detached node: nodeID = %s sessionID = %s", nodeID, sessionID)
}

func (m *MemoryRegistry) UpdateStatus(nodeID, sessionID string, busy bool) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	rec, ok := m.nodes[nodeID]
	if !ok {
		return errors.New("node not registered")
	}
	if rec.sessionID != sessionID {
		return errors.New("invalid sessionID")
	}
	if !rec.active {
		return errors.New("node is not active")
	}

	rec.node.Busy = busy
	logger.Log("INFO", "MEMORY", "Updated node status: nodeID = %s sessionID = %s busy = %v", nodeID, sessionID, busy)
	return nil
}

func (m *MemoryRegistry) UpdateEndpoints(nodeID, sessionID string, endpoints []model.Endpoint) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	rec, ok := m.nodes[nodeID]
	if !ok {
		return errors.New("node not registered")
	}
	if rec.sessionID != sessionID {
		return errors.New("invalid sessionID")
	}
	rec.node.Endpoints = endpoints
	logger.Log("INFO", "MEMORY", "Updated node endpoints: nodeID = %s sessionID = %s endpoints = %v", nodeID, sessionID, endpoints)
	return nil
}

func (m *MemoryRegistry) GetActive() []model.Node {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make([]model.Node, 0, len(m.nodes))
	for _, rec := range m.nodes {
		if rec.active {
			result = append(result, rec.node)
		}
	}
	logger.Log("INFO", "MEMORY", "GetActive result = %d", result, len(result))
	return result
}

func newSessionID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes[:]), nil
}
