package handlers

import (
	"encoding/json"
	"net/http"

	"job-service/internal/service"
	"job-service/internal/storage"

	"github.com/gorilla/mux"
)

// HTTPHandler handles HTTP requests for the job service
type HTTPHandler struct {
	jobService *service.JobService
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(jobService *service.JobService) *HTTPHandler {
	return &HTTPHandler{
		jobService: jobService,
	}
}

// RegisterRoutes sets up HTTP routes
func (h *HTTPHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/health", h.Health).Methods("GET")
	router.HandleFunc("/jobs", h.GetAllJobs).Methods("GET")
	router.HandleFunc("/jobs", h.CreateJob).Methods("POST")
	router.HandleFunc("/jobs/{id}", h.GetJob).Methods("GET")
	router.HandleFunc("/jobs/{id}/complete", h.CompleteJob).Methods("POST")
	router.HandleFunc("/jobs/status/{status}", h.GetJobsByStatus).Methods("GET")
	router.HandleFunc("/jobs/process-pending", h.ProcessPendingJobs).Methods("POST")
	router.HandleFunc("/revenue", h.GetRevenue).Methods("GET")
}

// Health returns service health status
func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// CreateJobRequest represents a job creation request
type CreateJobRequest struct {
	JobType         string                   `json:"job_type"` // "ride" or "delivery"
	CustomerID      string                   `json:"customer_id"`
	Region          string                   `json:"region"`
	PickupLat       float64                  `json:"pickup_lat"`
	PickupLng       float64                  `json:"pickup_lng"`
	DestinationLat  float64                  `json:"destination_lat"`
	DestinationLng  float64                  `json:"destination_lng"`
	DeliveryDetails *storage.DeliveryDetails `json:"delivery_details,omitempty"`
}

// GetAllJobs returns all jobs
func (h *HTTPHandler) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.jobService.GetAllJobs(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

// CreateJob creates a new ride or delivery job
func (h *HTTPHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.JobType == "" || req.CustomerID == "" || req.Region == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	var job *storage.Job
	var err error

	switch req.JobType {
	case "ride":
		job, err = h.jobService.CreateRideJob(
			r.Context(),
			req.CustomerID,
			req.Region,
			req.PickupLat,
			req.PickupLng,
			req.DestinationLat,
			req.DestinationLng,
		)
	case "delivery":
		job, err = h.jobService.CreateDeliveryJob(
			r.Context(),
			req.CustomerID,
			req.Region,
			req.PickupLat,
			req.PickupLng,
			req.DestinationLat,
			req.DestinationLng,
			req.DeliveryDetails,
		)
	default:
		http.Error(w, "Invalid job type. Must be 'ride' or 'delivery'", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

// GetJob retrieves a specific job
func (h *HTTPHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]

	job, err := h.jobService.GetJob(r.Context(), jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// CompleteJob marks a job as completed
func (h *HTTPHandler) CompleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]

	if err := h.jobService.CompleteJob(r.Context(), jobID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetJobsByStatus returns jobs with specific status
func (h *HTTPHandler) GetJobsByStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	status := vars["status"]

	jobs, err := h.jobService.GetJobsByStatus(r.Context(), status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

// ProcessPendingJobs attempts to assign all pending jobs
func (h *HTTPHandler) ProcessPendingJobs(w http.ResponseWriter, r *http.Request) {
	if err := h.jobService.ProcessPendingJobs(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Pending jobs processed"}`))
}

// GetRevenue returns revenue statistics
func (h *HTTPHandler) GetRevenue(w http.ResponseWriter, r *http.Request) {
	revenue, err := h.jobService.GetRevenue(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(revenue)
}
