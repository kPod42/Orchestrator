package actions

import (
	"Orch/internal/agent/model"
)

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
