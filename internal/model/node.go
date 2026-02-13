package model

import "time"

type Node struct {
	ID          string   `json:"id"`
	IP          string   `json:"ip"`
	Port        int      `json:"port"`
	Capabilites []string `json:"capabilites"`

	LastSeen time.Time `json:"-"`
}
