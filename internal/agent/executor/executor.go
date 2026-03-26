package executor

import (
	"context"
	"fmt"
	"strings"

	"Orch/internal/agent/model"
)

type ActionHandler func(ctx context.Context, args map[string]string) (<-chan model.Event, error)

type Executor struct {
	actions map[string]ActionHandler
}

func New() *Executor {
	e := &Executor{
		actions: make(map[string]ActionHandler),
	}

	e.registerBuiltinActions()
	return e
}

func (e *Executor) RunAction(ctx context.Context, action string, args map[string]string) (<-chan model.Event, error) {
	action = strings.ToLower(strings.TrimSpace(action))

	handler, ok := e.actions[action]
	if !ok {
		return nil, fmt.Errorf("unknown action: %s", action)
	}

	return handler(ctx, args)
}

func (e *Executor) registerAction(name string, handler ActionHandler) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" || handler == nil {
		return
	}

	e.actions[name] = handler
}

func outputEvent(stream, chunk string) model.Event {
	return model.Event{
		Output: &model.Output{
			Stream: stream,
			Chunk:  chunk,
		},
	}
}

func resultEvent(success bool, exitCode int32, message string) model.Event {
	return model.Event{
		Result: &model.Result{
			Success:  success,
			ExitCode: exitCode,
			Message:  message,
		},
	}
}
