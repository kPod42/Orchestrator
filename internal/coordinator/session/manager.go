package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"Orch/gen/go/presencepb"
)

type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
	}
}

func (m *Manager) Attach(nodeID string, sessionID string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	if oldSession, ok := m.sessions[nodeID]; ok {
		oldSession.Close()
	}

	s := newSession(nodeID, sessionID)
	m.sessions[nodeID] = s

	return s
}

func (m *Manager) Detach(nodeID string, sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[nodeID]
	if !ok {
		return
	}

	if s.SessionID() != sessionID {
		return
	}

	s.Close()
	delete(m.sessions, nodeID)
}

func (m *Manager) RunAction(
	ctx context.Context,
	nodeID string,
	action string,
	args map[string]string,
) (RunResult, error) {
	m.mu.RLock()
	s, ok := m.sessions[nodeID]
	m.mu.RUnlock()

	if !ok {
		return RunResult{}, fmt.Errorf("node has no active control session: %s", nodeID)
	}

	return s.RunAction(ctx, action, args)
}

func (m *Manager) HandleAgentMessage(
	nodeID string,
	sessionID string,
	msg *presencepb.AgentPresenceMessage,
) error {
	m.mu.RLock()
	s, ok := m.sessions[nodeID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("node has no active control session: %s", nodeID)
	}

	if s.SessionID() != sessionID {
		return fmt.Errorf("invalid session id for node: %s", nodeID)
	}

	return s.HandleAgentMessage(msg)
}

type RunResult struct {
	Status   string
	ExitCode int32
	Message  string
	Output   string
}

type Session struct {
	nodeID    string
	sessionID string

	sendCh chan *presencepb.CoordinatorPresenceMessage
	doneCh chan struct{}

	closeOnce sync.Once

	mu      sync.Mutex
	pending map[string]*pendingTask
}

type pendingTask struct {
	output strings.Builder
	done   chan RunResult
}

func newSession(nodeID string, sessionID string) *Session {
	return &Session{
		nodeID:    nodeID,
		sessionID: sessionID,
		sendCh:    make(chan *presencepb.CoordinatorPresenceMessage, 64),
		doneCh:    make(chan struct{}),
		pending:   make(map[string]*pendingTask),
	}
}

func (s *Session) NodeID() string {
	return s.nodeID
}

func (s *Session) SessionID() string {
	return s.sessionID
}

func (s *Session) SendChannel() <-chan *presencepb.CoordinatorPresenceMessage {
	return s.sendCh
}

func (s *Session) Done() <-chan struct{} {
	return s.doneCh
}

func (s *Session) Close() {
	s.closeOnce.Do(func() {
		close(s.doneCh)
	})
}

func (s *Session) RunAction(
	ctx context.Context,
	action string,
	args map[string]string,
) (RunResult, error) {
	action = strings.TrimSpace(action)
	if action == "" {
		return RunResult{}, fmt.Errorf("action is empty")
	}

	if args == nil {
		args = make(map[string]string)
	}

	taskID, err := newTaskID()
	if err != nil {
		return RunResult{}, err
	}

	task := &pendingTask{
		done: make(chan RunResult, 1),
	}

	s.mu.Lock()
	s.pending[taskID] = task
	s.mu.Unlock()

	defer s.removePending(taskID)

	msg := &presencepb.CoordinatorPresenceMessage{
		Payload: &presencepb.CoordinatorPresenceMessage_TaskRequest{
			TaskRequest: &presencepb.TaskRequest{
				TaskId: taskID,
				Action: action,
				Args:   args,
			},
		},
	}

	select {
	case s.sendCh <- msg:
	case <-ctx.Done():
		return RunResult{}, ctx.Err()
	case <-s.doneCh:
		return RunResult{}, fmt.Errorf("control session closed: nodeID = %s", s.nodeID)
	}

	select {
	case result := <-task.done:
		return result, nil

	case <-ctx.Done():
		return RunResult{}, ctx.Err()

	case <-s.doneCh:
		return RunResult{}, fmt.Errorf("control session closed: nodeID = %s", s.nodeID)
	}
}

func (s *Session) HandleAgentMessage(msg *presencepb.AgentPresenceMessage) error {
	switch payload := msg.Payload.(type) {
	case *presencepb.AgentPresenceMessage_TaskOutput:
		s.appendTaskOutput(
			payload.TaskOutput.TaskId,
			payload.TaskOutput.Stream,
			payload.TaskOutput.Chunk,
		)
		return nil

	case *presencepb.AgentPresenceMessage_TaskResult:
		status := "completed"
		if !payload.TaskResult.Success {
			status = "error"
		}

		s.completeTask(
			payload.TaskResult.TaskId,
			RunResult{
				Status:   status,
				ExitCode: payload.TaskResult.ExitCode,
				Message:  payload.TaskResult.Message,
			},
		)
		return nil

	default:
		return nil
	}
}

func (s *Session) appendTaskOutput(taskID string, stream string, chunk string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.pending[taskID]
	if !ok {
		return
	}

	task.output.WriteString(stream)
	task.output.WriteString(": ")
	task.output.WriteString(chunk)
	task.output.WriteString("\n")
}

func (s *Session) completeTask(taskID string, result RunResult) {
	s.mu.Lock()
	task, ok := s.pending[taskID]
	if ok {
		result.Output = task.output.String()
	}
	s.mu.Unlock()

	if !ok {
		return
	}

	select {
	case task.done <- result:
	default:
	}
}

func (s *Session) removePending(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.pending, taskID)
}

func newTaskID() (string, error) {
	var bytes [16]byte

	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate task id: %w", err)
	}

	return hex.EncodeToString(bytes[:]), nil
}
