package handler

import (
	"encoding/json"
	"net/http"

	"Orch/internal/coordinator/dispatcher"
	"Orch/internal/coordinator/model"
	"Orch/internal/coordinator/registry"
	"Orch/pkg/logger"
)

type HTTPHandler struct {
	reg        registry.Registry
	dispatcher *dispatcher.Dispatcher
}

func NewHTTPHandler(
	reg registry.Registry,
	dispatcher *dispatcher.Dispatcher,
) *HTTPHandler {
	return &HTTPHandler{
		reg:        reg,
		dispatcher: dispatcher,
	}
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

	writeJSON(w, http.StatusCreated, resp)
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

	writeJSON(w, http.StatusOK, filtered)
}

func (h *HTTPHandler) ExecuteAction(w http.ResponseWriter, r *http.Request) {
	var request model.ExecuteRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Log("WARNING", "HTTP", "Failed to decode execute request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if request.Mode != "" && request.Mode != "action" {
		http.Error(w, "only action mode is supported", http.StatusBadRequest)
		return
	}

	logger.Log(
		"INFO",
		"HTTP",
		"Received execute request: action = %s targets = %v",
		request.Action,
		request.Targets,
	)

	response, err := h.dispatcher.ExecuteAction(r.Context(), request)
	if err != nil {
		logger.Log("ERROR", "HTTP", "Failed to execute action: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("ok"))
	if err != nil {
		logger.Log("ERROR", "HTTP", "Failed to write health response: %v", err)
		return
	}
}

func (h *HTTPHandler) Info(w http.ResponseWriter, r *http.Request) {
	info := h.reg.GetCoordinatorInfo()

	logger.Log("INFO", "HTTP", "Received coordinator info request")

	writeJSON(w, http.StatusOK, info)
}

func hasCapability(node model.Node, queryCapability string) bool {
	for _, nodeCapability := range node.Capabilities {
		if nodeCapability == queryCapability {
			return true
		}
	}

	return false
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Log("ERROR", "HTTP", "Failed to encode json response: %v", err)
	}
}
