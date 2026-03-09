package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Job represents a job from the job service
type Job struct {
	ID                  string           `json:"id"`
	JobType             string           `json:"job_type"` // "ride", "delivery"
	Status              string           `json:"status"`   // "pending", "assigned", "in_progress", "completed", "failed"
	AssignedVehicleID   *string          `json:"assigned_vehicle_id,omitempty"`
	PickupLat           float64          `json:"pickup_lat"`
	PickupLng           float64          `json:"pickup_lng"`
	DestinationLat      float64          `json:"destination_lat"`
	DestinationLng      float64          `json:"destination_lng"`
	EstimatedDistanceKm float64          `json:"estimated_distance_km"`
	CustomerID          string           `json:"customer_id"`
	Region              string           `json:"region"`
	DeliveryDetails     *DeliveryDetails `json:"delivery_details,omitempty"`
}

// DeliveryDetails contains delivery-specific information
type DeliveryDetails struct {
	RestaurantName string   `json:"restaurant_name"`
	Items          []string `json:"items"`
	Instructions   string   `json:"instructions"`
}

// Client handles communication with the Job Service
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new job service client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetAssignedJobs retrieves jobs assigned to a specific vehicle
func (c *Client) GetAssignedJobs(ctx context.Context, vehicleID string) ([]*Job, error) {
	// Get all jobs and filter by vehicle ID (in a real system, this would be a dedicated endpoint)
	url := fmt.Sprintf("%s/jobs", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("job service returned status %d", resp.StatusCode)
	}

	var allJobs []*Job
	if err := json.NewDecoder(resp.Body).Decode(&allJobs); err != nil {
		return nil, err
	}

	// Filter jobs assigned to this vehicle
	var assignedJobs []*Job
	for _, job := range allJobs {
		if job.AssignedVehicleID != nil && *job.AssignedVehicleID == vehicleID {
			assignedJobs = append(assignedJobs, job)
		}
	}

	return assignedJobs, nil
}

// CompleteJob marks a job as completed
func (c *Client) CompleteJob(ctx context.Context, jobID string) error {
	url := fmt.Sprintf("%s/jobs/%s/complete", c.baseURL, jobID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to complete job, status: %d", resp.StatusCode)
	}

	return nil
}

// CreateTestRideJob creates a test ride job (for testing/demo purposes)
func (c *Client) CreateTestRideJob(ctx context.Context, customerID, region string, pickupLat, pickupLng, destLat, destLng float64) (*Job, error) {
	jobRequest := struct {
		JobType        string  `json:"job_type"`
		CustomerID     string  `json:"customer_id"`
		Region         string  `json:"region"`
		PickupLat      float64 `json:"pickup_lat"`
		PickupLng      float64 `json:"pickup_lng"`
		DestinationLat float64 `json:"destination_lat"`
		DestinationLng float64 `json:"destination_lng"`
	}{
		JobType:        "ride",
		CustomerID:     customerID,
		Region:         region,
		PickupLat:      pickupLat,
		PickupLng:      pickupLng,
		DestinationLat: destLat,
		DestinationLng: destLng,
	}

	jsonData, err := json.Marshal(jobRequest)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/jobs", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create job, status: %d", resp.StatusCode)
	}

	var job Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, err
	}

	return &job, nil
}
