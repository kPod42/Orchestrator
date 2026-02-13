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
	if node.ID == "" {
		http.Error(w, "Node ID can't be empty", http.StatusBadRequest)
		return
	}
	if err := h.reg.Register(node); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *HTTPHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NodeID string `json:"nodeId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	if err := h.reg.Heartbeat(req.NodeID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *HTTPHandler) GetNodes(w http.ResponseWriter, r *http.Request) {
	capability := r.URL.Query().Get("capability")

	nodes := h.reg.GetActive()

	if capability != "" {
		var filteredNodes []model.Node
		for _, node := range nodes {
			for _, capability := range node.Capabilites {
				if capability == capability {
					filteredNodes = append(filteredNodes, node)
					break
				}
			}
		}
		nodes = filteredNodes
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(nodes)
	if err != nil {
		return
	}
}

func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("ok"))
	if err != nil {
		return
	}
}
