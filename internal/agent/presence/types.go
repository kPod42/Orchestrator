package presence

import "Orch/internal/agent/config"

type registerRequest struct {
	ID           string            `json:"id"`
	Capabilities []string          `json:"capabilities"`
	Endpoints    []config.Endpoint `json:"endpoints"`
	Busy         bool              `json:"busy"`
}

type registerResponse struct {
	NodeID    string `json:"nodeId"`
	SessionID string `json:"sessionId"`
	//Legacy field. Старый вариант ответа
	GRPCAddress string `json:"grpcAddress"`

	CoordinatorEndpoints []config.Endpoint `json:"coordinatorEndpoints"`
	ConfigVersion        int               `json:"configVersion"`
}
type snapshot struct {
	busy           bool
	endpoints      []config.Endpoint
	dirtyBusy      bool
	dirtyEndpoints bool
}
