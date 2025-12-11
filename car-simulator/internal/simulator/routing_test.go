package simulator

import (
	"testing"
)

func TestNewRoutingService(t *testing.T) {
	service := NewRoutingService()

	if service == nil {
		t.Error("Expected routing service to be created")
	}

	if service.client == nil {
		t.Error("Expected HTTP client to be initialized")
	}
}

func TestRoutingService_CreateStraightLineRoute(t *testing.T) {
	service := NewRoutingService()

	// Test route from downtown Portland to airport
	startLat, startLng := 45.5152, -122.6784
	endLat, endLng := 45.5898, -122.5951

	route := service.createStraightLineRoute(startLat, startLng, endLat, endLng)

	if route == nil {
		t.Fatal("Expected route to be created")
	}

	if len(route.Points) != 11 {
		t.Errorf("Expected 11 points, got %d", len(route.Points))
	}

	// First point should be start location
	if route.Points[0].Lat != startLat || route.Points[0].Lng != startLng {
		t.Errorf("Expected first point to be start location")
	}

	// Last point should be end location
	lastPoint := route.Points[len(route.Points)-1]
	if lastPoint.Lat != endLat || lastPoint.Lng != endLng {
		t.Errorf("Expected last point to be end location")
	}

	// Distance should be positive
	if route.Distance <= 0 {
		t.Error("Expected positive distance")
	}

	// Duration should be positive
	if route.Duration <= 0 {
		t.Error("Expected positive duration")
	}
}

func TestHaversineDistance(t *testing.T) {
	// Test distance between downtown Portland and airport
	lat1, lng1 := 45.5152, -122.6784
	lat2, lng2 := 45.5898, -122.5951

	distance := haversineDistance(lat1, lng1, lat2, lng2)

	// Distance should be approximately 10-15 km
	if distance < 8 || distance > 20 {
		t.Errorf("Expected distance between 8-20 km, got %.2f", distance)
	}
}

func TestVehicle_SetRouteTarget(t *testing.T) {
	vehicle := NewVehicle("test-vehicle", "us-west-2", "http://localhost:8080", "http://localhost:8081", 45.5152, -122.6784)

	targetLat, targetLng := 45.5898, -122.5951
	vehicle.setRouteTarget(targetLat, targetLng)

	if vehicle.targetLat != targetLat {
		t.Errorf("Expected target lat %.6f, got %.6f", targetLat, vehicle.targetLat)
	}

	if vehicle.targetLng != targetLng {
		t.Errorf("Expected target lng %.6f, got %.6f", targetLng, vehicle.targetLng)
	}

	if !vehicle.isMoving {
		t.Error("Expected vehicle to be moving after setting route target")
	}
}
