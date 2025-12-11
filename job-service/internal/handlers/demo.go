package handlers

import (
	"encoding/json"
	"net/http"

	"job-service/internal/service"
)

// DemoHandler handles demo-related HTTP requests
type DemoHandler struct {
	demoGenerator *service.DemoJobGenerator
}

// NewDemoHandler creates a new demo handler
func NewDemoHandler(demoGenerator *service.DemoJobGenerator) *DemoHandler {
	return &DemoHandler{
		demoGenerator: demoGenerator,
	}
}

// RegisterDemoRoutes sets up demo HTTP routes
func (h *DemoHandler) RegisterDemoRoutes(router interface{}) {
	// This will be called from the main HTTP handler
}

// StartDemo starts the demo job generator
func (h *DemoHandler) StartDemo(w http.ResponseWriter, r *http.Request) {
	h.demoGenerator.Start()

	response := map[string]interface{}{
		"status":  "started",
		"message": "Demo job generator started",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StopDemo stops the demo job generator
func (h *DemoHandler) StopDemo(w http.ResponseWriter, r *http.Request) {
	h.demoGenerator.Stop()

	response := map[string]interface{}{
		"status":  "stopped",
		"message": "Demo job generator stopped",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDemoStatus returns the current demo status
func (h *DemoHandler) GetDemoStatus(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"running": h.demoGenerator.IsRunning(),
		"status":  "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
