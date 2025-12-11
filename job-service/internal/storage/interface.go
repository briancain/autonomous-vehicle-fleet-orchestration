package storage

import (
	"context"
	"time"
)

// Job represents a ride or delivery job
type Job struct {
	ID                  string           `json:"id" dynamodbav:"id"`
	JobType             string           `json:"job_type" dynamodbav:"job_type"`
	Status              string           `json:"status" dynamodbav:"status"`
	AssignedVehicleID   *string          `json:"assigned_vehicle_id,omitempty" dynamodbav:"assigned_vehicle_id,omitempty"`
	PickupLat           float64          `json:"pickup_lat" dynamodbav:"pickup_lat"`
	PickupLng           float64          `json:"pickup_lng" dynamodbav:"pickup_lng"`
	DestinationLat      float64          `json:"destination_lat" dynamodbav:"destination_lat"`
	DestinationLng      float64          `json:"destination_lng" dynamodbav:"destination_lng"`
	EstimatedDistanceKm float64          `json:"estimated_distance_km" dynamodbav:"estimated_distance_km"`
	CreatedAt           time.Time        `json:"created_at" dynamodbav:"created_at"`
	AssignedAt          *time.Time       `json:"assigned_at,omitempty" dynamodbav:"assigned_at,omitempty"`
	CompletedAt         *time.Time       `json:"completed_at,omitempty" dynamodbav:"completed_at,omitempty"`
	CustomerID          string           `json:"customer_id" dynamodbav:"customer_id"`
	Region              string           `json:"region" dynamodbav:"region"`
	DeliveryDetails     *DeliveryDetails `json:"delivery_details,omitempty" dynamodbav:"delivery_details,omitempty"`

	// Revenue tracking
	FareAmount   float64 `json:"fare_amount" dynamodbav:"fare_amount"`
	BaseFare     float64 `json:"base_fare" dynamodbav:"base_fare"`
	DistanceFare float64 `json:"distance_fare" dynamodbav:"distance_fare"`
}

// DeliveryDetails contains delivery-specific information
type DeliveryDetails struct {
	RestaurantName string   `json:"restaurant_name" dynamodbav:"restaurant_name"`
	Items          []string `json:"items" dynamodbav:"items"`
	Instructions   string   `json:"instructions" dynamodbav:"instructions"`
}

// JobStorage defines the interface for job data operations
type JobStorage interface {
	// CreateJob adds a new job
	CreateJob(ctx context.Context, job *Job) error

	// GetJob retrieves a job by ID
	GetJob(ctx context.Context, jobID string) (*Job, error)

	// UpdateJob updates an existing job
	UpdateJob(ctx context.Context, job *Job) error

	// GetJobsByStatus finds jobs by status
	GetJobsByStatus(ctx context.Context, status string) ([]*Job, error)

	// GetJobsByVehicle finds jobs assigned to a specific vehicle
	GetJobsByVehicle(ctx context.Context, vehicleID string) ([]*Job, error)

	// GetAllJobs returns all jobs (for dashboard)
	GetAllJobs(ctx context.Context) ([]*Job, error)

	// UpdateJobStatus updates job status and timestamps
	UpdateJobStatus(ctx context.Context, jobID, status string, vehicleID *string) error
}
