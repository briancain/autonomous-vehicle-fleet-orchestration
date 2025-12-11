package simulator

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"time"
)

// RoutePoint represents a coordinate point in a route
type RoutePoint struct {
	Lat float64
	Lng float64
}

// Route represents a complete route with waypoints
type Route struct {
	Points   []RoutePoint `json:"points"`
	Distance float64      `json:"distance"` // in meters
	Duration float64      `json:"duration"` // in seconds
}

// OSRMResponse represents the response from OSRM API
type OSRMResponse struct {
	Code   string `json:"code"`
	Routes []struct {
		Geometry struct {
			Coordinates [][]float64 `json:"coordinates"`
		} `json:"geometry"`
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
	} `json:"routes"`
}

// RoutingService handles route calculations
type RoutingService struct {
	client *http.Client
}

// NewRoutingService creates a new routing service
func NewRoutingService() *RoutingService {
	return &RoutingService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetRoute calculates a route between two points using OSRM
func (r *RoutingService) GetRoute(startLat, startLng, endLat, endLng float64) (*Route, error) {
	// OSRM API URL - using public demo server
	url := fmt.Sprintf("http://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		startLng, startLat, endLng, endLat)

	resp, err := r.client.Get(url)
	if err != nil {
		slog.Error("OSRM routing API failed, using straight-line fallback",
			"error", err,
			"start_lat", startLat,
			"start_lng", startLng,
			"end_lat", endLat,
			"end_lng", endLng,
			"url", url)
		// Fallback to straight line if routing fails
		return r.createStraightLineRoute(startLat, startLng, endLat, endLng), nil
	}
	defer resp.Body.Close()

	var osrmResp OSRMResponse
	if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
		slog.Error("OSRM response parsing failed, using straight-line fallback",
			"error", err,
			"status_code", resp.StatusCode,
			"start_lat", startLat,
			"start_lng", startLng,
			"end_lat", endLat,
			"end_lng", endLng)
		// Fallback to straight line if parsing fails
		return r.createStraightLineRoute(startLat, startLng, endLat, endLng), nil
	}

	if len(osrmResp.Routes) == 0 {
		slog.Error("OSRM returned no routes, using straight-line fallback",
			"osrm_code", osrmResp.Code,
			"start_lat", startLat,
			"start_lng", startLng,
			"end_lat", endLat,
			"end_lng", endLng)
		// Fallback to straight line if no routes found
		return r.createStraightLineRoute(startLat, startLng, endLat, endLng), nil
	}

	slog.Info("OSRM routing successful",
		"distance_m", osrmResp.Routes[0].Distance,
		"duration_s", osrmResp.Routes[0].Duration,
		"waypoints", len(osrmResp.Routes[0].Geometry.Coordinates))

	route := osrmResp.Routes[0]
	points := make([]RoutePoint, len(route.Geometry.Coordinates))

	for i, coord := range route.Geometry.Coordinates {
		points[i] = RoutePoint{
			Lat: coord[1], // OSRM returns [lng, lat]
			Lng: coord[0],
		}
	}

	return &Route{
		Points:   points,
		Distance: route.Distance,
		Duration: route.Duration,
	}, nil
}

// createStraightLineRoute creates a fallback straight-line route
func (r *RoutingService) createStraightLineRoute(startLat, startLng, endLat, endLng float64) *Route {
	// Create 10 intermediate points for smooth movement
	points := make([]RoutePoint, 11)

	for i := 0; i <= 10; i++ {
		ratio := float64(i) / 10.0
		lat := startLat + (endLat-startLat)*ratio
		lng := startLng + (endLng-startLng)*ratio
		points[i] = RoutePoint{Lat: lat, Lng: lng}
	}

	// Estimate distance using Haversine formula
	distance := haversineDistance(startLat, startLng, endLat, endLng) * 1000 // convert to meters
	duration := distance / 13.89                                             // assume 50 km/h average speed

	return &Route{
		Points:   points,
		Distance: distance,
		Duration: duration,
	}
}

// haversineDistance calculates distance between two points in kilometers
func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371 // Earth's radius in kilometers

	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLng := (lng2 - lng1) * (math.Pi / 180)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*(math.Pi/180))*math.Cos(lat2*(math.Pi/180))*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
