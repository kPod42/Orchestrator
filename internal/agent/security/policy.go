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
	execTime       time.Duration
	allowedAction  map[string]struct{}
	allowedTargets map[string]struct{}
}

func New(cfg config.SecurityConfig) *Policy {
	p := &Policy{
		devMode:        cfg.DevMode,
		allowExec:      cfg.AllowExec,
		execTime:       time.Duration(cfg.ExecTimeoutSec) * time.Second,
		allowedAction:  make(map[string]struct{}),
		allowedTargets: make(map[string]struct{}),
	}
	for _, a := range cfg.AllowedActions {
		a = strings.ToLower(strings.TrimSpace(a))
		if a != "" {
			p.allowedAction[a] = struct{}{}
		}
	}
	for _, s := range cfg.AllowedShells {
		s = strings.ToLower(strings.TrimSpace(s))
		if s != "" {
			p.allowedTargets[s] = struct{}{}
		}
		if p.execTime <= 0 {
			p.execTime = 60 * time.Second
		}
	}
	return p
}
func (p *Policy) CheckAction(action string) error {
	action = strings.ToLower(strings.TrimSpace(action))
	if action == "" {
		return fmt.Errorf("empty action")
	}
	if _, ok := p.allowedAction[action]; !ok {
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
	if _, ok := p.allowedTargets[shell]; !ok {
		return fmt.Errorf("shell %q not allowed", shell)
	}
	return nil
}

func (p *Policy) EffectiveExecTimeout(requestedSec int32) time.Duration {
	if p.execTime <= 0 {
		return 60 * time.Second
	}

	if requestedSec <= 0 {
		return p.execTime
	}

	requested := time.Duration(requestedSec) * time.Second
	if requested > p.execTime {
		return p.execTime
	}

	return requested
}
