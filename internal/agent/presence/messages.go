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

func (c *Client) sendSnapshot(stream presencepb.PresenceService_ConnectClient, snap snapshot) error {
	if err := stream.Send(endpointMessage(snap.endpoints)); err != nil {
		return fmt.Errorf("send endpoints update: %w", err)
	}

	if err := stream.Send(statusMessage(snap.busy)); err != nil {
		return fmt.Errorf("send busy update: %w", err)
	}

	return nil
}
