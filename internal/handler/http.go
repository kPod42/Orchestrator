package handler

import (
	"Coordinator/internal/model"
	"Coordinator/internal/registry"
	"encoding/json"
	"net/http"
)

type HTTPHandler struct {
	reg registry.Registry
}

func NewHTTPHandler(reg registry.Registry) *HTTPHandler {
	return &HTTPHandler{reg: reg}
}

func (h *HTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
	var node model.Node
	if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := h.reg.Register(node)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *HTTPHandler) GetNodes(w http.ResponseWriter, r *http.Request) {
	queryCapability := r.URL.Query().Get("capability")
	freeOnly := r.URL.Query().Get("free") == "true"

	nodes := h.reg.GetActive()
	filtered := make([]model.Node, 0, len(nodes))
	for _, node := range nodes {
		if freeOnly && node.Busy {
			continue
		}
		if queryCapability != "" && !hasCapability(node, queryCapability) {
			continue
		}
		filtered = append(filtered, node)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(filtered)
}

func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("ok"))
	if err != nil {
		return
	}
}
func hasCapability(node model.Node, queryCapability string) bool {
	for _, nodeCapability := range node.Capabilities {
		if nodeCapability == queryCapability {
			return true
		}
	}
	return false
}
