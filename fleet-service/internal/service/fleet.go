package service

import (
	"context"
	"fmt"
	"math"
	"sort"

	"fleet-service/internal/storage"
)

// FleetService handles fleet management operations
type FleetService struct {
	storage storage.VehicleStorage
}

// NewFleetService creates a new fleet service instance
func NewFleetService(storage storage.VehicleStorage) *FleetService {
	return &FleetService{
		storage: storage,
	}
}

// RegisterVehicle adds a new vehicle to the fleet
func (f *FleetService) RegisterVehicle(ctx context.Context, vehicle *storage.Vehicle) error {
	return f.storage.CreateVehicle(ctx, vehicle)
}

// UpdateVehicleLocationAndStatus updates a vehicle's position and status
func (f *FleetService) UpdateVehicleLocationAndStatus(ctx context.Context, vehicleID string, lat, lng float64, status string) error {
	return f.storage.UpdateVehicleLocationAndStatus(ctx, vehicleID, lat, lng, status)
}

// UpdateVehicleLocation updates a vehicle's position
func (f *FleetService) UpdateVehicleLocation(ctx context.Context, vehicleID string, lat, lng float64) error {
	return f.storage.UpdateVehicleLocation(ctx, vehicleID, lat, lng)
}

// AssignJob assigns a job to a vehicle and updates its status
func (f *FleetService) AssignJob(ctx context.Context, vehicleID, jobID string) error {
	return f.storage.UpdateVehicleStatus(ctx, vehicleID, "busy", &jobID)
}

// CompleteJob marks a vehicle as available after job completion
func (f *FleetService) CompleteJob(ctx context.Context, vehicleID string) error {
	return f.storage.UpdateVehicleStatus(ctx, vehicleID, "available", nil)
}

// FindNearestAvailableVehicle finds the closest available vehicle with sufficient battery
func (f *FleetService) FindNearestAvailableVehicle(ctx context.Context, region string, pickupLat, pickupLng, tripDistanceKm float64) (*storage.Vehicle, error) {
	vehicles, err := f.storage.GetVehiclesByRegionAndStatus(ctx, region, "available")
	if err != nil {
		return nil, err
	}

	var bestVehicle *storage.Vehicle
	var minDistance float64 = math.MaxFloat64

	for _, vehicle := range vehicles {
		// Calculate distance to pickup location
		distanceToPickup := calculateDistance(vehicle.LocationLat, vehicle.LocationLng, pickupLat, pickupLng)

		// Total distance = distance to pickup + trip distance + 20% safety buffer
		totalDistance := (distanceToPickup + tripDistanceKm) * 1.2

		// Check if vehicle has sufficient battery for total journey
		if vehicle.BatteryRangeKm < totalDistance {
			continue
		}

		if distanceToPickup < minDistance {
			minDistance = distanceToPickup
			bestVehicle = vehicle
		}
	}

	if bestVehicle == nil {
		return nil, fmt.Errorf("no available vehicle found with sufficient battery for trip")
	}

	return bestVehicle, nil
}

// GetAllVehicles returns all vehicles for dashboard display
func (f *FleetService) GetAllVehicles(ctx context.Context) ([]*storage.Vehicle, error) {
	vehicles, err := f.storage.GetAllVehicles(ctx)
	if err != nil {
		return nil, err
	}

	// Sort by status first (available, busy, offline), then by ID
	sort.Slice(vehicles, func(i, j int) bool {
		if vehicles[i].Status != vehicles[j].Status {
			// Status priority: available < busy < offline
			statusOrder := map[string]int{"available": 0, "busy": 1, "offline": 2}
			return statusOrder[vehicles[i].Status] < statusOrder[vehicles[j].Status]
		}
		return vehicles[i].ID < vehicles[j].ID
	})

	return vehicles, nil
}

// calculateDistance calculates the distance between two points using Haversine formula
func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371 // Earth's radius in kilometers

	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180

	dlat := lat2Rad - lat1Rad
	dlng := lng2Rad - lng1Rad

	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dlng/2)*math.Sin(dlng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
