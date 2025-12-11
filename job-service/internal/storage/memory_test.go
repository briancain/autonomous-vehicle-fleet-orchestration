package storage

import (
	"context"
	"testing"
)

func TestMemoryJobStorage_CreateJob(t *testing.T) {
	storage := NewMemoryJobStorage()
	ctx := context.Background()

	job := &Job{
		ID:                  "test-job-1",
		JobType:             "ride",
		Status:              "pending",
		PickupLat:           37.7749,
		PickupLng:           -122.4194,
		DestinationLat:      37.7849,
		DestinationLng:      -122.4094,
		EstimatedDistanceKm: 1.5,
		CustomerID:          "customer-123",
		Region:              "us-west-2",
	}

	err := storage.CreateJob(ctx, job)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Try to create the same job again - should fail
	err = storage.CreateJob(ctx, job)
	if err == nil {
		t.Fatal("Expected error when creating duplicate job")
	}
}

func TestMemoryJobStorage_GetJob(t *testing.T) {
	storage := NewMemoryJobStorage()
	ctx := context.Background()

	job := &Job{
		ID:                  "test-job-1",
		JobType:             "delivery",
		Status:              "pending",
		PickupLat:           37.7749,
		PickupLng:           -122.4194,
		DestinationLat:      37.7849,
		DestinationLng:      -122.4094,
		EstimatedDistanceKm: 1.5,
		CustomerID:          "customer-123",
		Region:              "us-west-2",
		DeliveryDetails: &DeliveryDetails{
			RestaurantName: "Test Restaurant",
			Items:          []string{"Pizza", "Soda"},
			Instructions:   "Ring doorbell",
		},
	}

	// Create job first
	storage.CreateJob(ctx, job)

	// Get the job
	retrieved, err := storage.GetJob(ctx, "test-job-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrieved.ID != job.ID {
		t.Errorf("Expected ID %s, got %s", job.ID, retrieved.ID)
	}

	if retrieved.JobType != "delivery" {
		t.Errorf("Expected job type 'delivery', got %s", retrieved.JobType)
	}

	if retrieved.DeliveryDetails == nil {
		t.Fatal("Expected delivery details, got nil")
	}

	if retrieved.DeliveryDetails.RestaurantName != "Test Restaurant" {
		t.Errorf("Expected restaurant name 'Test Restaurant', got %s", retrieved.DeliveryDetails.RestaurantName)
	}

	// Try to get non-existent job
	_, err = storage.GetJob(ctx, "non-existent")
	if err == nil {
		t.Fatal("Expected error when getting non-existent job")
	}
}

func TestMemoryJobStorage_UpdateJobStatus(t *testing.T) {
	storage := NewMemoryJobStorage()
	ctx := context.Background()

	job := &Job{
		ID:                  "test-job-1",
		JobType:             "ride",
		Status:              "pending",
		PickupLat:           37.7749,
		PickupLng:           -122.4194,
		DestinationLat:      37.7849,
		DestinationLng:      -122.4094,
		EstimatedDistanceKm: 1.5,
		CustomerID:          "customer-123",
		Region:              "us-west-2",
	}

	storage.CreateJob(ctx, job)

	vehicleID := "vehicle-123"
	err := storage.UpdateJobStatus(ctx, "test-job-1", "assigned", &vehicleID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := storage.GetJob(ctx, "test-job-1")
	if updated.Status != "assigned" {
		t.Errorf("Expected status 'assigned', got '%s'", updated.Status)
	}
	if updated.AssignedVehicleID == nil || *updated.AssignedVehicleID != vehicleID {
		t.Errorf("Expected vehicle ID '%s', got %v", vehicleID, updated.AssignedVehicleID)
	}
	if updated.AssignedAt == nil {
		t.Error("Expected AssignedAt to be set")
	}

	// Test completion
	err = storage.UpdateJobStatus(ctx, "test-job-1", "completed", &vehicleID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	completed, _ := storage.GetJob(ctx, "test-job-1")
	if completed.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", completed.Status)
	}
	if completed.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
}

func TestMemoryJobStorage_GetJobsByStatus(t *testing.T) {
	storage := NewMemoryJobStorage()
	ctx := context.Background()

	jobs := []*Job{
		{ID: "job1", JobType: "ride", Status: "pending", PickupLat: 37.7749, PickupLng: -122.4194, DestinationLat: 37.7849, DestinationLng: -122.4094, EstimatedDistanceKm: 1.5, CustomerID: "customer1", Region: "us-west-2"},
		{ID: "job2", JobType: "delivery", Status: "assigned", PickupLat: 37.7649, PickupLng: -122.4294, DestinationLat: 37.7749, DestinationLng: -122.4194, EstimatedDistanceKm: 1.2, CustomerID: "customer2", Region: "us-west-2"},
		{ID: "job3", JobType: "ride", Status: "pending", PickupLat: 37.7549, PickupLng: -122.4394, DestinationLat: 37.7649, DestinationLng: -122.4294, EstimatedDistanceKm: 1.8, CustomerID: "customer3", Region: "us-west-2"},
		{ID: "job4", JobType: "delivery", Status: "completed", PickupLat: 37.7449, PickupLng: -122.4494, DestinationLat: 37.7549, DestinationLng: -122.4394, EstimatedDistanceKm: 2.1, CustomerID: "customer4", Region: "us-west-2"},
	}

	for _, j := range jobs {
		storage.CreateJob(ctx, j)
	}

	// Get pending jobs
	pendingJobs, err := storage.GetJobsByStatus(ctx, "pending")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(pendingJobs) != 2 {
		t.Errorf("Expected 2 pending jobs, got %d", len(pendingJobs))
	}

	// Verify the correct jobs are returned
	ids := make(map[string]bool)
	for _, j := range pendingJobs {
		ids[j.ID] = true
	}

	if !ids["job1"] || !ids["job3"] {
		t.Error("Expected jobs job1 and job3 to be returned")
	}
}

func TestMemoryJobStorage_GetJobsByVehicle(t *testing.T) {
	storage := NewMemoryJobStorage()
	ctx := context.Background()

	vehicleID := "vehicle-123"
	jobs := []*Job{
		{ID: "job1", JobType: "ride", Status: "assigned", AssignedVehicleID: &vehicleID, PickupLat: 37.7749, PickupLng: -122.4194, DestinationLat: 37.7849, DestinationLng: -122.4094, EstimatedDistanceKm: 1.5, CustomerID: "customer1", Region: "us-west-2"},
		{ID: "job2", JobType: "delivery", Status: "pending", PickupLat: 37.7649, PickupLng: -122.4294, DestinationLat: 37.7749, DestinationLng: -122.4194, EstimatedDistanceKm: 1.2, CustomerID: "customer2", Region: "us-west-2"},
		{ID: "job3", JobType: "ride", Status: "completed", AssignedVehicleID: &vehicleID, PickupLat: 37.7549, PickupLng: -122.4394, DestinationLat: 37.7649, DestinationLng: -122.4294, EstimatedDistanceKm: 1.8, CustomerID: "customer3", Region: "us-west-2"},
	}

	for _, j := range jobs {
		storage.CreateJob(ctx, j)
	}

	// Get jobs for specific vehicle
	vehicleJobs, err := storage.GetJobsByVehicle(ctx, vehicleID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(vehicleJobs) != 2 {
		t.Errorf("Expected 2 jobs for vehicle, got %d", len(vehicleJobs))
	}

	// Verify the correct jobs are returned
	ids := make(map[string]bool)
	for _, j := range vehicleJobs {
		ids[j.ID] = true
	}

	if !ids["job1"] || !ids["job3"] {
		t.Error("Expected jobs job1 and job3 to be returned")
	}
}
