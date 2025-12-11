package service

import (
	"context"
	"testing"
	"time"

	"job-service/internal/storage"
)

func TestDemoJobGenerator_ActiveJobLimit(t *testing.T) {
	// Setup
	memStorage := storage.NewMemoryJobStorage()
	mockFleet := NewMockFleetClient()
	jobService := NewJobService(memStorage, mockFleet)

	// Create demo generator with low limit for testing
	generator := &DemoJobGenerator{
		jobService: jobService,
		interval:   100 * time.Millisecond,
		stopChan:   make(chan bool),
		maxJobs:    3, // Low limit for testing
	}

	// Test that generator respects active job limit
	ctx := context.Background()

	// Create jobs up to the limit (will be pending since no vehicles)
	for i := 0; i < 3; i++ {
		generator.createRandomJob()
	}

	// Verify we have 3 active jobs (all pending)
	activeCount, err := jobService.GetActiveJobCount()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if activeCount != 3 {
		t.Errorf("Expected 3 active jobs, got %d", activeCount)
	}

	// Complete one job to free up space
	jobs, _ := jobService.GetAllJobs(ctx)
	if len(jobs) > 0 {
		// Force complete the job even if it's pending
		memStorage.UpdateJobStatus(ctx, jobs[0].ID, "completed", nil)
	}

	// Verify we now have 2 active jobs
	activeCount, err = jobService.GetActiveJobCount()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if activeCount != 2 {
		t.Errorf("Expected 2 active jobs after completion, got %d", activeCount)
	}

	// Should be able to create one more job
	generator.createRandomJob()

	// Verify we're back to 3 active jobs
	activeCount, err = jobService.GetActiveJobCount()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if activeCount != 3 {
		t.Errorf("Expected 3 active jobs after new creation, got %d", activeCount)
	}
}

func TestDemoJobGenerator_NewDemoJobGenerator(t *testing.T) {
	memStorage := storage.NewMemoryJobStorage()
	mockFleet := NewMockFleetClient()
	jobService := NewJobService(memStorage, mockFleet)

	generator := NewDemoJobGenerator(jobService, 5*time.Second)

	if generator.jobService != jobService {
		t.Error("Expected job service to be set")
	}

	if generator.interval != 5*time.Second {
		t.Errorf("Expected interval 5s, got %v", generator.interval)
	}

	if generator.maxJobs != 25 {
		t.Errorf("Expected max jobs 25, got %d", generator.maxJobs)
	}

	if generator.stopChan == nil {
		t.Error("Expected stop channel to be initialized")
	}
}
