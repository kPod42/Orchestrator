package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Endpoint struct {
	Kind    string `json:"kind" yaml:"kind"`
	Address string `json:"address" yaml:"address"`
}

type AgentConfig struct {
	NodeID       string   `json:"node_id" yaml:"node_id"`
	Capabilities []string `json:"capabilities" yaml:"capabilities"`
}

type CoordinatorConfig struct {
	RegisterURL string `json:"register_url" yaml:"register_url"`
	GRPCAddress string `json:"grpc_address" yaml:"grpc_address"`
}

type WorkConfig struct {
	ListenAddress     string     `json:"listen_address" yaml:"listen_address"`
	AdvertiseEndpoint []Endpoint `json:"advertise_endpoint" yaml:"advertise_endpoint"`
}

type SecurityConfig struct {
	DevMode        bool     `json:"dev_mode" yaml:"dev_mode"`
	AllowExec      bool     `json:"allow_exec" yaml:"allow_exec"`
	AllowedShells  []string `json:"allowed_shells" yaml:"allowed_shells"`
	AllowedActions []string `json:"allowed_actions" yaml:"allowed_actions"`
	ExecTimeoutSec int      `json:"exec_timeout_sec" yaml:"exec_timeout_sec"`
}

type Config struct {
	Agent       AgentConfig       `json:"agent" yaml:"agent"`
	Coordinator CoordinatorConfig `json:"coordinator" yaml:"coordinator"`
	Work        WorkConfig        `json:"work" yaml:"work"`
	Security    SecurityConfig    `json:"security" yaml:"security"`
}

func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config

	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		dec := yaml.NewDecoder(file)
		dec.KnownFields(true)

		if err := dec.Decode(&cfg); err != nil {
			return nil, fmt.Errorf("failed to decode yaml config file: %w", err)
		}

	case ".json":
		dec := json.NewDecoder(file)
		dec.DisallowUnknownFields()

		if err := dec.Decode(&cfg); err != nil {
			return nil, fmt.Errorf("failed to decode json config file: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported config extension: %s", filepath.Ext(path))
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if strings.TrimSpace(c.Agent.NodeID) == "" {
		return fmt.Errorf("agent node ID is empty")
	}

	if strings.TrimSpace(c.Coordinator.RegisterURL) == "" {
		return fmt.Errorf("coordinator register URL is empty")
	}

	if strings.TrimSpace(c.Work.ListenAddress) == "" {
		return fmt.Errorf("work listen address is empty")
	}

	if len(c.Work.AdvertiseEndpoint) == 0 {
		return fmt.Errorf("work advertise endpoint is empty")
	}

	for i, ep := range c.Work.AdvertiseEndpoint {
		if strings.TrimSpace(ep.Kind) == "" {
			return fmt.Errorf("work advertise endpoint[%d] kind is empty", i)
		}
		if strings.TrimSpace(ep.Address) == "" {
			return fmt.Errorf("work advertise endpoint[%d] address is empty", i)
		}
	}

	if c.Security.ExecTimeoutSec <= 0 {
		c.Security.ExecTimeoutSec = 60
	}

	return nil
}
