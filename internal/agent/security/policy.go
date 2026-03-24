package security

import (
	"fmt"
	"strings"
	"time"

	"Orch/internal/agent/config"
)

type Policy struct {
	devMode        bool
	allowExec      bool
	execTimeoutSec time.Duration
	allowedActions map[string]struct{}
	allowedShells  map[string]struct{}
}

func New(cfg config.SecurityConfig) *Policy {
	p := &Policy{
		devMode:        cfg.DevMode,
		allowExec:      cfg.AllowExec,
		execTimeoutSec: time.Duration(cfg.ExecTimeoutSec) * time.Second,
		allowedActions: make(map[string]struct{}),
		allowedShells:  make(map[string]struct{}),
	}
	for _, a := range cfg.AllowedActions {
		p.allowedActions[strings.ToLower(strings.TrimSpace(a))] = struct{}{}
	}
	for _, s := range cfg.AllowedShells {
		p.allowedShells[strings.ToLower(strings.TrimSpace(s))] = struct{}{}
	}
	return p
}
func (p *Policy) CheckActions(action string) error {
	action = strings.ToLower(strings.TrimSpace(action))
	if action == "" {
		return fmt.Errorf("empty action")
	}
	if _, ok := p.allowedActions[action]; !ok {
		return fmt.Errorf("action is not allowed: %s", action)
	}
	return nil
}
func (p *Policy) CheckExec(shell string) error {
	if !p.devMode {
		return fmt.Errorf("exec is disabled: dev_mode = false")
	}
	if !p.allowExec {
		return fmt.Errorf("exec is disabled by policy")
	}
	shell = strings.ToLower(strings.TrimSpace(shell))
	if shell == "" {
		return fmt.Errorf("shell is empty")
	}
	if _, ok := p.allowedShells[shell]; !ok {
		return fmt.Errorf("shell %q not allowed", shell)
	}
	return nil
}
func (p *Policy) EffectiveExecTimeoutSec(requestedSec int32) time.Duration {
	if p.execTimeoutSec <= 0 {
		return p.execTimeoutSec
	}
	requested := time.Duration(requestedSec) * time.Second
	if requested > p.execTimeoutSec {
		return p.execTimeoutSec
	}
	return requested
}
