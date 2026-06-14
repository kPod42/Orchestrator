package executor

import (
	"context"
	"fmt"

	"Orch/internal/agent/action"
	"Orch/internal/agent/model"
)

type ActionPolicy interface {
	CheckAction(action string, known action.KnownActions) error
}

type Executor struct {
	registry *action.Registry
	policy   ActionPolicy
}

func New(policy ActionPolicy) *Executor {
	return NewWithRegistry(policy, action.NewDefaultRegistry())
}

func NewWithRegistry(policy ActionPolicy, registry *action.Registry) *Executor {
	if registry == nil {
		registry = action.NewDefaultRegistry()
	}

	return &Executor{
		registry: registry,
		policy:   policy,
	}
}

func (e *Executor) RunAction(ctx context.Context, actionName string, args map[string]string) (<-chan model.Event, error) {
	actionName = action.NormalizeName(actionName)
	if actionName == "" {
		return nil, fmt.Errorf("empty action")
	}

	if e.policy != nil {
		if err := e.policy.CheckAction(actionName, e); err != nil {
			return nil, err
		}
	}

	handler, err := e.registry.MustGet(actionName)
	if err != nil {
		return nil, err
	}

	return handler.Run(ctx, args)
}

func (e *Executor) HasAction(name string) bool {
	if e == nil || e.registry == nil {
		return false
	}

	return e.registry.HasAction(name)
}

func (e *Executor) ListActions() []action.Info {
	if e == nil || e.registry == nil {
		return nil
	}

	return e.registry.List()
}
