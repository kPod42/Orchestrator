package actions

import (
	"context"
	"os"

	"Orch/internal/agent/action"
	"Orch/internal/agent/model"
)

type HostnameAction struct{}

func init() {
	action.RegisterBuiltin(HostnameAction{})
}

func (a HostnameAction) Name() string {
	return "hostname"
}

func (a HostnameAction) Description() string {
	return "Returns host name"
}

func (a HostnameAction) Run(ctx context.Context, _ map[string]string) (<-chan model.Event, error) {
	ch := make(chan model.Event, 2)

	go func() {
		defer close(ch)

		select {
		case <-ctx.Done():
			ch <- resultEvent(false, 125, "action canceled")
			return
		default:
		}

		hostname, err := os.Hostname()
		if err != nil {
			ch <- outputEvent("stderr", err.Error())
			ch <- resultEvent(false, 1, err.Error())
			return
		}

		ch <- outputEvent("stdout", hostname)
		ch <- resultEvent(true, 0, "action completed")
	}()

	return ch, nil
}
