package handler

import (
	"Coordinator/internal/logger"
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
		logger.Error("Register decode failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.HTTP("register request: nodeId=%s capabilities=%v endpoints=%v",
		node.ID, node.Capabilities, node.Endpoints)

	resp, err := h.reg.Register(node)
	if err != nil {
		logger.Error("register failed: nodeId=%s err=%v", node.ID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logger.HTTP("register success: nodeId=%s sessionId=%s grpc=%s",
		resp.NodeID, resp.SessionID, resp.GRPCAddress)

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

	logger.HTTP("get nodes: capability=%s freeOnly=%v count=%d",
		queryCapability, freeOnly, len(filtered))

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
