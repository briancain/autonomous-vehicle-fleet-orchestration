package storage

import (
	"context"
	"time"
)

// Vehicle represents a vehicle in the fleet
type Vehicle struct {
	ID             string    `json:"id" dynamodbav:"id"`
	Region         string    `json:"region" dynamodbav:"region"`
	Status         string    `json:"status" dynamodbav:"status"` // available, busy, charging, maintenance
	BatteryLevel   int       `json:"battery_level" dynamodbav:"battery_level"`
	BatteryRangeKm float64   `json:"battery_range_km" dynamodbav:"battery_range_km"`
	LocationLat    float64   `json:"location_lat" dynamodbav:"location_lat"`
	LocationLng    float64   `json:"location_lng" dynamodbav:"location_lng"`
	CurrentJobID   *string   `json:"current_job_id,omitempty" dynamodbav:"current_job_id,omitempty"`
	LastUpdated    time.Time `json:"last_updated" dynamodbav:"last_updated"`
	VehicleType    string    `json:"vehicle_type" dynamodbav:"vehicle_type"`
}

// VehicleStorage defines the interface for vehicle data operations
type VehicleStorage interface {
	// CreateVehicle adds a new vehicle to the fleet
	CreateVehicle(ctx context.Context, vehicle *Vehicle) error

	// GetVehicle retrieves a vehicle by ID
	GetVehicle(ctx context.Context, vehicleID string) (*Vehicle, error)

	// UpdateVehicle updates an existing vehicle
	UpdateVehicle(ctx context.Context, vehicle *Vehicle) error

	// GetVehiclesByRegionAndStatus finds vehicles by region and status
	GetVehiclesByRegionAndStatus(ctx context.Context, region, status string) ([]*Vehicle, error)

	// GetAllVehicles returns all vehicles (for dashboard)
	GetAllVehicles(ctx context.Context) ([]*Vehicle, error)

	// UpdateVehicleLocationAndStatus updates location, status and timestamp
	UpdateVehicleLocationAndStatus(ctx context.Context, vehicleID string, lat, lng float64, status string) error

	// UpdateVehicleLocation updates just the location and timestamp
	UpdateVehicleLocation(ctx context.Context, vehicleID string, lat, lng float64) error

	// UpdateVehicleStatus updates status and clears/sets job ID
	UpdateVehicleStatus(ctx context.Context, vehicleID string, status string, jobID *string) error
}
