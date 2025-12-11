package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"job-service/internal/fleet"
	"job-service/internal/kinesis"
	"job-service/internal/storage"
)

// JobService handles job management operations
type JobService struct {
	storage     storage.JobStorage
	fleetClient fleet.FleetClient
	pricing     *PricingConfig
	streamer    *kinesis.Streamer
}

// NewJobService creates a new job service instance
func NewJobService(storage storage.JobStorage, fleetClient fleet.FleetClient) *JobService {
	return &JobService{
		storage:     storage,
		fleetClient: fleetClient,
		pricing:     DefaultPricingConfig(),
	}
}

// SetKinesisStreamer sets the Kinesis streamer for job events
func (j *JobService) SetKinesisStreamer(streamer *kinesis.Streamer) {
	j.streamer = streamer
}

// CreateRideJob creates a new ride request
func (j *JobService) CreateRideJob(ctx context.Context, customerID, region string, pickupLat, pickupLng, destLat, destLng float64) (*storage.Job, error) {
	jobID := fmt.Sprintf("ride-%d", generateJobID())

	job := &storage.Job{
		ID:                  jobID,
		JobType:             "ride",
		Status:              "pending",
		PickupLat:           pickupLat,
		PickupLng:           pickupLng,
		DestinationLat:      destLat,
		DestinationLng:      destLng,
		EstimatedDistanceKm: calculateDistance(pickupLat, pickupLng, destLat, destLng),
		CustomerID:          customerID,
		Region:              region,
		CreatedAt:           time.Now(),
	}

	// Calculate pricing
	j.pricing.CalculateFare(job)

	if err := j.storage.CreateJob(ctx, job); err != nil {
		return nil, err
	}

	// Stream job creation event
	if j.streamer != nil {
		j.streamer.StreamJobEvent("created", job)
	}

	// Try to assign immediately
	if err := j.assignJob(ctx, job); err != nil {
		fmt.Printf("Failed to assign job %s immediately: %v\n", jobID, err)
		// Job remains in pending status
	}

	return job, nil
}

// CreateDeliveryJob creates a new delivery request
func (j *JobService) CreateDeliveryJob(ctx context.Context, customerID, region string, pickupLat, pickupLng, destLat, destLng float64, details *storage.DeliveryDetails) (*storage.Job, error) {
	jobID := fmt.Sprintf("delivery-%d", generateJobID())

	job := &storage.Job{
		ID:                  jobID,
		JobType:             "delivery",
		Status:              "pending",
		PickupLat:           pickupLat,
		PickupLng:           pickupLng,
		DestinationLat:      destLat,
		DestinationLng:      destLng,
		EstimatedDistanceKm: calculateDistance(pickupLat, pickupLng, destLat, destLng),
		CustomerID:          customerID,
		Region:              region,
		DeliveryDetails:     details,
		CreatedAt:           time.Now(),
	}

	// Calculate pricing
	j.pricing.CalculateFare(job)

	if err := j.storage.CreateJob(ctx, job); err != nil {
		return nil, err
	}

	// Stream job creation event
	if j.streamer != nil {
		j.streamer.StreamJobEvent("created", job)
	}

	// Try to assign immediately
	if err := j.assignJob(ctx, job); err != nil {
		fmt.Printf("Failed to assign job %s immediately: %v\n", jobID, err)
		// Job remains in pending status
	}

	return job, nil
}

// assignJob attempts to assign a job to an available vehicle
func (j *JobService) assignJob(ctx context.Context, job *storage.Job) error {
	// Find nearest available vehicle
	vehicle, err := j.fleetClient.FindNearestVehicle(ctx, job.Region, job.PickupLat, job.PickupLng, job.EstimatedDistanceKm)
	if err != nil {
		return fmt.Errorf("no available vehicle found: %v", err)
	}

	// Assign job to vehicle in fleet service
	if err := j.fleetClient.AssignJob(ctx, vehicle.ID, job.ID); err != nil {
		return fmt.Errorf("failed to assign job to vehicle: %v", err)
	}

	// Update job status
	if err := j.storage.UpdateJobStatus(ctx, job.ID, "assigned", &vehicle.ID); err != nil {
		return fmt.Errorf("failed to update job status: %v", err)
	}

	// Stream job assignment event
	if j.streamer != nil {
		// Update job object with assigned vehicle for streaming
		job.AssignedVehicleID = &vehicle.ID
		job.Status = "assigned"
		j.streamer.StreamJobEvent("assigned", job)
	}

	fmt.Printf("Job %s assigned to vehicle %s\n", job.ID, vehicle.ID)
	return nil
}

// ProcessPendingJobs attempts to assign all pending jobs
func (j *JobService) ProcessPendingJobs(ctx context.Context) error {
	pendingJobs, err := j.storage.GetJobsByStatus(ctx, "pending")
	if err != nil {
		return err
	}

	for _, job := range pendingJobs {
		if err := j.assignJob(ctx, job); err != nil {
			fmt.Printf("Failed to assign pending job %s: %v\n", job.ID, err)
			continue
		}
	}

	return nil
}

// CompleteJob marks a job as completed
func (j *JobService) CompleteJob(ctx context.Context, jobID string) error {
	job, err := j.storage.GetJob(ctx, jobID)
	if err != nil {
		return err
	}

	if job.Status != "assigned" && job.Status != "in_progress" {
		return fmt.Errorf("job %s is not in progress, current status: %s", jobID, job.Status)
	}

	if err := j.storage.UpdateJobStatus(ctx, jobID, "completed", job.AssignedVehicleID); err != nil {
		return err
	}

	// Stream job completion event
	if j.streamer != nil {
		job.Status = "completed"
		j.streamer.StreamJobEvent("completed", job)
	}

	return nil
}

// GetJob retrieves a job by ID
func (j *JobService) GetJob(ctx context.Context, jobID string) (*storage.Job, error) {
	return j.storage.GetJob(ctx, jobID)
}

// GetAllJobs returns all jobs for dashboard
func (j *JobService) GetAllJobs(ctx context.Context) ([]*storage.Job, error) {
	return j.storage.GetAllJobs(ctx)
}

// GetActiveJobCount returns the count of active jobs (pending + assigned)
func (j *JobService) GetActiveJobCount() (int, error) {
	jobs, err := j.storage.GetAllJobs(context.Background())
	if err != nil {
		return 0, err
	}

	activeCount := 0
	for _, job := range jobs {
		if job.Status == "pending" || job.Status == "assigned" {
			activeCount++
		}
	}

	return activeCount, nil
}

// GetJobsByStatus returns jobs with specific status
func (j *JobService) GetJobsByStatus(ctx context.Context, status string) ([]*storage.Job, error) {
	return j.storage.GetJobsByStatus(ctx, status)
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

// generateJobID generates a simple job ID (in production, use UUID)
var jobCounter int64

func generateJobID() int64 {
	jobCounter++
	return jobCounter
}

// GetRevenue calculates total revenue from completed jobs
func (j *JobService) GetRevenue(ctx context.Context) (map[string]interface{}, error) {
	jobs, err := j.storage.GetAllJobs(ctx)
	if err != nil {
		return nil, err
	}

	var totalRevenue float64
	var rideRevenue float64
	var deliveryRevenue float64
	var completedJobs int
	var rideCount int
	var deliveryCount int

	for _, job := range jobs {
		if job.Status == "completed" {
			totalRevenue += job.FareAmount
			completedJobs++

			if job.JobType == "ride" {
				rideRevenue += job.FareAmount
				rideCount++
			} else {
				deliveryRevenue += job.FareAmount
				deliveryCount++
			}
		}
	}

	avgRideFare := 0.0
	if rideCount > 0 {
		avgRideFare = rideRevenue / float64(rideCount)
	}

	avgDeliveryFare := 0.0
	if deliveryCount > 0 {
		avgDeliveryFare = deliveryRevenue / float64(deliveryCount)
	}

	return map[string]interface{}{
		"total_revenue":     totalRevenue,
		"ride_revenue":      rideRevenue,
		"delivery_revenue":  deliveryRevenue,
		"completed_jobs":    completedJobs,
		"ride_count":        rideCount,
		"delivery_count":    deliveryCount,
		"avg_ride_fare":     avgRideFare,
		"avg_delivery_fare": avgDeliveryFare,
	}, nil
}
