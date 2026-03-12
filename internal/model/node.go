package model

type Endpoint struct {
	Type    string `json:"kind"`
	Address string `json:"address"`
}
type Node struct {
	ID           string     `json:"id"`
	Capabilities []string   `json:"capabilities"`
	Endpoints    []Endpoint `json:"endpoints"`
	Busy         bool       `json:"busy"`
}

type RegisterResponse struct {
	NodeID      string `json:"nodeId"`
	SessionID   string `json:"sessionId"`
	GRPCAddress string `json:"grpcAddress"`
}
