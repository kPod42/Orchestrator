package presence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"Orch/gen/go/presencepb"
	"Orch/internal/agent/config"
	"Orch/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	cfg    *config.Config
	sendCh chan *presencepb.AgentPresenceMessage
}

type registerRequest struct {
	ID           string            `json:"id"`
	Capabilities []string          `json:"capabilities"`
	Endpoints    []config.Endpoint `json:"endpoints"`
	Busy         bool              `json:"busy"`
}

type registerResponse struct {
	NodeID      string `json:"nodeId"`
	SessionID   string `json:"sessionId"`
	GRPCAddress string `json:"grpcAddress"`
}

func New(cfg *config.Config) *Client {
	return &Client{
		cfg:    cfg,
		sendCh: make(chan *presencepb.AgentPresenceMessage, 16),
	}
}

func (c *Client) Start(ctx context.Context) error {
	logger.Log("INFO", "PRESENCE", "Registred agent: nodeID = %s", c.cfg.Agent.NodeID)

	regResp, err := c.register(ctx)
	if err != nil {
		return err
	}

	grpcAddr := regResp.GRPCAddress
	if grpcAddr == "" {
		grpcAddr = c.cfg.Coordinator.GRPCAddress
	}
	if grpcAddr == "" {
		return fmt.Errorf("presence grpc address is empty")
	}

	conn, err := grpc.DialContext(ctx, grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("dial coordinator presence grpc: %w", err)
	}
	defer conn.Close()

	client := presencepb.NewPresenceServiceClient(conn)

	stream, err := client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("open presence stream: %w", err)
	}

	if err := stream.Send(&presencepb.AgentPresenceMessage{
		Payload: &presencepb.AgentPresenceMessage_Connect{
			Connect: &presencepb.ConnectRequest{
				NodeId:    regResp.NodeID,
				SessionId: regResp.SessionID,
			},
		},
	}); err != nil {
		return fmt.Errorf("send connect message: %w", err)
	}

	ack, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("receive connect ack: %w", err)
	}

	switch payload := ack.Payload.(type) {
	case *presencepb.CoordinatorPresenceMessage_ConnectAck:
		logger.Log("INFO", "PRESENCE", "connected to coordinator: nodeID = %s", payload.ConnectAck.NodeId)
	case *presencepb.CoordinatorPresenceMessage_Error:
		return fmt.Errorf("coordinator returned error: %s", payload.Error.Message)
	default:
		return fmt.Errorf("unexpected ack payload")
	}

	if err := stream.Send(c.endpointMessage(c.cfg.Work.AdvertiseEndpoint)); err != nil {
		return fmt.Errorf("send initial endpoints: %w", err)
	}

	if err := stream.Send(c.statusMessage(false)); err != nil {
		return fmt.Errorf("send initial status: %w", err)
	}

	logger.Log("INFO", "PRESENCE", "initial endpoints and busy=false sent", c.cfg.Agent.NodeID)

	for {
		select {
		case <-ctx.Done():
			logger.Log("INFO", "PRESENCE", "presence client shutting down", c.cfg.Agent.NodeID)
			_ = stream.CloseSend()
			return nil
		case msg := <-c.sendCh:
			if err := stream.Send(msg); err != nil {
				return fmt.Errorf("send presence update: %w", err)
			}
		}
	}
}

func (c *Client) SetBusy(busy bool) {
	select {
	case c.sendCh <- c.statusMessage(busy):
	default:
		logger.Log("INFO", "PRESENCE", "busy update dropped: busy = %v", c.cfg.Agent.NodeID)
	}
}

func (c *Client) SetEndpoints(endpoints []config.Endpoint) {
	select {
	case c.sendCh <- c.endpointMessage(endpoints):
	default:
		logger.Log("INFO", "PRESENCE", "endpoint update dropped", c.cfg.Agent.NodeID)
	}
}

func (c *Client) statusMessage(busy bool) *presencepb.AgentPresenceMessage {
	return &presencepb.AgentPresenceMessage{
		Payload: &presencepb.AgentPresenceMessage_Status{
			Status: &presencepb.StatusUpdate{
				Busy: busy,
			},
		},
	}
}

func (c *Client) endpointMessage(endpoints []config.Endpoint) *presencepb.AgentPresenceMessage {
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

func (c *Client) register(ctx context.Context) (*registerResponse, error) {
	reqBody := registerRequest{
		ID:           c.cfg.Agent.NodeID,
		Capabilities: c.cfg.Agent.Capabilities,
		Endpoints:    c.cfg.Work.AdvertiseEndpoint,
		Busy:         false,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal register request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.cfg.Coordinator.RegisterURL,
		bytes.NewReader(data),
	)
	if err != nil {
		return nil, fmt.Errorf("build register request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform register request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("register failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var out registerResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode register response: %w", err)
	}

	logger.Log("INFO", "PRESENCE", "registered: nodeID = %s sessionID = %s grpcAddress = %s",
		out.NodeID, out.SessionID, out.GRPCAddress)

	return &out, nil
}
