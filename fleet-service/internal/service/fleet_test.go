package service

import (
	"context"
	"testing"

	"fleet-service/internal/storage"
)

func TestFleetService_FindNearestAvailableVehicle(t *testing.T) {
	// Create in-memory storage and service
	vehicleStorage := storage.NewMemoryVehicleStorage()
	fleetService := NewFleetService(vehicleStorage)
	ctx := context.Background()

	// Create test vehicles
	vehicles := []*storage.Vehicle{
		{
			ID:             "v1",
			Region:         "us-west-2",
			Status:         "available",
			BatteryLevel:   80,
			BatteryRangeKm: 200.0,   // Sufficient for 50km trip
			LocationLat:    37.7749, // San Francisco
			LocationLng:    -122.4194,
			VehicleType:    "sedan",
		},
		{
			ID:             "v2",
			Region:         "us-west-2",
			Status:         "available",
			BatteryLevel:   30,
			BatteryRangeKm: 50.0, // Insufficient for 50km trip (needs 60km with buffer)
			LocationLat:    37.7849,
			LocationLng:    -122.4094,
			VehicleType:    "sedan",
		},
		{
			ID:             "v3",
			Region:         "us-west-2",
			Status:         "available",
			BatteryLevel:   90,
			BatteryRangeKm: 250.0,   // Sufficient but farther away
			LocationLat:    37.8049, // Farther north
			LocationLng:    -122.4394,
			VehicleType:    "sedan",
		},
		{
			ID:             "v4",
			Region:         "us-east-1", // Different region
			Status:         "available",
			BatteryLevel:   100,
			BatteryRangeKm: 300.0,
			LocationLat:    40.7128,
			LocationLng:    -74.0060,
			VehicleType:    "sedan",
		},
	}

	// Register all vehicles
	for _, v := range vehicles {
		err := fleetService.RegisterVehicle(ctx, v)
		if err != nil {
			t.Fatalf("Failed to register vehicle %s: %v", v.ID, err)
		}
	}

	// Test finding nearest vehicle
	pickupLat, pickupLng := 37.7649, -122.4294 // Close to v1
	tripDistance := 50.0                       // 50km trip

	vehicle, err := fleetService.FindNearestAvailableVehicle(ctx, "us-west-2", pickupLat, pickupLng, tripDistance)
	if err != nil {
		t.Fatalf("Expected to find a vehicle, got error: %v", err)
	}

	// Should return v1 (closest with sufficient battery)
	if vehicle.ID != "v1" {
		t.Errorf("Expected vehicle v1, got %s", vehicle.ID)
	}
}

func TestFleetService_FindNearestAvailableVehicle_NoBatteryCapacity(t *testing.T) {
	vehicleStorage := storage.NewMemoryVehicleStorage()
	fleetService := NewFleetService(vehicleStorage)
	ctx := context.Background()

	// Create vehicle with insufficient battery
	vehicle := &storage.Vehicle{
		ID:             "v1",
		Region:         "us-west-2",
		Status:         "available",
		BatteryLevel:   20,
		BatteryRangeKm: 50.0, // Insufficient for total journey (distance to pickup + 50km trip + 20% buffer ≈ 60km+)
		LocationLat:    37.7749,
		LocationLng:    -122.4194,
		VehicleType:    "sedan",
	}

	fleetService.RegisterVehicle(ctx, vehicle)

	pickupLat, pickupLng := 37.7649, -122.4294
	tripDistance := 50.0

	_, err := fleetService.FindNearestAvailableVehicle(ctx, "us-west-2", pickupLat, pickupLng, tripDistance)
	if err == nil {
		t.Fatal("Expected error when no vehicle has sufficient battery")
	}
}

func TestFleetService_AssignAndCompleteJob(t *testing.T) {
	vehicleStorage := storage.NewMemoryVehicleStorage()
	fleetService := NewFleetService(vehicleStorage)
	ctx := context.Background()

	vehicle := &storage.Vehicle{
		ID:             "v1",
		Region:         "us-west-2",
		Status:         "available",
		BatteryLevel:   80,
		BatteryRangeKm: 200.0,
		LocationLat:    37.7749,
		LocationLng:    -122.4194,
		VehicleType:    "sedan",
	}

	fleetService.RegisterVehicle(ctx, vehicle)

	// Assign job
	jobID := "job-123"
	err := fleetService.AssignJob(ctx, "v1", jobID)
	if err != nil {
		t.Fatalf("Failed to assign job: %v", err)
	}

	// Verify vehicle status changed
	updated, _ := vehicleStorage.GetVehicle(ctx, "v1")
	if updated.Status != "busy" {
		t.Errorf("Expected status 'busy', got '%s'", updated.Status)
	}
	if updated.CurrentJobID == nil || *updated.CurrentJobID != jobID {
		t.Errorf("Expected job ID '%s', got %v", jobID, updated.CurrentJobID)
	}

	// Complete job
	err = fleetService.CompleteJob(ctx, "v1")
	if err != nil {
		t.Fatalf("Failed to complete job: %v", err)
	}

	// Verify vehicle is available again
	completed, _ := vehicleStorage.GetVehicle(ctx, "v1")
	if completed.Status != "available" {
		t.Errorf("Expected status 'available', got '%s'", completed.Status)
	}
	if completed.CurrentJobID != nil {
		t.Errorf("Expected no job ID, got %v", completed.CurrentJobID)
	}
}

func TestFleetService_UpdateVehicleLocation(t *testing.T) {
	vehicleStorage := storage.NewMemoryVehicleStorage()
	fleetService := NewFleetService(vehicleStorage)
	ctx := context.Background()

	vehicle := &storage.Vehicle{
		ID:             "v1",
		Region:         "us-west-2",
		Status:         "available",
		BatteryLevel:   80,
		BatteryRangeKm: 200.0,
		LocationLat:    37.7749,
		LocationLng:    -122.4194,
		VehicleType:    "sedan",
	}

	fleetService.RegisterVehicle(ctx, vehicle)

	// Update location
	newLat, newLng := 37.7849, -122.4094
	err := fleetService.UpdateVehicleLocation(ctx, "v1", newLat, newLng)
	if err != nil {
		t.Fatalf("Failed to update location: %v", err)
	}

	// Verify location was updated
	updated, _ := vehicleStorage.GetVehicle(ctx, "v1")
	if updated.LocationLat != newLat || updated.LocationLng != newLng {
		t.Errorf("Expected location (%f, %f), got (%f, %f)", newLat, newLng, updated.LocationLat, updated.LocationLng)
	}
}

// Test the distance calculation function
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
