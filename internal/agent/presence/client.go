package presence

import (
	"sync"

	"Orch/internal/agent/config"
)

type Client struct {
	cfg *config.Config

	mu    sync.Mutex
	state snapshot

	flushCh chan struct{}

	readyCh   chan struct{}
	readyOnce sync.Once
}

func New(cfg *config.Config) *Client {
	initialEndpoints := append([]config.Endpoint(nil), cfg.Work.AdvertiseEndpoint...)

	return &Client{
		cfg: cfg,
		state: snapshot{
			busy:           false,
			endpoints:      initialEndpoints,
			dirtyBusy:      true,
			dirtyEndpoints: true,
		},
		flushCh: make(chan struct{}, 1),
		readyCh: make(chan struct{}),
	}
}

func (c *Client) Name() string {
	return "presence"
}

func (c *Client) Ready() <-chan struct{} {
	return c.readyCh
}

func (c *Client) SetBusy(b bool) {
	c.mu.Lock()
	changed := c.state.busy != b
	if changed {
		c.state.busy = b
		c.state.dirtyBusy = true
	}
	c.mu.Unlock()

	if changed {
		c.notifyFlush()
	}
}

func (c *Client) SetEndpoints(endpoints []config.Endpoint) {
	copyEndpoints := append([]config.Endpoint(nil), endpoints...)

	c.mu.Lock()
	changed := !endpointsEqual(c.state.endpoints, copyEndpoints)
	if changed {
		c.state.endpoints = copyEndpoints
		c.state.dirtyEndpoints = true
	}
	c.mu.Unlock()

	if changed {
		c.notifyFlush()
	}
}

func (c *Client) takeDirtySnapshot() snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	snap := snapshot{
		busy:           c.state.busy,
		endpoints:      append([]config.Endpoint(nil), c.state.endpoints...),
		dirtyBusy:      c.state.dirtyBusy,
		dirtyEndpoints: c.state.dirtyEndpoints,
	}

	c.state.dirtyBusy = false
	c.state.dirtyEndpoints = false

	return snap
}

func (c *Client) notifyFlush() {
	select {
	case c.flushCh <- struct{}{}:
	default:
	}
}

func (c *Client) markReady() {
	c.readyOnce.Do(func() {
		close(c.readyCh)
	})
}

func endpointsEqual(a, b []config.Endpoint) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Kind != b[i].Kind || a[i].Address != b[i].Address {
			return false
		}
	}

	return true
}
func (c *Client) currentSnapshot() snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	return snapshot{
		busy:           c.state.busy,
		endpoints:      append([]config.Endpoint(nil), c.state.endpoints...),
		dirtyBusy:      c.state.dirtyBusy,
		dirtyEndpoints: c.state.dirtyEndpoints,
	}
}
