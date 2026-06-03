package presence

import (
	"fmt"

	"Orch/gen/go/presencepb"
	"Orch/internal/agent/config"
)

func connectMessage(nodeID, sessionID string) *presencepb.AgentPresenceMessage {
	return &presencepb.AgentPresenceMessage{
		Payload: &presencepb.AgentPresenceMessage_Connect{
			Connect: &presencepb.ConnectRequest{
				NodeId:    nodeID,
				SessionId: sessionID,
			},
		},
	}
}

func statusMessage(busy bool) *presencepb.AgentPresenceMessage {
	return &presencepb.AgentPresenceMessage{
		Payload: &presencepb.AgentPresenceMessage_Status{
			Status: &presencepb.StatusUpdate{
				Busy: busy,
			},
		},
	}
}

func endpointMessage(endpoints []config.Endpoint) *presencepb.AgentPresenceMessage {
	pbEndpoints := make([]*presencepb.Endpoint, 0, len(endpoints))

	for _, ep := range endpoints {
		pbEndpoints = append(pbEndpoints, &presencepb.Endpoint{
			Kind:    ep.Kind,
			Address: ep.Address,
		})
	}

	return &presencepb.AgentPresenceMessage{
		Payload: &presencepb.AgentPresenceMessage_EndpointUpdate{
			EndpointUpdate: &presencepb.EndpointUpdate{
				Endpoints: pbEndpoints,
			},
		},
	}
}

func taskOutputMessage(
	taskID string,
	stream string,
	chunk string,
) *presencepb.AgentPresenceMessage {
	return &presencepb.AgentPresenceMessage{
		Payload: &presencepb.AgentPresenceMessage_TaskOutput{
			TaskOutput: &presencepb.TaskOutput{
				TaskId: taskID,
				Stream: stream,
				Chunk:  chunk,
			},
		},
	}
}

func taskResultMessage(
	taskID string,
	success bool,
	exitCode int32,
	message string,
) *presencepb.AgentPresenceMessage {
	return &presencepb.AgentPresenceMessage{
		Payload: &presencepb.AgentPresenceMessage_TaskResult{
			TaskResult: &presencepb.TaskResult{
				TaskId:   taskID,
				Success:  success,
				ExitCode: exitCode,
				Message:  message,
			},
		},
	}
}

type sendPresenceMessageFunc func(*presencepb.AgentPresenceMessage) error

func (c *Client) sendSnapshot(send sendPresenceMessageFunc, snap snapshot) error {
	if err := send(endpointMessage(snap.endpoints)); err != nil {
		return fmt.Errorf("send endpoints update: %w", err)
	}

	if err := send(statusMessage(snap.busy)); err != nil {
		return fmt.Errorf("send busy update: %w", err)
	}

	return nil
}
