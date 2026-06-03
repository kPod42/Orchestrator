package model

type ExecuteRequest struct {
	Mode    string            `json:"mode"`
	Action  string            `json:"action"`
	Targets []string          `json:"targets"`
	Args    map[string]string `json:"args,omitempty"`
}

type NodeExecutionResult struct {
	NodeID   string `json:"nodeId"`
	Status   string `json:"status"`
	ExitCode int32  `json:"exitCode"`
	Message  string `json:"message"`
	Output   string `json:"output,omitempty"`
}

type ExecuteResponse struct {
	Results []NodeExecutionResult `json:"results"`
}
