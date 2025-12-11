package fleet

import "context"

// FleetClient defines the interface for fleet service operations
type FleetClient interface {
	FindNearestVehicle(ctx context.Context, region string, pickupLat, pickupLng, tripDistanceKm float64) (*Vehicle, error)
	AssignJob(ctx context.Context, vehicleID, jobID string) error
	GetAllVehicles(ctx context.Context) ([]*Vehicle, error)
}
