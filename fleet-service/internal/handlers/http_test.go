package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fleet-service/internal/service"
	"fleet-service/internal/storage"

	"github.com/gorilla/mux"
)

func setupTestHandler() (*HTTPHandler, *storage.MemoryVehicleStorage) {
	vehicleStorage := storage.NewMemoryVehicleStorage()
	fleetService := service.NewFleetService(vehicleStorage)
	handler := NewHTTPHandler(fleetService)
	return handler, vehicleStorage
}

func TestHTTPHandler_RegisterVehicle(t *testing.T) {
	handler, _ := setupTestHandler()

	vehicle := storage.Vehicle{
		ID:             "test-vehicle-1",
		Region:         "us-west-2",
		Status:         "available",
		BatteryLevel:   80,
		BatteryRangeKm: 200.0,
		LocationLat:    37.7749,
		LocationLng:    -122.4194,
		VehicleType:    "sedan",
	}

	jsonData, _ := json.Marshal(vehicle)
	req := httptest.NewRequest("POST", "/vehicles", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.RegisterVehicle(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	var response storage.Vehicle
	json.NewDecoder(rr.Body).Decode(&response)

	if response.ID != vehicle.ID {
		t.Errorf("Expected ID %s, got %s", vehicle.ID, response.ID)
	}
}

func TestHTTPHandler_GetAllVehicles(t *testing.T) {
	handler, vehicleStorage := setupTestHandler()

	// Add test vehicle
	vehicle := &storage.Vehicle{
		ID:             "test-vehicle-1",
		Region:         "us-west-2",
		Status:         "available",
		BatteryLevel:   80,
		BatteryRangeKm: 200.0,
		LocationLat:    37.7749,
		LocationLng:    -122.4194,
		VehicleType:    "sedan",
	}
	vehicleStorage.CreateVehicle(nil, vehicle)

	req := httptest.NewRequest("GET", "/vehicles", nil)
	rr := httptest.NewRecorder()

	handler.GetAllVehicles(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var vehicles []*storage.Vehicle
	json.NewDecoder(rr.Body).Decode(&vehicles)

	if len(vehicles) != 1 {
		t.Errorf("Expected 1 vehicle, got %d", len(vehicles))
	}

	if vehicles[0].ID != "test-vehicle-1" {
		t.Errorf("Expected vehicle ID test-vehicle-1, got %s", vehicles[0].ID)
	}
}

func TestHTTPHandler_UpdateVehicleLocation(t *testing.T) {
	handler, vehicleStorage := setupTestHandler()

	// Add test vehicle
	vehicle := &storage.Vehicle{
		ID:             "test-vehicle-1",
		Region:         "us-west-2",
		Status:         "available",
		BatteryLevel:   80,
		BatteryRangeKm: 200.0,
		LocationLat:    37.7749,
		LocationLng:    -122.4194,
		VehicleType:    "sedan",
	}
	vehicleStorage.CreateVehicle(nil, vehicle)

	locationUpdate := struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}{
		Lat: 37.7849,
		Lng: -122.4094,
	}

	jsonData, _ := json.Marshal(locationUpdate)
	req := httptest.NewRequest("PUT", "/vehicles/test-vehicle-1/location", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Setup router to handle path variables
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify location was updated
	updated, _ := vehicleStorage.GetVehicle(nil, "test-vehicle-1")
	if updated.LocationLat != locationUpdate.Lat || updated.LocationLng != locationUpdate.Lng {
		t.Errorf("Expected location (%f, %f), got (%f, %f)",
			locationUpdate.Lat, locationUpdate.Lng, updated.LocationLat, updated.LocationLng)
	}
}

func TestHTTPHandler_AssignJob(t *testing.T) {
	handler, vehicleStorage := setupTestHandler()

	// Add test vehicle
	vehicle := &storage.Vehicle{
		ID:             "test-vehicle-1",
		Region:         "us-west-2",
		Status:         "available",
		BatteryLevel:   80,
		BatteryRangeKm: 200.0,
		LocationLat:    37.7749,
		LocationLng:    -122.4194,
		VehicleType:    "sedan",
	}
	vehicleStorage.CreateVehicle(nil, vehicle)

	jobAssignment := struct {
		JobID string `json:"job_id"`
	}{
		JobID: "job-123",
	}

	jsonData, _ := json.Marshal(jobAssignment)
	req := httptest.NewRequest("POST", "/vehicles/test-vehicle-1/assign", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify job was assigned
	updated, _ := vehicleStorage.GetVehicle(nil, "test-vehicle-1")
	if updated.Status != "busy" {
		t.Errorf("Expected status 'busy', got '%s'", updated.Status)
	}
	if updated.CurrentJobID == nil || *updated.CurrentJobID != "job-123" {
		t.Errorf("Expected job ID 'job-123', got %v", updated.CurrentJobID)
	}
}

func TestHTTPHandler_FindNearestVehicle(t *testing.T) {
	handler, vehicleStorage := setupTestHandler()

	// Add test vehicles
	vehicles := []*storage.Vehicle{
		{
			ID:             "v1",
			Region:         "us-west-2",
			Status:         "available",
			BatteryLevel:   80,
			BatteryRangeKm: 200.0,
			LocationLat:    37.7749,
			LocationLng:    -122.4194,
			VehicleType:    "sedan",
		},
		{
			ID:             "v2",
			Region:         "us-west-2",
			Status:         "available",
			BatteryLevel:   90,
			BatteryRangeKm: 250.0,
			LocationLat:    37.8049, // Farther away
			LocationLng:    -122.4394,
			VehicleType:    "sedan",
		},
	}

	for _, v := range vehicles {
		vehicleStorage.CreateVehicle(nil, v)
	}

	req := httptest.NewRequest("GET", "/vehicles/find?region=us-west-2&pickup_lat=37.7649&pickup_lng=-122.4294&trip_distance_km=50", nil)
	rr := httptest.NewRecorder()

	handler.FindNearestVehicle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var vehicle storage.Vehicle
	json.NewDecoder(rr.Body).Decode(&vehicle)

	// Should return v1 (closer to pickup location)
	if vehicle.ID != "v1" {
		t.Errorf("Expected vehicle v1, got %s", vehicle.ID)
	}
}

func TestHTTPHandler_FindNearestVehicle_MissingParams(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest("GET", "/vehicles/find?region=us-west-2", nil) // Missing required params
	rr := httptest.NewRecorder()

	handler.FindNearestVehicle(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}
