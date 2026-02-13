package main

import (
	"Coordinator/internal/logger"
	"log"
	"net/http"
	"time"

	"Coordinator/internal/handler"
	"Coordinator/internal/registry"
)

func main() {
	reg := registry.NewInMemoryRegistry()
	h := handler.NewHTTPHandler(reg)
	logger.Init()

	http.HandleFunc("/register", h.Register)
	http.HandleFunc("/heartbeat", h.Heartbeat)
	http.HandleFunc("/nodes", h.GetNodes)
	http.HandleFunc("/health", h.Health)

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for range ticker.C {
			err := reg.RemoveStale(30 * time.Second)
			if err != nil {
				return
			}
		}
	}()
	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
