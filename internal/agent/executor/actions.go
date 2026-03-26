package executor

import (
	"context"
	"runtime"

	"Orch/internal/agent/model"
)

func (e *Executor) registerBuiltinActions() {
	e.registerAction("echo", e.echoAction)
	e.registerAction("hostname", e.hostnameAction)
	e.registerAction("list_processes", e.listProcessesAction)
}

func (e *Executor) echoAction(ctx context.Context, args map[string]string) (<-chan model.Event, error) {
	ch := make(chan model.Event, 2)

	go func() {
		defer close(ch)

		select {
		case <-ctx.Done():
			ch <- resultEvent(false, 125, "action canceled")
			return
		default:
		}

		text := args["text"]

		ch <- outputEvent("stdout", text)
		ch <- resultEvent(true, 0, "action completed")
	}()

	return ch, nil
}

func (e *Executor) hostnameAction(ctx context.Context, _ map[string]string) (<-chan model.Event, error) {
	if runtime.GOOS == "windows" {
		return e.runCommand(ctx, "cmd", "hostname", 0)
	}

	return e.runCommand(ctx, "sh", "hostname", 0)
}

func (e *Executor) listProcessesAction(ctx context.Context, _ map[string]string) (<-chan model.Event, error) {
	if runtime.GOOS == "windows" {
		return e.runCommand(ctx, "cmd", "tasklist", 0)
	}

	return e.runCommand(ctx, "sh", "ps -eo pid,comm,%cpu,%mem", 0)
}
