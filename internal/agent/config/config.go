package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Endpoint struct {
	Kind    string `json:"kind"`
	Address string `json:"address"`
}

type AgentConfig struct {
	NodeID       string   `json:"node_id"`
	Capabilities []string `json:"capabilities"`
}

type CoordinatorConfig struct {
	RegisterURL string `json:"register_url"`
	GRPCAddress string `json:"grpc_address"`
}

type WorkConfig struct {
	ListenAddress     string     `json:"listen_address"`
	AdvertiseEndpoint []Endpoint `json:"advertise_endpoint"`
}

type SecurityConfig struct {
	DevMode        bool     `json:"dev_mode"`
	AllowExec      bool     `json:"allow_exec"`
	AllowedShells  []string `json:"allowed_shells"`
	AllowedActions []string `json:"allowed_actions"`
	ExecTimeoutSec int      `json:"exec_timeout_sec"`
}

type Config struct {
	Agent       AgentConfig       `json:"agent"`
	Coordinator CoordinatorConfig `json:"coordinator"`
	Work        WorkConfig        `json:"work"`
	Security    SecurityConfig    `json:"security"`
}

func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config
	dec := json.NewDecoder(file)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Agent.NodeID == "" {
		return fmt.Errorf("agent node ID is empty")
	}
	if c.Coordinator.RegisterURL == "" {
		return fmt.Errorf("coordinator register URL is empty")
	}
	if c.Work.ListenAddress == "" {
		return fmt.Errorf("work listen address is empty")
	}
	if (len(c.Work.AdvertiseEndpoint) == 0) || (len(c.Work.AdvertiseEndpoint[0].Address) == 0) {
		return fmt.Errorf("work advertise endpoint is empty")
	}
	if c.Security.ExecTimeoutSec <= 0 {
		c.Security.ExecTimeoutSec = 60
	}
	return nil
}
