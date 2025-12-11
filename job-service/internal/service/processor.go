package service

import (
	"context"
	"fmt"
	"time"
)

// JobProcessor handles background processing of pending jobs
type JobProcessor struct {
	jobService *JobService
	stopChan   chan struct{}
}

// NewJobProcessor creates a new job processor
func NewJobProcessor(jobService *JobService) *JobProcessor {
	return &JobProcessor{
		jobService: jobService,
		stopChan:   make(chan struct{}),
	}
}

// Start begins the background job processing
func (jp *JobProcessor) Start() {
	go jp.processLoop()
	fmt.Println("Job processor started")
}

// Stop stops the background job processing
func (jp *JobProcessor) Stop() {
	close(jp.stopChan)
	fmt.Println("Job processor stopped")
}

// processLoop runs the background job processing loop
func (jp *JobProcessor) processLoop() {
	ticker := time.NewTicker(5 * time.Second) // Process every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			jp.processPendingJobs()
		case <-jp.stopChan:
			return
		}
	}
}

// processPendingJobs attempts to assign all pending jobs
func (jp *JobProcessor) processPendingJobs() {
	ctx := context.Background()

	if err := jp.jobService.ProcessPendingJobs(ctx); err != nil {
		fmt.Printf("Error processing pending jobs: %v\n", err)
	}
}
