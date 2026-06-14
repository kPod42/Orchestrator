package actions

import (
	"context"

	"Orch/internal/agent/action"
	"Orch/internal/agent/model"
)

type EchoAction struct{}

func init() {
	action.RegisterBuiltin(EchoAction{})
}

func (a EchoAction) Name() string {
	return "echo"
}

func (a EchoAction) Description() string {
	return "Returns provided text"
}

func (a EchoAction) Run(ctx context.Context, args map[string]string) (<-chan model.Event, error) {
	ch := make(chan model.Event, 2)

	go func() {
		defer close(ch)

		select {
		case <-ctx.Done():
			ch <- resultEvent(false, 125, "action canceled")
			return
		default:
		}

		ch <- outputEvent("stdout", args["text"])
		ch <- resultEvent(true, 0, "action completed")
	}()

	return ch, nil
}
