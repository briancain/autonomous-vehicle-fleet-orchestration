package testhelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HTTPClient provides helper methods for making HTTP requests in tests
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient creates a new HTTP client for testing
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{},
	}
}

// Vehicle represents a vehicle from the fleet service
type Vehicle struct {
	ID              string  `json:"id"`
	Region          string  `json:"region"`
	Status          string  `json:"status"`
	BatteryLevel    int     `json:"battery_level"`
	BatteryRangeKm  float64 `json:"battery_range_km"`
	LocationLat     float64 `json:"location_lat"`
	LocationLng     float64 `json:"location_lng"`
	CurrentJobID    *string `json:"current_job_id,omitempty"`
	VehicleType     string  `json:"vehicle_type"`
}

// Job represents a job from the job service
type Job struct {
	ID                  string           `json:"id"`
	JobType             string           `json:"job_type"`
	Status              string           `json:"status"`
	AssignedVehicleID   *string          `json:"assigned_vehicle_id,omitempty"`
	PickupLat           float64          `json:"pickup_lat"`
	PickupLng           float64          `json:"pickup_lng"`
	DestinationLat      float64          `json:"destination_lat"`
	DestinationLng      float64          `json:"destination_lng"`
	EstimatedDistanceKm float64          `json:"estimated_distance_km"`
	CustomerID          string           `json:"customer_id"`
	Region              string           `json:"region"`
	DeliveryDetails     *DeliveryDetails `json:"delivery_details,omitempty"`
	CreatedAt           *string          `json:"created_at,omitempty"`
	AssignedAt          *string          `json:"assigned_at,omitempty"`
	CompletedAt         *string          `json:"completed_at,omitempty"`
}

// DeliveryDetails contains delivery-specific information
type DeliveryDetails struct {
	RestaurantName string   `json:"restaurant_name"`
	Items          []string `json:"items"`
	Instructions   string   `json:"instructions"`
}

// CreateJobRequest represents a job creation request
type CreateJobRequest struct {
	JobType        string           `json:"job_type"`
	CustomerID     string           `json:"customer_id"`
	Region         string           `json:"region"`
	PickupLat      float64          `json:"pickup_lat"`
	PickupLng      float64          `json:"pickup_lng"`
	DestinationLat float64          `json:"destination_lat"`
	DestinationLng float64          `json:"destination_lng"`
	DeliveryDetails *DeliveryDetails `json:"delivery_details,omitempty"`
}

// GetVehicles retrieves all vehicles from the fleet service
func (c *HTTPClient) GetVehicles() ([]*Vehicle, error) {
	resp, err := c.client.Get("http://localhost:8080/vehicles")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fleet service returned status %d", resp.StatusCode)
	}

	var vehicles []*Vehicle
	if err := json.NewDecoder(resp.Body).Decode(&vehicles); err != nil {
		return nil, err
	}

	return vehicles, nil
}

// GetJobs retrieves all jobs from the job service
func (c *HTTPClient) GetJobs() ([]*Job, error) {
	resp, err := c.client.Get("http://localhost:8081/jobs")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("job service returned status %d", resp.StatusCode)
	}

	var jobs []*Job
	if err := json.NewDecoder(resp.Body).Decode(&jobs); err != nil {
		return nil, err
	}

	return jobs, nil
}

// CreateRideJob creates a new ride job
func (c *HTTPClient) CreateRideJob(customerID, region string, pickupLat, pickupLng, destLat, destLng float64) (*Job, error) {
	jobRequest := CreateJobRequest{
		JobType:        "ride",
		CustomerID:     customerID,
		Region:         region,
		PickupLat:      pickupLat,
		PickupLng:      pickupLng,
		DestinationLat: destLat,
		DestinationLng: destLng,
	}

	return c.createJob(jobRequest)
}

// CreateDeliveryJob creates a new delivery job
func (c *HTTPClient) CreateDeliveryJob(customerID, region string, pickupLat, pickupLng, destLat, destLng float64, details *DeliveryDetails) (*Job, error) {
	jobRequest := CreateJobRequest{
		JobType:         "delivery",
		CustomerID:      customerID,
		Region:          region,
		PickupLat:       pickupLat,
		PickupLng:       pickupLng,
		DestinationLat:  destLat,
		DestinationLng:  destLng,
		DeliveryDetails: details,
	}

	return c.createJob(jobRequest)
}

// createJob creates a job via the job service API
func (c *HTTPClient) createJob(jobRequest CreateJobRequest) (*Job, error) {
	jsonData, err := json.Marshal(jobRequest)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post("http://localhost:8081/jobs", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("job service returned status %d: %s", resp.StatusCode, string(body))
	}

	var job Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, err
	}

	return &job, nil
}

// GetJob retrieves a specific job by ID
func (c *HTTPClient) GetJob(jobID string) (*Job, error) {
	resp, err := c.client.Get(fmt.Sprintf("http://localhost:8081/jobs/%s", jobID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("job service returned status %d", resp.StatusCode)
	}

	var job Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, err
	}

	return &job, nil
}

// parseJSONResponse is a helper function to parse JSON responses
func parseJSONResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}
