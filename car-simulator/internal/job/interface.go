package job

import "context"

// JobClient defines the interface for job service operations
type JobClient interface {
	GetAssignedJobs(ctx context.Context, vehicleID string) ([]*Job, error)
	CompleteJob(ctx context.Context, jobID string) error
	CreateTestRideJob(ctx context.Context, customerID, region string, pickupLat, pickupLng, destLat, destLng float64) (*Job, error)
}
