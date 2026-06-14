package executor

import (
	"context"
	"time"

	"Orch/internal/agent/model"
	"Orch/internal/agent/shellcmd"
)

func (e *Executor) RunExec(
	ctx context.Context,
	shell string,
	command string,
	timeout time.Duration,
) (<-chan model.Event, error) {
	return shellcmd.Run(ctx, shell, command, timeout)
}
