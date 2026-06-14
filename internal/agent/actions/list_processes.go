package actions

import (
	"context"
	"runtime"
	"time"

	"Orch/internal/agent/action"
	"Orch/internal/agent/model"
	"Orch/internal/agent/shellcmd"
)

type ListProcessesAction struct{}

func init() {
	action.RegisterBuiltin(ListProcessesAction{})
}

func (a ListProcessesAction) Name() string {
	return "list_processes"
}

func (a ListProcessesAction) Description() string {
	return "Returns process list"
}

func (a ListProcessesAction) Run(ctx context.Context, _ map[string]string) (<-chan model.Event, error) {
	if runtime.GOOS == "windows" {
		return shellcmd.Run(
			ctx,
			"powershell",
			"Get-Process | Select-Object -First 20 ProcessName, Id, WorkingSet64 | ConvertTo-Csv -NoTypeInformation",
			30*time.Second,
		)
	}

	return shellcmd.Run(ctx, "sh", "ps -eo pid,comm,%cpu,%mem", 30*time.Second)
}
