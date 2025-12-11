package service

import (
	"context"
	"testing"
	"time"

	"job-service/internal/storage"
)

func TestJobService_GetRevenue(t *testing.T) {
	jobStorage := storage.NewMemoryJobStorage()
	mockFleetClient := NewMockFleetClient()
	jobService := NewJobService(jobStorage, mockFleetClient)
	ctx := context.Background()

	// Create some completed jobs with different fares
	completedRide := &storage.Job{
		ID:          "ride-1",
		JobType:     "ride",
		Status:      "completed",
		FareAmount:  15.50,
		CreatedAt:   time.Now(),
		CompletedAt: &time.Time{},
	}

	completedDelivery := &storage.Job{
		ID:          "delivery-1",
		JobType:     "delivery",
		Status:      "completed",
		FareAmount:  8.99,
		CreatedAt:   time.Now(),
		CompletedAt: &time.Time{},
	}

	pendingJob := &storage.Job{
		ID:         "ride-2",
		JobType:    "ride",
		Status:     "pending",
		FareAmount: 12.00,
		CreatedAt:  time.Now(),
	}

	// Add jobs to storage
	jobStorage.CreateJob(ctx, completedRide)
	jobStorage.CreateJob(ctx, completedDelivery)
	jobStorage.CreateJob(ctx, pendingJob)

	revenue, err := jobService.GetRevenue(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check total revenue (only completed jobs)
	expectedTotal := 15.50 + 8.99 // $24.49
	actualTotal := revenue["total_revenue"].(float64)
	if actualTotal < expectedTotal-0.01 || actualTotal > expectedTotal+0.01 {
		t.Errorf("Expected total revenue %.2f, got %.2f", expectedTotal, actualTotal)
	}

	// Check ride revenue
	expectedRideRevenue := 15.50
	if revenue["ride_revenue"].(float64) != expectedRideRevenue {
		t.Errorf("Expected ride revenue %.2f, got %.2f", expectedRideRevenue, revenue["ride_revenue"].(float64))
	}

	// Check delivery revenue
	expectedDeliveryRevenue := 8.99
	if revenue["delivery_revenue"].(float64) != expectedDeliveryRevenue {
		t.Errorf("Expected delivery revenue %.2f, got %.2f", expectedDeliveryRevenue, revenue["delivery_revenue"].(float64))
	}

	// Check completed jobs count
	expectedCompleted := 2
	if revenue["completed_jobs"].(int) != expectedCompleted {
		t.Errorf("Expected %d completed jobs, got %d", expectedCompleted, revenue["completed_jobs"].(int))
	}

	// Check ride count
	expectedRideCount := 1
	if revenue["ride_count"].(int) != expectedRideCount {
		t.Errorf("Expected %d rides, got %d", expectedRideCount, revenue["ride_count"].(int))
	}

	// Check delivery count
	expectedDeliveryCount := 1
	if revenue["delivery_count"].(int) != expectedDeliveryCount {
		t.Errorf("Expected %d deliveries, got %d", expectedDeliveryCount, revenue["delivery_count"].(int))
	}
}

func TestJobService_GetRevenue_NoCompletedJobs(t *testing.T) {
	jobStorage := storage.NewMemoryJobStorage()
	mockFleetClient := NewMockFleetClient()
	jobService := NewJobService(jobStorage, mockFleetClient)
	ctx := context.Background()

	revenue, err := jobService.GetRevenue(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// All values should be zero
	if revenue["total_revenue"].(float64) != 0.0 {
		t.Errorf("Expected total revenue 0.0, got %.2f", revenue["total_revenue"].(float64))
	}

	if revenue["completed_jobs"].(int) != 0 {
		t.Errorf("Expected 0 completed jobs, got %d", revenue["completed_jobs"].(int))
	}
}
