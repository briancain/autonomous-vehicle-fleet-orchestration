package service

import (
	"context"
	"testing"

	"job-service/internal/fleet"
	"job-service/internal/storage"
)

// MockFleetClient implements fleet.FleetClient interface for testing
type MockFleetClient struct {
	vehicles    map[string]*fleet.Vehicle
	assignments map[string]string // vehicleID -> jobID
}

func NewMockFleetClient() *MockFleetClient {
	return &MockFleetClient{
		vehicles:    make(map[string]*fleet.Vehicle),
		assignments: make(map[string]string),
	}
}

func (m *MockFleetClient) AddVehicle(vehicle *fleet.Vehicle) {
	m.vehicles[vehicle.ID] = vehicle
}

func (m *MockFleetClient) FindNearestVehicle(ctx context.Context, region string, pickupLat, pickupLng, tripDistanceKm float64) (*fleet.Vehicle, error) {
	// Simple mock: return first available vehicle with sufficient battery
	for _, vehicle := range m.vehicles {
		if vehicle.Region == region && vehicle.Status == "available" && vehicle.BatteryRangeKm >= tripDistanceKm*1.2 {
			return vehicle, nil
		}
	}
	return nil, fleet.ErrNoVehicleAvailable
}

func (m *MockFleetClient) AssignJob(ctx context.Context, vehicleID, jobID string) error {
	if vehicle, exists := m.vehicles[vehicleID]; exists {
		vehicle.Status = "busy"
		vehicle.CurrentJobID = &jobID
		m.assignments[vehicleID] = jobID
		return nil
	}
	return fleet.ErrVehicleNotFound
}

func (m *MockFleetClient) GetAllVehicles(ctx context.Context) ([]*fleet.Vehicle, error) {
	var result []*fleet.Vehicle
	for _, vehicle := range m.vehicles {
		result = append(result, vehicle)
	}
	return result, nil
}

// Define mock errors
var (
	ErrNoVehicleAvailable = fleet.ErrNoVehicleAvailable
	ErrVehicleNotFound    = fleet.ErrVehicleNotFound
)

func TestJobService_CreateRideJob(t *testing.T) {
	jobStorage := storage.NewMemoryJobStorage()
	mockFleetClient := NewMockFleetClient()
	jobService := NewJobService(jobStorage, mockFleetClient)
	ctx := context.Background()

	// Add a mock vehicle
	vehicle := &fleet.Vehicle{
		ID:             "vehicle-1",
		Region:         "us-west-2",
		Status:         "available",
		BatteryLevel:   80,
		BatteryRangeKm: 200.0,
		LocationLat:    37.7749,
		LocationLng:    -122.4194,
		VehicleType:    "sedan",
	}
	mockFleetClient.AddVehicle(vehicle)

	job, err := jobService.CreateRideJob(
		ctx,
		"customer-123",
		"us-west-2",
		37.7749, -122.4194, // pickup
		37.7849, -122.4094, // destination
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if job.JobType != "ride" {
		t.Errorf("Expected job type 'ride', got %s", job.JobType)
	}

	if job.CustomerID != "customer-123" {
		t.Errorf("Expected customer ID 'customer-123', got %s", job.CustomerID)
	}

	if job.Status != "assigned" {
		t.Errorf("Expected status 'assigned', got %s", job.Status)
	}

	if job.AssignedVehicleID == nil || *job.AssignedVehicleID != "vehicle-1" {
		t.Errorf("Expected assigned vehicle 'vehicle-1', got %v", job.AssignedVehicleID)
	}

	// Verify distance calculation
	if job.EstimatedDistanceKm <= 0 {
		t.Error("Expected positive distance calculation")
	}

	// Verify pricing calculation
	if job.FareAmount <= 0 {
		t.Error("Expected positive fare amount")
	}

	if job.BaseFare <= 0 {
		t.Error("Expected positive base fare")
	}

	// For rides, distance fare should be positive
	if job.DistanceFare <= 0 {
		t.Error("Expected positive distance fare for ride")
	}
}

func TestJobService_CreateDeliveryJob(t *testing.T) {
	jobStorage := storage.NewMemoryJobStorage()
	mockFleetClient := NewMockFleetClient()
	jobService := NewJobService(jobStorage, mockFleetClient)
	ctx := context.Background()

	// Add a mock vehicle
	vehicle := &fleet.Vehicle{
		ID:             "vehicle-1",
		Region:         "us-west-2",
		Status:         "available",
		BatteryLevel:   90,
		BatteryRangeKm: 250.0,
		LocationLat:    37.7749,
		LocationLng:    -122.4194,
		VehicleType:    "sedan",
	}
	mockFleetClient.AddVehicle(vehicle)

	deliveryDetails := &storage.DeliveryDetails{
		RestaurantName: "Pizza Palace",
		Items:          []string{"Large Pizza", "Garlic Bread"},
		Instructions:   "Leave at door",
	}

	job, err := jobService.CreateDeliveryJob(
		ctx,
		"customer-456",
		"us-west-2",
		37.7749, -122.4194, // pickup
		37.7849, -122.4094, // destination
		deliveryDetails,
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if job.JobType != "delivery" {
		t.Errorf("Expected job type 'delivery', got %s", job.JobType)
	}

	if job.DeliveryDetails == nil {
		t.Fatal("Expected delivery details, got nil")
	}

	if job.DeliveryDetails.RestaurantName != "Pizza Palace" {
		t.Errorf("Expected restaurant 'Pizza Palace', got %s", job.DeliveryDetails.RestaurantName)
	}

	if job.Status != "assigned" {
		t.Errorf("Expected status 'assigned', got %s", job.Status)
	}

	// Verify pricing calculation for delivery
	if job.FareAmount <= 0 {
		t.Error("Expected positive fare amount")
	}

	if job.BaseFare <= 0 {
		t.Error("Expected positive base fare")
	}

	// For deliveries, distance fare should be 0 (flat rate)
	if job.DistanceFare != 0 {
		t.Error("Expected zero distance fare for delivery (flat rate)")
	}
}

func TestJobService_CreateJobNoVehicleAvailable(t *testing.T) {
	jobStorage := storage.NewMemoryJobStorage()
	mockFleetClient := NewMockFleetClient()
	jobService := NewJobService(jobStorage, mockFleetClient)
	ctx := context.Background()

	// No vehicles available
	job, err := jobService.CreateRideJob(
		ctx,
		"customer-123",
		"us-west-2",
		37.7749, -122.4194,
		37.7849, -122.4094,
	)

	if err != nil {
		t.Fatalf("Expected no error creating job, got %v", err)
	}

	// Job should be created but remain pending
	if job.Status != "pending" {
		t.Errorf("Expected status 'pending', got %s", job.Status)
	}

	if job.AssignedVehicleID != nil {
		t.Errorf("Expected no assigned vehicle, got %v", job.AssignedVehicleID)
	}
}

func TestJobService_ProcessPendingJobs(t *testing.T) {
	jobStorage := storage.NewMemoryJobStorage()
	mockFleetClient := NewMockFleetClient()
	jobService := NewJobService(jobStorage, mockFleetClient)
	ctx := context.Background()

	// Create a pending job first (no vehicles available)
	job, _ := jobService.CreateRideJob(
		ctx,
		"customer-123",
		"us-west-2",
		37.7749, -122.4194,
		37.7849, -122.4094,
	)

	// Verify it's pending
	if job.Status != "pending" {
		t.Errorf("Expected status 'pending', got %s", job.Status)
	}

	// Now add a vehicle
	vehicle := &fleet.Vehicle{
		ID:             "vehicle-1",
		Region:         "us-west-2",
		Status:         "available",
		BatteryLevel:   80,
		BatteryRangeKm: 200.0,
		LocationLat:    37.7749,
		LocationLng:    -122.4194,
		VehicleType:    "sedan",
	}
	mockFleetClient.AddVehicle(vehicle)

	// Process pending jobs
	err := jobService.ProcessPendingJobs(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify job is now assigned
	updatedJob, _ := jobService.GetJob(ctx, job.ID)
	if updatedJob.Status != "assigned" {
		t.Errorf("Expected status 'assigned', got %s", updatedJob.Status)
	}

	if updatedJob.AssignedVehicleID == nil || *updatedJob.AssignedVehicleID != "vehicle-1" {
		t.Errorf("Expected assigned vehicle 'vehicle-1', got %v", updatedJob.AssignedVehicleID)
	}
}

func TestJobService_CompleteJob(t *testing.T) {
	jobStorage := storage.NewMemoryJobStorage()
	mockFleetClient := NewMockFleetClient()
	jobService := NewJobService(jobStorage, mockFleetClient)
	ctx := context.Background()

	// Add a mock vehicle and create an assigned job
	vehicle := &fleet.Vehicle{
		ID:             "vehicle-1",
		Region:         "us-west-2",
		Status:         "available",
		BatteryLevel:   80,
		BatteryRangeKm: 200.0,
		LocationLat:    37.7749,
		LocationLng:    -122.4194,
		VehicleType:    "sedan",
	}
	mockFleetClient.AddVehicle(vehicle)

	job, _ := jobService.CreateRideJob(
		ctx,
		"customer-123",
		"us-west-2",
		37.7749, -122.4194,
		37.7849, -122.4094,
	)

	// Complete the job
	err := jobService.CompleteJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify job is completed
	completedJob, _ := jobService.GetJob(ctx, job.ID)
	if completedJob.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", completedJob.Status)
	}

	if completedJob.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
}

func TestJobService_CompleteJobInvalidStatus(t *testing.T) {
	jobStorage := storage.NewMemoryJobStorage()
	mockFleetClient := NewMockFleetClient()
	jobService := NewJobService(jobStorage, mockFleetClient)
	ctx := context.Background()

	// Create a pending job (not assigned)
	job, _ := jobService.CreateRideJob(
		ctx,
		"customer-123",
		"us-west-2",
		37.7749, -122.4194,
		37.7849, -122.4094,
	)

	// Try to complete pending job - should fail
	err := jobService.CompleteJob(ctx, job.ID)
	if err == nil {
		t.Fatal("Expected error when completing pending job")
	}
}

func TestCalculateDistance(t *testing.T) {
	// Test distance between San Francisco and Los Angeles (approximately 560km)
	sfLat, sfLng := 37.7749, -122.4194
	laLat, laLng := 34.0522, -118.2437

	distance := calculateDistance(sfLat, sfLng, laLat, laLng)

	// Should be approximately 560km (allow some tolerance)
	if distance < 500 || distance > 600 {
		t.Errorf("Expected distance around 560km, got %f", distance)
	}

	// Test distance between same points (should be 0)
	sameDistance := calculateDistance(sfLat, sfLng, sfLat, sfLng)
	if sameDistance > 0.001 { // Allow tiny floating point errors
		t.Errorf("Expected distance 0, got %f", sameDistance)
	}
}

func TestJobService_GetActiveJobCount(t *testing.T) {
	// Setup
	memStorage := storage.NewMemoryJobStorage()
	mockFleet := NewMockFleetClient()
	jobService := NewJobService(memStorage, mockFleet)
	ctx := context.Background()

	// Test with no jobs
	count, err := jobService.GetActiveJobCount()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 active jobs, got %d", count)
	}

	// Create jobs without vehicles (will be pending)
	job1, _ := jobService.CreateRideJob(ctx, "customer-1", "us-west-2", 37.7749, -122.4194, 37.7849, -122.4094)
	job2, _ := jobService.CreateRideJob(ctx, "customer-2", "us-west-2", 37.7749, -122.4194, 37.7849, -122.4094)

	// Both jobs should be pending (active)
	count, err = jobService.GetActiveJobCount()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 active jobs, got %d", count)
	}

	// Complete one job directly
	memStorage.UpdateJobStatus(ctx, job1.ID, "completed", nil)

	// Count should be 1 (only job2 is active)
	count, err = jobService.GetActiveJobCount()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 active job, got %d", count)
	}

	// Complete second job directly
	memStorage.UpdateJobStatus(ctx, job2.ID, "completed", nil)

	// Count should be 0
	count, err = jobService.GetActiveJobCount()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 active jobs, got %d", count)
	}
}
