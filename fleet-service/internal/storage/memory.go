package storage

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MemoryVehicleStorage implements VehicleStorage using in-memory maps
type MemoryVehicleStorage struct {
	vehicles map[string]*Vehicle
	mu       sync.RWMutex
}

// NewMemoryVehicleStorage creates a new in-memory storage instance
func NewMemoryVehicleStorage() *MemoryVehicleStorage {
	return &MemoryVehicleStorage{
		vehicles: make(map[string]*Vehicle),
	}
}

func (m *MemoryVehicleStorage) CreateVehicle(ctx context.Context, vehicle *Vehicle) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.vehicles[vehicle.ID]; exists {
		return fmt.Errorf("vehicle %s already exists", vehicle.ID)
	}

	vehicle.LastUpdated = time.Now()
	m.vehicles[vehicle.ID] = vehicle
	return nil
}

func (m *MemoryVehicleStorage) GetVehicle(ctx context.Context, vehicleID string) (*Vehicle, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	vehicle, exists := m.vehicles[vehicleID]
	if !exists {
		return nil, fmt.Errorf("vehicle %s not found", vehicleID)
	}

	return vehicle, nil
}

func (m *MemoryVehicleStorage) UpdateVehicle(ctx context.Context, vehicle *Vehicle) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.vehicles[vehicle.ID]; !exists {
		return fmt.Errorf("vehicle %s not found", vehicle.ID)
	}

	vehicle.LastUpdated = time.Now()
	m.vehicles[vehicle.ID] = vehicle
	return nil
}

func (m *MemoryVehicleStorage) GetVehiclesByRegionAndStatus(ctx context.Context, region, status string) ([]*Vehicle, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Vehicle
	for _, vehicle := range m.vehicles {
		if vehicle.Region == region && vehicle.Status == status {
			result = append(result, vehicle)
		}
	}

	return result, nil
}

func (m *MemoryVehicleStorage) GetAllVehicles(ctx context.Context) ([]*Vehicle, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Vehicle
	for _, vehicle := range m.vehicles {
		result = append(result, vehicle)
	}

	return result, nil
}

func (m *MemoryVehicleStorage) UpdateVehicleLocationAndStatus(ctx context.Context, vehicleID string, lat, lng float64, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	vehicle, exists := m.vehicles[vehicleID]
	if !exists {
		return fmt.Errorf("vehicle %s not found", vehicleID)
	}

	vehicle.LocationLat = lat
	vehicle.LocationLng = lng
	vehicle.Status = status
	vehicle.LastUpdated = time.Now()
	return nil
}

func (m *MemoryVehicleStorage) UpdateVehicleLocation(ctx context.Context, vehicleID string, lat, lng float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	vehicle, exists := m.vehicles[vehicleID]
	if !exists {
		return fmt.Errorf("vehicle %s not found", vehicleID)
	}

	vehicle.LocationLat = lat
	vehicle.LocationLng = lng
	vehicle.LastUpdated = time.Now()

	return nil
}

func (m *MemoryVehicleStorage) UpdateVehicleStatus(ctx context.Context, vehicleID string, status string, jobID *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	vehicle, exists := m.vehicles[vehicleID]
	if !exists {
		return fmt.Errorf("vehicle %s not found", vehicleID)
	}

	vehicle.Status = status
	vehicle.CurrentJobID = jobID
	vehicle.LastUpdated = time.Now()

	return nil
}
