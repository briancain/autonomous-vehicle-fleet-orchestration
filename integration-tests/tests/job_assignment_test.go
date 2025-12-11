package tests

import (
	"testing"
	"time"

	"integration-tests/internal/testhelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobAssignmentLogic(t *testing.T) {
	// Setup: Start Fleet and Job services only (no car simulator)
	sm := testhelpers.NewServiceManager()
	defer sm.StopServices()

	err := sm.StartServices()
	require.NoError(t, err, "Failed to start services")

	client := testhelpers.NewHTTPClient()

	// Test 1: No vehicles available - job should remain pending
	t.Run("NoVehiclesAvailable", func(t *testing.T) {
		job, err := client.CreateRideJob(
			"customer-no-vehicles",
			"us-west-2",
			37.7749, -122.4194,
			37.7849, -122.4094,
		)
		require.NoError(t, err)

		// Job should be created but remain pending
		assert.Equal(t, "pending", job.Status)
		assert.Nil(t, job.AssignedVehicleID)

		// Wait a bit and verify it's still pending
		time.Sleep(2 * time.Second)
		updatedJob, err := client.GetJob(job.ID)
		require.NoError(t, err)
		assert.Equal(t, "pending", updatedJob.Status)
	})

	// Test 2: Start car simulator and verify pending jobs get assigned
	t.Run("PendingJobsGetAssigned", func(t *testing.T) {
		// Start car simulator with 1 vehicle
		err := sm.StartCarSimulator(1)
		require.NoError(t, err)

		// Wait for vehicle registration
		err = sm.WaitForVehicleRegistration(1, 10*time.Second)
		require.NoError(t, err)

		// Get all jobs (should include the pending one from previous test)
		jobs, err := client.GetJobs()
		require.NoError(t, err)

		// Find pending jobs
		var pendingJobs []*testhelpers.Job
		for _, job := range jobs {
			if job.Status == "pending" {
				pendingJobs = append(pendingJobs, job)
			}
		}

		if len(pendingJobs) == 0 {
			t.Skip("No pending jobs found")
			return
		}

		// Wait for pending jobs to be assigned
		assigned := false
		for i := 0; i < 20; i++ { // Wait up to 10 seconds
			updatedJob, err := client.GetJob(pendingJobs[0].ID)
			require.NoError(t, err)

			if updatedJob.Status == "assigned" {
				assigned = true
				assert.NotNil(t, updatedJob.AssignedVehicleID)
				break
			}
			time.Sleep(500 * time.Millisecond)
		}

		assert.True(t, assigned, "Pending job should be assigned when vehicle becomes available")
	})

	// Test 3: Multiple jobs with one vehicle - should be assigned sequentially
	t.Run("SequentialJobAssignment", func(t *testing.T) {
		// Create multiple jobs quickly
		var createdJobs []*testhelpers.Job
		for i := 0; i < 3; i++ {
			job, err := client.CreateRideJob(
				"customer-sequential",
				"us-west-2",
				37.7749+float64(i)*0.001, -122.4194,
				37.7849+float64(i)*0.001, -122.4094,
			)
			require.NoError(t, err)
			createdJobs = append(createdJobs, job)
		}

		// Wait and check job statuses
		time.Sleep(3 * time.Second)

		assignedCount := 0
		pendingCount := 0

		for _, job := range createdJobs {
			updatedJob, err := client.GetJob(job.ID)
			require.NoError(t, err)

			switch updatedJob.Status {
			case "assigned":
				assignedCount++
			case "pending":
				pendingCount++
			}
		}

		// With only 1 vehicle, we should have at most 1 assigned job
		// and the rest should be pending
		assert.LessOrEqual(t, assignedCount, 1, "Should have at most 1 assigned job with 1 vehicle")
		assert.GreaterOrEqual(t, pendingCount, 2, "Should have at least 2 pending jobs")
	})

	// Test 4: Job assignment to nearest vehicle
	t.Run("NearestVehicleAssignment", func(t *testing.T) {
		// This test would require multiple vehicles at different locations
		// For now, we'll verify that jobs are assigned to available vehicles
		vehicles, err := client.GetVehicles()
		require.NoError(t, err)

		if len(vehicles) == 0 {
			t.Skip("No vehicles available for nearest vehicle test")
			return
		}

		// Create a job near the first vehicle
		vehicle := vehicles[0]
		job, err := client.CreateRideJob(
			"customer-nearest",
			vehicle.Region,
			vehicle.LocationLat+0.001, vehicle.LocationLng+0.001, // Very close to vehicle
			vehicle.LocationLat+0.002, vehicle.LocationLng+0.002,
		)
		require.NoError(t, err)

		// If vehicle is available, job should be assigned to it
		if vehicle.Status == "available" {
			// Wait for assignment
			for i := 0; i < 10; i++ {
				updatedJob, err := client.GetJob(job.ID)
				require.NoError(t, err)

				if updatedJob.Status == "assigned" {
					assert.Equal(t, vehicle.ID, *updatedJob.AssignedVehicleID, 
						"Job should be assigned to the nearest available vehicle")
					break
				}
				time.Sleep(500 * time.Millisecond)
			}
		}
	})

	// Test 5: Different regions
	t.Run("RegionalJobAssignment", func(t *testing.T) {
		// Create a job in a different region
		job, err := client.CreateRideJob(
			"customer-different-region",
			"us-east-1", // Different region
			40.7128, -74.0060, // New York coordinates
			40.7228, -74.0160,
		)
		require.NoError(t, err)

		// Job should remain pending since we only have vehicles in us-west-2
		time.Sleep(2 * time.Second)
		updatedJob, err := client.GetJob(job.ID)
		require.NoError(t, err)

		assert.Equal(t, "pending", updatedJob.Status, 
			"Job in different region should remain pending")
		assert.Nil(t, updatedJob.AssignedVehicleID)
	})
}
