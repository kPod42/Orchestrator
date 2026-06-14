package security

import (
	"Orch/internal/agent/action"
	"fmt"
	"strings"
	"time"

	"Orch/internal/agent/config"
)

type ActionPolicyMode string

const (
	ActionPolicyAllowAllKnown ActionPolicyMode = "allow_all_known"
	ActionPolicyWhitelist     ActionPolicyMode = "whitelist"
	ActionPolicyBlacklist     ActionPolicyMode = "blacklist"
)

type Policy struct {
	allowExec     bool
	execTime      time.Duration
	allowedShells map[string]struct{}

	actionMode ActionPolicyMode
	whitelist  map[string]struct{}
	blacklist  map[string]struct{}
}

type KnownActions interface {
	HasAction(name string) bool
}

func New(cfg config.SecurityConfig) *Policy {
	p := &Policy{
		allowExec:     cfg.AllowExec,
		execTime:      time.Duration(cfg.ExecTimeoutSec) * time.Second,
		allowedShells: make(map[string]struct{}),

		actionMode: normalizeActionPolicyMode(cfg.ActionPolicy.Mode),
		whitelist:  make(map[string]struct{}),
		blacklist:  make(map[string]struct{}),
	}

	if p.execTime <= 0 {
		p.execTime = 60 * time.Second
	}

	for _, shell := range cfg.AllowedShells {
		shell = normalizeName(shell)
		if shell != "" {
			p.allowedShells[shell] = struct{}{}
		}
	}

	for _, act := range cfg.ActionPolicy.Whitelist {
		act = normalizeName(act)
		if act != "" {
			p.whitelist[act] = struct{}{}
		}
	}

	for _, act := range cfg.ActionPolicy.Blacklist {
		act = normalizeName(act)
		if act != "" {
			p.blacklist[act] = struct{}{}
		}
	}

	return p
}

func (p *Policy) CheckAction(action string, known action.KnownActions) error {
	action = normalizeName(action)
	if action == "" {
		return fmt.Errorf("empty action")
	}

	if known == nil {
		return fmt.Errorf("known actions registry is not configured")
	}

	if !known.HasAction(action) {
		return fmt.Errorf("unknown action: %s", action)
	}

	switch p.actionMode {
	case ActionPolicyAllowAllKnown:
		return nil

	case ActionPolicyWhitelist:
		if _, ok := p.whitelist[action]; !ok {
			return fmt.Errorf("action is not in whitelist: %s", action)
		}
		return nil

	case ActionPolicyBlacklist:
		if _, blocked := p.blacklist[action]; blocked {
			return fmt.Errorf("action is blocked by blacklist: %s", action)
		}
		return nil

	default:
		return fmt.Errorf("unknown action policy mode: %s", p.actionMode)
	}
}

func (p *Policy) CheckExec(shell string) error {

	if !p.allowExec {
		return fmt.Errorf("exec is disabled by policy")
	}

	shell = normalizeName(shell)
	if shell == "" {
		return fmt.Errorf("shell is empty")
	}

	if _, ok := p.allowedShells[shell]; !ok {
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
func normalizeActionPolicyMode(mode string) ActionPolicyMode {
	mode = normalizeName(mode)
	if mode == "" {
		return ActionPolicyAllowAllKnown
	}

	return ActionPolicyMode(mode)
}

func normalizeName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
