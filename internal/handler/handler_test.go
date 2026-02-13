package handler

import (
	"Coordinator/internal/registry"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterHandler(t *testing.T) {
	reg := registry.NewInMemoryRegistry()
	h := NewHTTPHandler(reg)

	body := `{
        "id":"agent-1",
        "ip":"127.0.0.1",
        "port":9000,
        "capabilities":["worker"]
    }`
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.Register(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Register failed: %v", w.Code)
	}
}
