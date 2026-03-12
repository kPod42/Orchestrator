package registry

import (
	"Coordinator/internal/logger"
	"Coordinator/internal/model"
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

type memoryRegistry struct {
	mutex       sync.RWMutex
	nodes       map[string]*nodeRecord
	grpcAddress string
}

func NewMemoryRegistry(grpcAddress string) *memoryRegistry {
	return &memoryRegistry{
		nodes:       make(map[string]*nodeRecord),
		grpcAddress: grpcAddress,
	}
}

func (m *memoryRegistry) Register(node model.Node) (model.RegisterResponse, error) {
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

	logger.Registry("register: nodeId=%s sessionId=%s", node.ID, sessionID)

	return model.RegisterResponse{
		NodeID:      node.ID,
		SessionID:   sessionID,
		GRPCAddress: m.grpcAddress,
	}, nil
}

func (m *memoryRegistry) Attach(nodeID, sessionID string) error {
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
	logger.Registry("attach: nodeId=%s active=true", nodeID)
	return nil
}

func (m *memoryRegistry) Detach(nodeID string, sessionID string) {
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
	logger.Registry("detach: nodeId=%s active=false busy=false", nodeID)
}

func (m *memoryRegistry) UpdateStatus(nodeID, sessionID string, busy bool) error {
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
	logger.Registry("status update: nodeId=%s busy=%v", nodeID, busy)
	return nil
}

func (m *memoryRegistry) UpdateEndpoints(nodeID, sessionID string, endpoints []model.Endpoint) error {
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
	logger.Registry("endpoints update: nodeId=%s endpoints=%v", nodeID, endpoints)
	return nil
}

func (m *memoryRegistry) GetActive() []model.Node {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make([]model.Node, 0, len(m.nodes))
	for _, rec := range m.nodes {
		if rec.active {
			result = append(result, rec.node)
		}
	}
	logger.Registry("get active: count=%d", len(result))
	return result
}

func newSessionID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes[:]), nil
}
