package storage

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MemoryJobStorage implements JobStorage using in-memory maps
type MemoryJobStorage struct {
	jobs map[string]*Job
	mu   sync.RWMutex
}

// NewMemoryJobStorage creates a new in-memory storage instance
func NewMemoryJobStorage() *MemoryJobStorage {
	return &MemoryJobStorage{
		jobs: make(map[string]*Job),
	}
}

func (m *MemoryJobStorage) CreateJob(ctx context.Context, job *Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.jobs[job.ID]; exists {
		return fmt.Errorf("job %s already exists", job.ID)
	}

	job.CreatedAt = time.Now()
	m.jobs[job.ID] = job
	return nil
}

func (m *MemoryJobStorage) GetJob(ctx context.Context, jobID string) (*Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job %s not found", jobID)
	}

	return job, nil
}

func (m *MemoryJobStorage) UpdateJob(ctx context.Context, job *Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.jobs[job.ID]; !exists {
		return fmt.Errorf("job %s not found", job.ID)
	}

	m.jobs[job.ID] = job
	return nil
}

func (m *MemoryJobStorage) GetJobsByStatus(ctx context.Context, status string) ([]*Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Job
	for _, job := range m.jobs {
		if job.Status == status {
			result = append(result, job)
		}
	}

	return result, nil
}

func (m *MemoryJobStorage) GetJobsByVehicle(ctx context.Context, vehicleID string) ([]*Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Job
	for _, job := range m.jobs {
		if job.AssignedVehicleID != nil && *job.AssignedVehicleID == vehicleID {
			result = append(result, job)
		}
	}

	return result, nil
}

func (m *MemoryJobStorage) GetAllJobs(ctx context.Context) ([]*Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Job
	for _, job := range m.jobs {
		result = append(result, job)
	}

	return result, nil
}

func (m *MemoryJobStorage) UpdateJobStatus(ctx context.Context, jobID, status string, vehicleID *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}

	job.Status = status
	job.AssignedVehicleID = vehicleID

	now := time.Now()
	switch status {
	case "assigned":
		job.AssignedAt = &now
	case "completed":
		job.CompletedAt = &now
	}

	return nil
}
