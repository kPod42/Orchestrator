package handler

import (
	"Orch/internal/coordinator/model"
	"Orch/internal/coordinator/registry"
	"Orch/pkg/logger"
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
		logger.Log("WARNING", "HTTP", "Failed to decode register request %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Log("INFO", "HTTP", "Received register request: nodeID = %s", node.ID)

	resp, err := h.reg.Register(node)
	if err != nil {
		logger.Log("ERROR", "HTTP", "Failed to register node: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logger.Log("INFO", "HTTP", "Successfully registered node: nodeID = %s", node.ID)

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

	logger.Log("INFO", "HTTP", "Received request for active nodes: count = %d", len(filtered))

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(filtered)
}

func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("ok"))
	if err != nil {
		logger.Log("ERROR", "HTTP", "Failed to write health response: %v", err)
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
func (h *HTTPHandler) Info(w http.ResponseWriter, r *http.Request) {
	info := h.reg.GetCoordinatorInfo()
	logger.Log("INFO", "HTTP", "Received coordinator info request")
	writeJSON(w, http.StatusOK, info)
}
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Log("ERROR", "HTTP", "Failed to encode json response: %v", err)
	}
}
