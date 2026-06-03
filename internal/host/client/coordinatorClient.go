package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"Orch/internal/host/model"
)

type CoordinatorClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewCoordinatorClient(baseURL string, timeout time.Duration) *CoordinatorClient {
	return &CoordinatorClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *CoordinatorClient) GetNodes(ctx context.Context) ([]model.Node, error) {
	url := c.baseURL + "/coordinator/nodes"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build nodes request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform nodes request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nodes request failed: status=%d body=%s", resp.Status, string(body))
	}
	var nodes []model.Node
	if err = json.Unmarshal(body, &nodes); err != nil {
		return nil, fmt.Errorf("decode nodes response: %w", err)
	}
	return nodes, nil
}

func (c *CoordinatorClient) Execute(
	ctx context.Context,
	request model.ExecuteRequest,
) (model.ExecuteResponse, error) {
	url := c.baseURL + "/coordinator/action"

	data, err := json.Marshal(request)
	if err != nil {
		return model.ExecuteResponse{}, fmt.Errorf("marshal execute request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return model.ExecuteResponse{}, fmt.Errorf("build execute request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return model.ExecuteResponse{}, fmt.Errorf("perform execute request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return model.ExecuteResponse{}, fmt.Errorf("execute request failed: status=%d body=%s", resp.Status, string(body))
	}
	var out model.ExecuteResponse
	if err = json.Unmarshal(body, &out); err != nil {
		return model.ExecuteResponse{}, fmt.Errorf("decode execute response: %w", err)
	}
	return out, nil
}
