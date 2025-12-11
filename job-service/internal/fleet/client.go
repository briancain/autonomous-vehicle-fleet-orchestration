package fleet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Common errors
var (
	ErrNoVehicleAvailable = errors.New("no vehicle available")
	ErrVehicleNotFound    = errors.New("vehicle not found")
)

// Vehicle represents a vehicle from the fleet service
type Vehicle struct {
	ID             string  `json:"id"`
	Region         string  `json:"region"`
	Status         string  `json:"status"`
	BatteryLevel   int     `json:"battery_level"`
	BatteryRangeKm float64 `json:"battery_range_km"`
	LocationLat    float64 `json:"location_lat"`
	LocationLng    float64 `json:"location_lng"`
	CurrentJobID   *string `json:"current_job_id,omitempty"`
	VehicleType    string  `json:"vehicle_type"`
}

// Client handles communication with the Fleet Service
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new fleet service client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FindNearestVehicle finds the nearest available vehicle for a job
func (c *Client) FindNearestVehicle(ctx context.Context, region string, pickupLat, pickupLng, tripDistanceKm float64) (*Vehicle, error) {
	params := url.Values{}
	params.Add("region", region)
	params.Add("pickup_lat", strconv.FormatFloat(pickupLat, 'f', 6, 64))
	params.Add("pickup_lng", strconv.FormatFloat(pickupLng, 'f', 6, 64))
	params.Add("trip_distance_km", strconv.FormatFloat(tripDistanceKm, 'f', 2, 64))

	url := fmt.Sprintf("%s/vehicles/find?%s", c.baseURL, params.Encode())

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
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrNoVehicleAvailable
		}
		return nil, fmt.Errorf("fleet service returned status %d", resp.StatusCode)
	}

	var vehicle Vehicle
	if err := json.NewDecoder(resp.Body).Decode(&vehicle); err != nil {
		return nil, err
	}

	return &vehicle, nil
}

// AssignJob assigns a job to a vehicle
func (c *Client) AssignJob(ctx context.Context, vehicleID, jobID string) error {
	assignment := struct {
		JobID string `json:"job_id"`
	}{
		JobID: jobID,
	}

	jsonData, err := json.Marshal(assignment)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/vehicles/%s/assign", c.baseURL, vehicleID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return ErrVehicleNotFound
		}
		return fmt.Errorf("failed to assign job, status: %d", resp.StatusCode)
	}

	return nil
}

// GetAllVehicles retrieves all vehicles from the fleet service
func (c *Client) GetAllVehicles(ctx context.Context) ([]*Vehicle, error) {
	url := fmt.Sprintf("%s/vehicles", c.baseURL)

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
		return nil, fmt.Errorf("fleet service returned status %d", resp.StatusCode)
	}

	var vehicles []*Vehicle
	if err := json.NewDecoder(resp.Body).Decode(&vehicles); err != nil {
		return nil, err
	}

	return vehicles, nil
}
