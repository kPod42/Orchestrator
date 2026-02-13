package registry

import (
	"Coordinator/internal/model"
	"testing"
	"time"
)

func TestRegistry(t *testing.T) {
	reg := NewInMemoryRegistry()

	node := model.Node{
		ID:   "agent-1",
		IP:   "127.0.0.1",
		Port: 8030,
	}

	err := reg.Register(node)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	nodes := reg.GetActive()
	if len(nodes) != 1 {
		t.Fatalf("GetActive failed: %v", nodes)
	}
}
func TestHeartbeat(t *testing.T) {
	reg := NewInMemoryRegistry()
	node := model.Node{
		ID: "agent-1",
	}
	reg.Register(node)

	err := reg.Heartbeat("agent-1")
	if err != nil {
		t.Fatalf("heartbeat failed: %v", err)
	}
}

func TestRemoveNode(t *testing.T) {
	reg := NewInMemoryRegistry()
	node := model.Node{
		ID:       "agent-1",
		LastSeen: time.Now().Add(-1 * time.Minute),
	}
	reg.nodes[node.ID] = &node
	err := reg.RemoveStale(30 * time.Second)
	if err != nil {
		return
	}
	if len(reg.nodes) != 0 {
		t.Fatalf("Remove failed: %v", reg.nodes)
	}
}
