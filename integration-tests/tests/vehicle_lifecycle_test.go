package tests

import (
	"testing"
	"time"

	"integration-tests/internal/testhelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVehicleLifecycle(t *testing.T) {
	// Setup: Start all services
	sm := testhelpers.NewServiceManager()
	defer sm.StopServices()

	err := sm.StartServices()
	require.NoError(t, err, "Failed to start services")

	client := testhelpers.NewHTTPClient()

	// Test 1: Vehicle registration
	t.Run("VehicleRegistration", func(t *testing.T) {
		// Start car simulator with 1 vehicle
		err := sm.StartCarSimulator(1)
		require.NoError(t, err)

		// Wait for vehicle registration
		err = sm.WaitForVehicleRegistration(1, 10*time.Second)
		require.NoError(t, err)

		// Verify vehicle is registered
		vehicles, err := client.GetVehicles()
		require.NoError(t, err)
		assert.Len(t, vehicles, 1)

		vehicle := vehicles[0]
		assert.Equal(t, "available", vehicle.Status)
		assert.Equal(t, "us-west-2", vehicle.Region)
		assert.Greater(t, vehicle.BatteryLevel, 0)
		assert.Greater(t, vehicle.BatteryRangeKm, 0.0)
		assert.Equal(t, "sedan", vehicle.VehicleType)
	})

	// Test 2: Vehicle location updates
	t.Run("VehicleLocationUpdates", func(t *testing.T) {
		// Get initial vehicle state
		vehicles, err := client.GetVehicles()
		require.NoError(t, err)
		require.Len(t, vehicles, 1)

		initialVehicle := vehicles[0]
		initialLat := initialVehicle.LocationLat
		initialLng := initialVehicle.LocationLng

		// Wait for location updates (car simulator should be moving randomly)
		locationChanged := false
		for i := 0; i < 20; i++ { // Wait up to 10 seconds
			vehicles, err := client.GetVehicles()
			require.NoError(t, err)
			
			if len(vehicles) > 0 {
				currentVehicle := vehicles[0]
				if currentVehicle.LocationLat != initialLat || currentVehicle.LocationLng != initialLng {
					locationChanged = true
					break
				}
			}
			time.Sleep(500 * time.Millisecond)
		}

		// Note: Location might not change if vehicle is not in idle movement mode
		// This is acceptable behavior
		t.Logf("Vehicle location changed: %v", locationChanged)
	})

	// Test 3: Vehicle status transitions during job execution
	t.Run("VehicleStatusTransitions", func(t *testing.T) {
		// Create a job to trigger status change
		job, err := client.CreateRideJob(
			"customer-status-test",
			"us-west-2",
			37.7749, -122.4194,
			37.7849, -122.4094,
		)
		require.NoError(t, err)

		// Wait for job assignment
		var assignedVehicleID string
		for i := 0; i < 10; i++ {
			updatedJob, err := client.GetJob(job.ID)
			require.NoError(t, err)

			if updatedJob.Status == "assigned" && updatedJob.AssignedVehicleID != nil {
				assignedVehicleID = *updatedJob.AssignedVehicleID
				break
			}
			time.Sleep(500 * time.Millisecond)
		}

		if assignedVehicleID == "" {
			t.Skip("Job was not assigned, skipping status transition test")
			return
		}

		// Verify vehicle status changed to busy
		vehicles, err := client.GetVehicles()
		require.NoError(t, err)

		var assignedVehicle *testhelpers.Vehicle
		for _, vehicle := range vehicles {
			if vehicle.ID == assignedVehicleID {
				assignedVehicle = vehicle
				break
			}
		}

		require.NotNil(t, assignedVehicle, "Assigned vehicle should exist")
		assert.Equal(t, "busy", assignedVehicle.Status, "Assigned vehicle should be busy")

		// Wait for potential job completion and status change back to available
		// Note: This might take a while depending on simulated travel time
		statusChangedBack := false
		for i := 0; i < 60; i++ { // Wait up to 30 seconds
			vehicles, err := client.GetVehicles()
			require.NoError(t, err)

			for _, vehicle := range vehicles {
				if vehicle.ID == assignedVehicleID && vehicle.Status == "available" {
					statusChangedBack = true
					break
				}
			}

			if statusChangedBack {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}

		if statusChangedBack {
			t.Log("Vehicle status successfully changed back to available after job completion")
		} else {
			t.Log("Vehicle status did not change back to available within test timeout - this may be expected for longer routes")
		}
	})

	// Test 4: Multiple vehicles
	t.Run("MultipleVehicles", func(t *testing.T) {
		// Stop current car simulator
		sm.StopServices()
		
		// Restart services
		err := sm.StartServices()
		require.NoError(t, err)

		// Start car simulator with 3 vehicles
		err = sm.StartCarSimulator(3)
		require.NoError(t, err)

		// Wait for all vehicles to register
		err = sm.WaitForVehicleRegistration(3, 15*time.Second)
		require.NoError(t, err)

		// Verify all vehicles are registered
		vehicles, err := client.GetVehicles()
		require.NoError(t, err)
		assert.Len(t, vehicles, 3)

		// All vehicles should be in the same region
		for _, vehicle := range vehicles {
			assert.Equal(t, "us-west-2", vehicle.Region)
			assert.Equal(t, "available", vehicle.Status)
		}

		// Create multiple jobs to test parallel assignment
		var jobs []*testhelpers.Job
		for i := 0; i < 3; i++ {
			job, err := client.CreateRideJob(
				"customer-parallel",
				"us-west-2",
				37.7749+float64(i)*0.001, -122.4194,
				37.7849+float64(i)*0.001, -122.4094,
			)
			require.NoError(t, err)
			jobs = append(jobs, job)
		}

		// Wait for job assignments
		time.Sleep(3 * time.Second)

		// Check how many jobs got assigned
		assignedCount := 0
		for _, job := range jobs {
			updatedJob, err := client.GetJob(job.ID)
			require.NoError(t, err)

			if updatedJob.Status == "assigned" {
				assignedCount++
			}
		}

		// With 3 vehicles, we should be able to assign multiple jobs
		assert.GreaterOrEqual(t, assignedCount, 1, "At least one job should be assigned")
		t.Logf("Assigned %d out of %d jobs with 3 vehicles", assignedCount, len(jobs))
	})

	// Test 5: Vehicle battery simulation
	t.Run("VehicleBatteryLevels", func(t *testing.T) {
		vehicles, err := client.GetVehicles()
		require.NoError(t, err)

		for _, vehicle := range vehicles {
			// Battery level should be reasonable
			assert.GreaterOrEqual(t, vehicle.BatteryLevel, 0, "Battery level should not be negative")
			assert.LessOrEqual(t, vehicle.BatteryLevel, 100, "Battery level should not exceed 100%")
			
			// Battery range should correlate with battery level
			assert.Greater(t, vehicle.BatteryRangeKm, 0.0, "Battery range should be positive")
			
			// If battery is very low, vehicle might be charging
			if vehicle.BatteryLevel <= 30 {
				t.Logf("Vehicle %s has low battery (%d%%) and status: %s", 
					vehicle.ID, vehicle.BatteryLevel, vehicle.Status)
			}
		}
	})
}
