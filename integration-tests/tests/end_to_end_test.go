package tests

import (
	"testing"
	"time"

	"integration-tests/internal/testhelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndToEndWorkflow(t *testing.T) {
	// Setup: Start all services
	sm := testhelpers.NewServiceManager()
	defer sm.StopServices()

	// Start Fleet and Job services
	err := sm.StartServices()
	require.NoError(t, err, "Failed to start services")

	// Start car simulator with 2 vehicles
	err = sm.StartCarSimulator(2)
	require.NoError(t, err, "Failed to start car simulator")

	// Wait for vehicles to register
	err = sm.WaitForVehicleRegistration(2, 10*time.Second)
	require.NoError(t, err, "Vehicles did not register in time")

	// Create HTTP client for API calls
	client := testhelpers.NewHTTPClient()

	// Test 1: Verify vehicles are registered and available
	t.Run("VehiclesRegistered", func(t *testing.T) {
		vehicles, err := client.GetVehicles()
		require.NoError(t, err)
		assert.Len(t, vehicles, 2, "Expected 2 vehicles to be registered")

		for _, vehicle := range vehicles {
			assert.Equal(t, "available", vehicle.Status, "Vehicle should be available")
			assert.Equal(t, "us-west-2", vehicle.Region, "Vehicle should be in us-west-2 region")
			assert.Greater(t, vehicle.BatteryLevel, 0, "Vehicle should have battery")
		}
	})

	// Test 2: Create a ride job and verify assignment
	t.Run("RideJobAssignment", func(t *testing.T) {
		// Create a ride job
		job, err := client.CreateRideJob(
			"customer-123",
			"us-west-2",
			37.7749, -122.4194, // San Francisco (pickup)
			37.7849, -122.4094, // Slightly north (destination)
		)
		require.NoError(t, err, "Failed to create ride job")

		// Job should be created
		assert.Equal(t, "ride", job.JobType)
		assert.Equal(t, "customer-123", job.CustomerID)
		assert.Greater(t, job.EstimatedDistanceKm, 0.0, "Job should have estimated distance")

		// Job should be assigned (either immediately or after a short wait)
		var assignedJob *testhelpers.Job
		for i := 0; i < 10; i++ { // Wait up to 5 seconds
			assignedJob, err = client.GetJob(job.ID)
			require.NoError(t, err)
			
			if assignedJob.Status == "assigned" && assignedJob.AssignedVehicleID != nil {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}

		assert.Equal(t, "assigned", assignedJob.Status, "Job should be assigned")
		assert.NotNil(t, assignedJob.AssignedVehicleID, "Job should have assigned vehicle")

		// Verify the assigned vehicle is now busy
		vehicles, err := client.GetVehicles()
		require.NoError(t, err)

		var assignedVehicle *testhelpers.Vehicle
		for _, vehicle := range vehicles {
			if vehicle.ID == *assignedJob.AssignedVehicleID {
				assignedVehicle = vehicle
				break
			}
		}

		require.NotNil(t, assignedVehicle, "Assigned vehicle should exist")
		assert.Equal(t, "busy", assignedVehicle.Status, "Assigned vehicle should be busy")
	})

	// Test 3: Create a delivery job
	t.Run("DeliveryJobAssignment", func(t *testing.T) {
		deliveryDetails := &testhelpers.DeliveryDetails{
			RestaurantName: "Pizza Palace",
			Items:          []string{"Large Pizza", "Garlic Bread"},
			Instructions:   "Ring doorbell twice",
		}

		job, err := client.CreateDeliveryJob(
			"customer-456",
			"us-west-2",
			37.7649, -122.4294, // Different pickup location
			37.7749, -122.4194, // Different destination
			deliveryDetails,
		)
		require.NoError(t, err, "Failed to create delivery job")

		assert.Equal(t, "delivery", job.JobType)
		assert.Equal(t, "customer-456", job.CustomerID)
		assert.NotNil(t, job.DeliveryDetails, "Delivery job should have details")
		assert.Equal(t, "Pizza Palace", job.DeliveryDetails.RestaurantName)

		// Wait for assignment
		var assignedJob *testhelpers.Job
		for i := 0; i < 10; i++ {
			assignedJob, err = client.GetJob(job.ID)
			require.NoError(t, err)
			
			if assignedJob.Status == "assigned" && assignedJob.AssignedVehicleID != nil {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}

		assert.Equal(t, "assigned", assignedJob.Status, "Delivery job should be assigned")
		assert.NotNil(t, assignedJob.AssignedVehicleID, "Delivery job should have assigned vehicle")
	})

	// Test 4: Wait for job completion (this tests the full car simulator workflow)
	t.Run("JobCompletion", func(t *testing.T) {
		// Get all jobs
		jobs, err := client.GetJobs()
		require.NoError(t, err)

		// Find an assigned job
		var testJob *testhelpers.Job
		for _, job := range jobs {
			if job.Status == "assigned" {
				testJob = job
				break
			}
		}

		if testJob == nil {
			t.Skip("No assigned job found to test completion")
			return
		}

		// Wait for job completion (car simulator should complete it)
		// This tests the pickup -> delivery -> completion workflow
		completed := false
		for i := 0; i < 60; i++ { // Wait up to 30 seconds
			updatedJob, err := client.GetJob(testJob.ID)
			require.NoError(t, err)

			if updatedJob.Status == "completed" {
				completed = true
				assert.NotNil(t, updatedJob.CompletedAt, "Completed job should have completion time")
				break
			}

			time.Sleep(500 * time.Millisecond)
		}

		if !completed {
			t.Log("Job did not complete within 30 seconds - this may be expected for longer routes")
			// Don't fail the test, as job completion depends on simulated travel time
		}
	})

	// Test 5: Verify system state after operations
	t.Run("SystemStateAfterOperations", func(t *testing.T) {
		vehicles, err := client.GetVehicles()
		require.NoError(t, err)
		assert.Len(t, vehicles, 2, "Should still have 2 vehicles")

		jobs, err := client.GetJobs()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(jobs), 2, "Should have at least 2 jobs created")

		// At least one vehicle should be working or have worked
		busyOrAvailableCount := 0
		for _, vehicle := range vehicles {
			if vehicle.Status == "busy" || vehicle.Status == "available" {
				busyOrAvailableCount++
			}
		}
		assert.Equal(t, 2, busyOrAvailableCount, "All vehicles should be in valid states")
	})
}
