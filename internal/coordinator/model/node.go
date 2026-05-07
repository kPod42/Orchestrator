package model

type Endpoint struct {
	Name     string `json:"name,omitempty"`
	Type     string `json:"kind"`
	Address  string `json:"address"`
	Scope    string `json:"scope,omitempty"`
	Priority int    `json:"priority,omitempty"`
}
type Node struct {
	ID           string     `json:"id"`
	Capabilities []string   `json:"capabilities"`
	Endpoints    []Endpoint `json:"endpoints"`
	Busy         bool       `json:"busy"`
}

type RegisterResponse struct {
	NodeID    string `json:"nodeId"`
	SessionID string `json:"sessionId"`
	// Deprecated: оставлено временно, чтобы старый агент не умер сразу.
	// Новый агент должен использовать CoordinatorEndpoints.
	GRPCAddress string `json:"grpcAddress,omitempty"`

	CoordinatorEndpoints []Endpoint `json:"coordinatorEndpoints,omitempty"`
	ConfigVersion        int        `json:"configVersion"`
}

type CoordinatorInfo struct {
	ClusterID            string     `json:"clusterId"`
	CoordinatorID        string     `json:"coordinatorId"`
	ConfigVersion        int        `json:"configVersion"`
	CoordinatorEndpoints []Endpoint `json:"coordinatorEndpoints"`
}
