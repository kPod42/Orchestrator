package presence

import (
	"Orch/internal/agent/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"Orch/pkg/logger"
)

func (c *Client) register(ctx context.Context) (*registerResponse, error) {
	snap := c.currentSnapshot()

	reqBody := registerRequest{
		ID:           c.cfg.Agent.NodeID,
		Capabilities: c.cfg.Agent.Capabilities,
		Endpoints:    snap.endpoints,
		Busy:         snap.busy,
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

	normalizeRegisterResponse(&out)

	logger.Log(
		"INFO",
		"PRESENCE",
		"registered: nodeID = %s sessionID = %s coordinatorEndpoints = %d configVersion = %d",
		out.NodeID,
		out.SessionID,
		len(out.CoordinatorEndpoints),
		out.ConfigVersion,
	)

	return &out, nil
}

func normalizeRegisterResponse(resp *registerResponse) {
	if len(resp.CoordinatorEndpoints) > 0 {
		return
	}

	if resp.GRPCAddress == "" {
		return
	}

	resp.CoordinatorEndpoints = []config.Endpoint{
		{
			Name:     "legacy",
			Kind:     "grpc",
			Address:  resp.GRPCAddress,
			Scope:    "legacy",
			Priority: 1000,
		},
	}
}
