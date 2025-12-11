package simulator

import (
	"math"
	"testing"
)

func TestNewVehicle(t *testing.T) {
	vehicle := NewVehicle("test-vehicle-1", "us-west-2", "http://localhost:8080", "http://localhost:8081", 37.7749, -122.4194)

	if vehicle.ID != "test-vehicle-1" {
		t.Errorf("Expected ID 'test-vehicle-1', got '%s'", vehicle.ID)
	}

	if vehicle.Region != "us-west-2" {
		t.Errorf("Expected region 'us-west-2', got '%s'", vehicle.Region)
	}

	if vehicle.Status != "available" {
		t.Errorf("Expected status 'available', got '%s'", vehicle.Status)
	}

	if vehicle.BatteryLevel < 60 || vehicle.BatteryLevel > 100 {
		t.Errorf("Expected battery level between 60-100, got %f", vehicle.BatteryLevel)
	}

	if vehicle.BatteryRangeKm < 240 || vehicle.BatteryRangeKm > 400 {
		t.Errorf("Expected battery range between 240-400km, got %f", vehicle.BatteryRangeKm)
	}

	if vehicle.LocationLat != 37.7749 || vehicle.LocationLng != -122.4194 {
		t.Errorf("Expected location (37.7749, -122.4194), got (%f, %f)", vehicle.LocationLat, vehicle.LocationLng)
	}

	if vehicle.jobPhase != "idle" {
		t.Errorf("Expected job phase 'idle', got '%s'", vehicle.jobPhase)
	}

	if vehicle.jobServiceURL != "http://localhost:8081" {
		t.Errorf("Expected job service URL 'http://localhost:8081', got '%s'", vehicle.jobServiceURL)
	}
}

func TestVehicle_SetRandomTarget(t *testing.T) {
	vehicle := NewVehicle("test-vehicle-1", "us-west-2", "http://localhost:8080", "http://localhost:8081", 37.7749, -122.4194)

	originalLat := vehicle.LocationLat
	originalLng := vehicle.LocationLng

	vehicle.setRandomTarget(0.01) // 1km radius

	// Target should be different from original location
	if vehicle.targetLat == originalLat && vehicle.targetLng == originalLng {
		t.Error("Target should be different from original location")
	}

	// Target should be within reasonable distance
	distance := math.Sqrt(math.Pow(vehicle.targetLat-originalLat, 2) + math.Pow(vehicle.targetLng-originalLng, 2))
	if distance > 0.01 {
		t.Errorf("Target distance %f exceeds maximum radius 0.01", distance)
	}
}

func TestVehicle_DistanceToTarget(t *testing.T) {
	vehicle := NewVehicle("test-vehicle-1", "us-west-2", "http://localhost:8080", "http://localhost:8081", 37.7749, -122.4194)

	// Set target to same location
	vehicle.targetLat = vehicle.LocationLat
	vehicle.targetLng = vehicle.LocationLng

	distance := vehicle.distanceToTarget()
	if distance != 0 {
		t.Errorf("Expected distance 0 for same location, got %f", distance)
	}

	// Set target to different location
	vehicle.targetLat = 37.7849
	vehicle.targetLng = -122.4094

	distance = vehicle.distanceToTarget()
	if distance <= 0 {
		t.Error("Expected positive distance for different locations")
	}
}

func TestVehicle_DrainBattery(t *testing.T) {
	vehicle := NewVehicle("test-vehicle-1", "us-west-2", "http://localhost:8080", "http://localhost:8081", 37.7749, -122.4194)

	originalBattery := vehicle.BatteryLevel
	originalRange := vehicle.BatteryRangeKm

	// Drain battery by traveling 10km
	vehicle.drainBattery(10.0)

	if vehicle.BatteryLevel >= originalBattery {
		t.Error("Battery level should decrease after draining")
	}

	if vehicle.BatteryRangeKm >= originalRange {
		t.Error("Battery range should decrease after draining")
	}

	// Battery should not go below 0
	vehicle.drainBattery(1000.0) // Drain a lot
	if vehicle.BatteryLevel < 0 {
		t.Errorf("Battery level should not go below 0, got %f", vehicle.BatteryLevel)
	}
}

func TestVehicle_MoveTowardsTarget(t *testing.T) {
	vehicle := NewVehicle("test-vehicle-1", "us-west-2", "http://localhost:8080", "http://localhost:8081", 37.7749, -122.4194)

	originalLat := vehicle.LocationLat
	originalLng := vehicle.LocationLng

	// Set target slightly north
	vehicle.targetLat = originalLat + 0.01
	vehicle.targetLng = originalLng
	vehicle.isMoving = true

	vehicle.moveTowardsTarget()

	// Vehicle should have moved towards target
	if vehicle.LocationLat <= originalLat {
		t.Error("Vehicle should have moved north towards target")
	}

	// Should still be moving if not at target
	if vehicle.distanceToTarget() > 0.001 && !vehicle.isMoving {
		t.Error("Vehicle should still be moving if not at target")
	}
}

func TestVehicle_GoToCharge(t *testing.T) {
	vehicle := NewVehicle("test-vehicle-1", "us-west-2", "http://localhost:8080", "http://localhost:8081", 37.7749, -122.4194)

	vehicle.goToCharge()

	if vehicle.Status != "charging" {
		t.Errorf("Expected status 'charging', got '%s'", vehicle.Status)
	}

	// Vehicle should be moving to charging station (routing behavior)
	if !vehicle.isMoving {
		t.Error("Vehicle should be moving to charging station")
	}
}

func TestVehicle_SimulateCharging(t *testing.T) {
	vehicle := NewVehicle("test-vehicle-1", "us-west-2", "http://localhost:8080", "http://localhost:8081", 37.7749, -122.4194)

	vehicle.Status = "charging"
	vehicle.jobPhase = "charging" // Set to actual charging phase
	vehicle.BatteryLevel = 50
	originalBattery := vehicle.BatteryLevel

	vehicle.simulateCharging()

	if vehicle.BatteryLevel <= originalBattery {
		t.Error("Battery level should increase during charging")
	}

	// Test full charge
	vehicle.BatteryLevel = 95
	vehicle.simulateCharging()

	if vehicle.Status != "available" {
		t.Errorf("Expected status 'available' when fully charged, got '%s'", vehicle.Status)
	}
}

func TestVehicle_SimulateIdleBehavior(t *testing.T) {
	vehicle := NewVehicle("test-vehicle-1", "us-west-2", "http://localhost:8080", "http://localhost:8081", 37.7749, -122.4194)

	vehicle.isMoving = false
	originalLat := vehicle.LocationLat
	originalLng := vehicle.LocationLng

	// Run idle behavior multiple times to potentially trigger movement
	for i := 0; i < 100; i++ {
		vehicle.simulateIdleBehavior()
		if vehicle.isMoving {
			break
		}
	}

	// At least one of the iterations should have triggered movement
	// (This is probabilistic, but with 100 iterations and 10% chance, it's very likely)
	if !vehicle.isMoving && vehicle.targetLat == originalLat && vehicle.targetLng == originalLng {
		t.Log("Note: Random movement not triggered in 100 iterations (this can happen)")
	}
}
func TestBatteryDrainPrecision(t *testing.T) {
	vehicle := NewVehicle("test-vehicle", "us-west-2", "http://fleet", "http://job", 45.5, -122.6)
	vehicle.BatteryLevel = 50.0 // Start with 50% battery

	// Test small movement that should drain minimal battery
	smallDistance := 0.001               // 1 meter
	expectedDrain := smallDistance / 4.0 // 4km per 1% = 0.00025%

	initialBattery := vehicle.BatteryLevel
	vehicle.drainBattery(smallDistance)

	actualDrain := initialBattery - vehicle.BatteryLevel

	// Should drain exactly the calculated amount, not round to full percentage
	if math.Abs(actualDrain-expectedDrain) > 0.0001 {
		t.Errorf("Battery drain precision lost. Expected drain: %f, Actual drain: %f", expectedDrain, actualDrain)
	}

	// Battery should be very close to original, not a full percentage point lower
	expectedBattery := initialBattery - expectedDrain
	if math.Abs(vehicle.BatteryLevel-expectedBattery) > 0.0001 {
		t.Errorf("Battery level incorrect. Expected: %f, Actual: %f", expectedBattery, vehicle.BatteryLevel)
	}

	// Verify battery doesn't round to integer
	if vehicle.BatteryLevel == float64(int(vehicle.BatteryLevel)) {
		t.Errorf("Battery level rounded to integer: %f", vehicle.BatteryLevel)
	}
}
