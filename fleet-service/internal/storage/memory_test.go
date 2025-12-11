package storage

import (
	"context"
	"testing"
)

func TestMemoryVehicleStorage_CreateVehicle(t *testing.T) {
	storage := NewMemoryVehicleStorage()
	ctx := context.Background()

	vehicle := &Vehicle{
		ID:             "test-vehicle-1",
		Region:         "us-west-2",
		Status:         "available",
		BatteryLevel:   80,
		BatteryRangeKm: 200.0,
		LocationLat:    37.7749,
		LocationLng:    -122.4194,
		VehicleType:    "sedan",
	}

	err := storage.CreateVehicle(ctx, vehicle)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Try to create the same vehicle again - should fail
	err = storage.CreateVehicle(ctx, vehicle)
	if err == nil {
		t.Fatal("Expected error when creating duplicate vehicle")
	}
}

func TestMemoryVehicleStorage_GetVehicle(t *testing.T) {
	storage := NewMemoryVehicleStorage()
	ctx := context.Background()

	vehicle := &Vehicle{
		ID:             "test-vehicle-1",
		Region:         "us-west-2",
		Status:         "available",
		BatteryLevel:   80,
		BatteryRangeKm: 200.0,
		LocationLat:    37.7749,
		LocationLng:    -122.4194,
		VehicleType:    "sedan",
	}

	// Create vehicle first
	storage.CreateVehicle(ctx, vehicle)

	// Get the vehicle
	retrieved, err := storage.GetVehicle(ctx, "test-vehicle-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrieved.ID != vehicle.ID {
		t.Errorf("Expected ID %s, got %s", vehicle.ID, retrieved.ID)
	}

	// Try to get non-existent vehicle
	_, err = storage.GetVehicle(ctx, "non-existent")
	if err == nil {
		t.Fatal("Expected error when getting non-existent vehicle")
	}
}

func TestMemoryVehicleStorage_UpdateVehicleLocation(t *testing.T) {
	storage := NewMemoryVehicleStorage()
	ctx := context.Background()

	vehicle := &Vehicle{
		ID:             "test-vehicle-1",
		Region:         "us-west-2",
		Status:         "available",
		BatteryLevel:   80,
		BatteryRangeKm: 200.0,
		LocationLat:    37.7749,
		LocationLng:    -122.4194,
		VehicleType:    "sedan",
	}

	storage.CreateVehicle(ctx, vehicle)

	newLat, newLng := 37.7849, -122.4094
	err := storage.UpdateVehicleLocation(ctx, "test-vehicle-1", newLat, newLng)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := storage.GetVehicle(ctx, "test-vehicle-1")
	if updated.LocationLat != newLat || updated.LocationLng != newLng {
		t.Errorf("Expected location (%f, %f), got (%f, %f)", newLat, newLng, updated.LocationLat, updated.LocationLng)
	}
}

func TestMemoryVehicleStorage_UpdateVehicleStatus(t *testing.T) {
	storage := NewMemoryVehicleStorage()
	ctx := context.Background()

	vehicle := &Vehicle{
		ID:             "test-vehicle-1",
		Region:         "us-west-2",
		Status:         "available",
		BatteryLevel:   80,
		BatteryRangeKm: 200.0,
		LocationLat:    37.7749,
		LocationLng:    -122.4194,
		VehicleType:    "sedan",
	}

	storage.CreateVehicle(ctx, vehicle)

	jobID := "job-123"
	err := storage.UpdateVehicleStatus(ctx, "test-vehicle-1", "busy", &jobID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := storage.GetVehicle(ctx, "test-vehicle-1")
	if updated.Status != "busy" {
		t.Errorf("Expected status 'busy', got '%s'", updated.Status)
	}
	if updated.CurrentJobID == nil || *updated.CurrentJobID != jobID {
		t.Errorf("Expected job ID '%s', got %v", jobID, updated.CurrentJobID)
	}
}

func TestMemoryVehicleStorage_GetVehiclesByRegionAndStatus(t *testing.T) {
	storage := NewMemoryVehicleStorage()
	ctx := context.Background()

	vehicles := []*Vehicle{
		{ID: "v1", Region: "us-west-2", Status: "available", BatteryLevel: 80, BatteryRangeKm: 200.0, LocationLat: 37.7749, LocationLng: -122.4194, VehicleType: "sedan"},
		{ID: "v2", Region: "us-west-2", Status: "busy", BatteryLevel: 60, BatteryRangeKm: 150.0, LocationLat: 37.7849, LocationLng: -122.4094, VehicleType: "sedan"},
		{ID: "v3", Region: "us-east-1", Status: "available", BatteryLevel: 90, BatteryRangeKm: 250.0, LocationLat: 40.7128, LocationLng: -74.0060, VehicleType: "sedan"},
		{ID: "v4", Region: "us-west-2", Status: "available", BatteryLevel: 70, BatteryRangeKm: 180.0, LocationLat: 37.7649, LocationLng: -122.4294, VehicleType: "sedan"},
	}

	for _, v := range vehicles {
		storage.CreateVehicle(ctx, v)
	}

	// Get available vehicles in us-west-2
	result, err := storage.GetVehiclesByRegionAndStatus(ctx, "us-west-2", "available")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 vehicles, got %d", len(result))
	}

	// Verify the correct vehicles are returned
	ids := make(map[string]bool)
	for _, v := range result {
		ids[v.ID] = true
	}

	if !ids["v1"] || !ids["v4"] {
		t.Error("Expected vehicles v1 and v4 to be returned")
	}
}
