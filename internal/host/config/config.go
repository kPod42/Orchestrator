package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Host HostConfig `json:"host", yaml:"host"`
}

type HostConfig struct {
	CoordinatorBaseUrl string `json:"coordinator_base_url" yaml:"coordinator_base_url"`
	RequestTimeoutSec  int    `json:"request_timeout_sec" yaml:"request_timeout_sec"`
}

func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %v", path, err)
	}
	defer file.Close()

	var cfg Config

	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		dec := yaml.NewDecoder(file)
		dec.KnownFields(true)

		if err := dec.Decode(&cfg); err != nil {
			return nil, fmt.Errorf("decode yaml host config: %w", err)
		}
	case ".json":
		dec := json.NewDecoder(file)
		dec.DisallowUnknownFields()

		if err := dec.Decode(&cfg); err != nil {
			return nil, fmt.Errorf("decode json host config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported host config extension: %s", filepath.Ext(path))
	}
	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
func (cfg *Config) applyDefaults() {
	if cfg.Host.RequestTimeoutSec <= 0 {
		cfg.Host.RequestTimeoutSec = 30
	}
}
func (cfg *Config) validate() error {
	if strings.TrimSpace(cfg.Host.CoordinatorBaseUrl) == "" {
		return fmt.Errorf("host coordinator base url is required")
	}
	return nil
}
