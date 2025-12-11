package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"fleet-service/internal/service"
	"fleet-service/internal/storage"

	"github.com/gorilla/mux"
)

// HTTPHandler handles HTTP requests for the fleet service
type HTTPHandler struct {
	fleetService *service.FleetService
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(fleetService *service.FleetService) *HTTPHandler {
	return &HTTPHandler{
		fleetService: fleetService,
	}
}

// RegisterRoutes sets up HTTP routes
func (h *HTTPHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/health", h.Health).Methods("GET")
	router.HandleFunc("/vehicles", h.GetAllVehicles).Methods("GET")
	router.HandleFunc("/vehicles", h.RegisterVehicle).Methods("POST")
	router.HandleFunc("/vehicles/{id}/location", h.UpdateVehicleLocation).Methods("PUT")
	router.HandleFunc("/vehicles/{id}/assign", h.AssignJob).Methods("POST")
	router.HandleFunc("/vehicles/{id}/complete", h.CompleteJob).Methods("POST")
	router.HandleFunc("/vehicles/find", h.FindNearestVehicle).Methods("GET")
}

// Health returns service health status
func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// GetAllVehicles returns all vehicles
func (h *HTTPHandler) GetAllVehicles(w http.ResponseWriter, r *http.Request) {
	vehicles, err := h.fleetService.GetAllVehicles(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vehicles)
}

// RegisterVehicle adds a new vehicle to the fleet
func (h *HTTPHandler) RegisterVehicle(w http.ResponseWriter, r *http.Request) {
	var vehicle storage.Vehicle
	if err := json.NewDecoder(r.Body).Decode(&vehicle); err != nil {
		slog.Error("Failed to decode vehicle registration request", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	slog.Info("Vehicle registration request received",
		"vehicle_id", vehicle.ID,
		"region", vehicle.Region,
		"location_lat", vehicle.LocationLat,
		"location_lng", vehicle.LocationLng)

	if err := h.fleetService.RegisterVehicle(r.Context(), &vehicle); err != nil {
		slog.Error("Vehicle registration failed",
			"vehicle_id", vehicle.ID,
			"error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.Info("Vehicle registration successful", "vehicle_id", vehicle.ID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(vehicle)
}

// UpdateVehicleLocation updates a vehicle's position
func (h *HTTPHandler) UpdateVehicleLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vehicleID := vars["id"]

	var locationUpdate struct {
		Lat    float64 `json:"lat"`
		Lng    float64 `json:"lng"`
		Status string  `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&locationUpdate); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.fleetService.UpdateVehicleLocationAndStatus(r.Context(), vehicleID, locationUpdate.Lat, locationUpdate.Lng, locationUpdate.Status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// AssignJob assigns a job to a vehicle
func (h *HTTPHandler) AssignJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vehicleID := vars["id"]

	var jobAssignment struct {
		JobID string `json:"job_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&jobAssignment); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.fleetService.AssignJob(r.Context(), vehicleID, jobAssignment.JobID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CompleteJob marks a job as completed
func (h *HTTPHandler) CompleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vehicleID := vars["id"]

	if err := h.fleetService.CompleteJob(r.Context(), vehicleID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// FindNearestVehicle finds the nearest available vehicle
func (h *HTTPHandler) FindNearestVehicle(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("region")
	latStr := r.URL.Query().Get("pickup_lat")
	lngStr := r.URL.Query().Get("pickup_lng")
	distanceStr := r.URL.Query().Get("trip_distance_km")

	if region == "" || latStr == "" || lngStr == "" || distanceStr == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "Invalid latitude", http.StatusBadRequest)
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		http.Error(w, "Invalid longitude", http.StatusBadRequest)
		return
	}

	distance, err := strconv.ParseFloat(distanceStr, 64)
	if err != nil {
		http.Error(w, "Invalid trip distance", http.StatusBadRequest)
		return
	}

	vehicle, err := h.fleetService.FindNearestAvailableVehicle(r.Context(), region, lat, lng, distance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vehicle)
}
