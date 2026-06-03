package model

type Endpoint struct {
	Name     string `json:"name,omitempty" yaml:"name,omitempty"`
	Kind     string `json:"kind" yaml:"kind"`
	Address  string `json:"address" yaml:"address"`
	Scope    string `json:"scope,omitempty" yaml:"scope,omitempty"`
	Priority int    `json:"priority,omitempty" yaml:"priority,omitempty"`
}

type Node struct {
	ID           string     `json:"id"`
	Capabilities []string   `json:"capabilities"`
	Endpoints    []Endpoint `json:"endpoints"`
	Busy         bool       `json:"busy"`
}
