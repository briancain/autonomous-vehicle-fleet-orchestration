package job

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetAssignedJobs(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jobs" {
			t.Errorf("Expected path '/jobs', got %s", r.URL.Path)
		}

		jobs := []*Job{
			{
				ID:                "job-1",
				JobType:           "ride",
				Status:            "assigned",
				AssignedVehicleID: stringPtr("vehicle-1"),
				PickupLat:         37.7749,
				PickupLng:         -122.4194,
				DestinationLat:    37.7849,
				DestinationLng:    -122.4094,
				CustomerID:        "customer-1",
				Region:            "us-west-2",
			},
			{
				ID:                "job-2",
				JobType:           "delivery",
				Status:            "assigned",
				AssignedVehicleID: stringPtr("vehicle-2"),
				PickupLat:         37.7649,
				PickupLng:         -122.4294,
				DestinationLat:    37.7749,
				DestinationLng:    -122.4194,
				CustomerID:        "customer-2",
				Region:            "us-west-2",
			},
			{
				ID:         "job-3",
				JobType:    "ride",
				Status:     "pending",
				PickupLat:  37.7549,
				PickupLng:  -122.4394,
				CustomerID: "customer-3",
				Region:     "us-west-2",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jobs)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	jobs, err := client.GetAssignedJobs(ctx, "vehicle-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(jobs) != 1 {
		t.Errorf("Expected 1 job for vehicle-1, got %d", len(jobs))
	}

	if jobs[0].ID != "job-1" {
		t.Errorf("Expected job ID 'job-1', got %s", jobs[0].ID)
	}

	if jobs[0].JobType != "ride" {
		t.Errorf("Expected job type 'ride', got %s", jobs[0].JobType)
	}
}

func TestClient_CompleteJob(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/jobs/job-123/complete"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got %s", expectedPath, r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	err := client.CompleteJob(ctx, "job-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestClient_CompleteJob_Error(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	err := client.CompleteJob(ctx, "job-123")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestClient_CreateTestRideJob(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jobs" {
			t.Errorf("Expected path '/jobs', got %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Verify request body
		var jobRequest struct {
			JobType        string  `json:"job_type"`
			CustomerID     string  `json:"customer_id"`
			Region         string  `json:"region"`
			PickupLat      float64 `json:"pickup_lat"`
			PickupLng      float64 `json:"pickup_lng"`
			DestinationLat float64 `json:"destination_lat"`
			DestinationLng float64 `json:"destination_lng"`
		}

		if err := json.NewDecoder(r.Body).Decode(&jobRequest); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if jobRequest.JobType != "ride" {
			t.Errorf("Expected job type 'ride', got %s", jobRequest.JobType)
		}

		if jobRequest.CustomerID != "test-customer" {
			t.Errorf("Expected customer ID 'test-customer', got %s", jobRequest.CustomerID)
		}

		// Return created job
		job := Job{
			ID:                  "job-456",
			JobType:             jobRequest.JobType,
			Status:              "pending",
			PickupLat:           jobRequest.PickupLat,
			PickupLng:           jobRequest.PickupLng,
			DestinationLat:      jobRequest.DestinationLat,
			DestinationLng:      jobRequest.DestinationLng,
			CustomerID:          jobRequest.CustomerID,
			Region:              jobRequest.Region,
			EstimatedDistanceKm: 1.5,
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	job, err := client.CreateTestRideJob(ctx, "test-customer", "us-west-2", 37.7749, -122.4194, 37.7849, -122.4094)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if job.ID != "job-456" {
		t.Errorf("Expected job ID 'job-456', got %s", job.ID)
	}

	if job.JobType != "ride" {
		t.Errorf("Expected job type 'ride', got %s", job.JobType)
	}

	if job.CustomerID != "test-customer" {
		t.Errorf("Expected customer ID 'test-customer', got %s", job.CustomerID)
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
