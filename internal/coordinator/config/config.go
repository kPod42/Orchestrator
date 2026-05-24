package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"Orch/internal/coordinator/model"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Cluster     ClusterConfig     `json:"cluster" yaml:"cluster"`
	Coordinator CoordinatorConfig `json:"coordinator" yaml:"coordinator"`
}

type ClusterConfig struct {
	ID string `json:"id" yaml:"id"`
}

type CoordinatorConfig struct {
	ID            string `json:"id" yaml:"id"`
	ConfigVersion int    `json:"config_version" yaml:"config_version"`

	HTTP HTTPConfig `json:"http" yaml:"http"`
	GRPC GRPCConfig `json:"grpc" yaml:"grpc"`

	Endpoints []model.Endpoint `json:"endpoints" yaml:"endpoints"`
}

type HTTPConfig struct {
	ListenAddr string `json:"listen_addr" yaml:"listen_addr"`
}

type GRPCConfig struct {
	ListenAddr string `json:"listen_addr" yaml:"listen_addr"`
}

func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open coordinator config: %w", err)
	}
	defer file.Close()

	var cfg Config

	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		dec := yaml.NewDecoder(file)
		dec.KnownFields(true)

		if err := dec.Decode(&cfg); err != nil {
			return nil, fmt.Errorf("decode yaml coordinator config: %w", err)
		}

	case ".json":
		dec := json.NewDecoder(file)
		dec.DisallowUnknownFields()

		if err := dec.Decode(&cfg); err != nil {
			return nil, fmt.Errorf("decode json coordinator config: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported coordinator config extension: %s", filepath.Ext(path))
	}

	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Cluster.ID == "" {
		c.Cluster.ID = "default-cluster"
	}

	if c.Coordinator.ID == "" {
		c.Coordinator.ID = "main"
	}

	if c.Coordinator.ConfigVersion <= 0 {
		c.Coordinator.ConfigVersion = 1
	}

	if c.Coordinator.HTTP.ListenAddr == "" {
		c.Coordinator.HTTP.ListenAddr = "0.0.0.0:8080"
	}

	if c.Coordinator.GRPC.ListenAddr == "" {
		c.Coordinator.GRPC.ListenAddr = "0.0.0.0:9090"
	}
}

func (c *Config) validate() error {
	if c.Cluster.ID == "" {
		return fmt.Errorf("cluster.id is required")
	}

	if c.Coordinator.ID == "" {
		return fmt.Errorf("coordinator.id is required")
	}

	if c.Coordinator.HTTP.ListenAddr == "" {
		return fmt.Errorf("coordinator.http.listen_addr is required")
	}

	if c.Coordinator.GRPC.ListenAddr == "" {
		return fmt.Errorf("coordinator.grpc.listen_addr is required")
	}

	if len(c.Coordinator.Endpoints) == 0 {
		return fmt.Errorf("coordinator.endpoints must contain at least one endpoint")
	}

	hasGRPC := false

	for i, endpoint := range c.Coordinator.Endpoints {
		if endpoint.Type == "" {
			return fmt.Errorf("coordinator.endpoints[%d].kind is required", i)
		}

		if endpoint.Address == "" {
			return fmt.Errorf("coordinator.endpoints[%d].address is required", i)
		}

		if endpoint.Type == "grpc" {
			hasGRPC = true
		}
	}

	if !hasGRPC {
		return fmt.Errorf("coordinator.endpoints must contain at least one grpc endpoint")
	}

	return nil
}
