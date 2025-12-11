package testhelpers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// ServiceManager manages the lifecycle of test services
type ServiceManager struct {
	fleetCmd *exec.Cmd
	jobCmd   *exec.Cmd
	carCmd   *exec.Cmd
}

// NewServiceManager creates a new service manager
func NewServiceManager() *ServiceManager {
	return &ServiceManager{}
}

// StartServices starts all required services for integration testing
func (sm *ServiceManager) StartServices() error {
	// Start Fleet Service on port 8080
	sm.fleetCmd = exec.Command("../../fleet-service/bin/fleet-service")
	sm.fleetCmd.Env = append(os.Environ(), "PORT=8080")
	if err := sm.fleetCmd.Start(); err != nil {
		return fmt.Errorf("failed to start fleet service: %v", err)
	}

	// Wait for Fleet Service to be ready
	if err := sm.waitForService("http://localhost:8080/vehicles", 10*time.Second); err != nil {
		sm.StopServices()
		return fmt.Errorf("fleet service not ready: %v", err)
	}

	// Start Job Service on port 8081
	sm.jobCmd = exec.Command("../../job-service/bin/job-service")
	sm.jobCmd.Env = append(os.Environ(), 
		"PORT=8081",
		"FLEET_SERVICE_URL=http://localhost:8080",
	)
	if err := sm.jobCmd.Start(); err != nil {
		sm.StopServices()
		return fmt.Errorf("failed to start job service: %v", err)
	}

	// Wait for Job Service to be ready
	if err := sm.waitForService("http://localhost:8081/jobs", 10*time.Second); err != nil {
		sm.StopServices()
		return fmt.Errorf("job service not ready: %v", err)
	}

	fmt.Println("✅ Fleet Service and Job Service started successfully")
	return nil
}

// StartCarSimulator starts a car simulator with specified parameters
func (sm *ServiceManager) StartCarSimulator(vehicleCount int) error {
	sm.carCmd = exec.Command("../../car-simulator/bin/car-simulator")
	sm.carCmd.Env = append(os.Environ(),
		"FLEET_SERVICE_URL=http://localhost:8080",
		"JOB_SERVICE_URL=http://localhost:8081",
		"REGION=us-west-2",
		fmt.Sprintf("VEHICLE_COUNT=%d", vehicleCount),
		"START_LAT=37.7749",
		"START_LNG=-122.4194",
	)
	
	if err := sm.carCmd.Start(); err != nil {
		return fmt.Errorf("failed to start car simulator: %v", err)
	}

	// Give car simulator time to register vehicles
	time.Sleep(2 * time.Second)
	
	fmt.Printf("✅ Car Simulator started with %d vehicles\n", vehicleCount)
	return nil
}

// StopServices stops all running services
func (sm *ServiceManager) StopServices() {
	if sm.carCmd != nil && sm.carCmd.Process != nil {
		sm.carCmd.Process.Signal(syscall.SIGTERM)
		sm.carCmd.Wait()
	}
	
	if sm.jobCmd != nil && sm.jobCmd.Process != nil {
		sm.jobCmd.Process.Signal(syscall.SIGTERM)
		sm.jobCmd.Wait()
	}
	
	if sm.fleetCmd != nil && sm.fleetCmd.Process != nil {
		sm.fleetCmd.Process.Signal(syscall.SIGTERM)
		sm.fleetCmd.Wait()
	}
	
	fmt.Println("🛑 All services stopped")
}

// waitForService waits for a service to become available
func (sm *ServiceManager) waitForService(url string, timeout time.Duration) error {
	client := &http.Client{Timeout: 1 * time.Second}
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	
	return fmt.Errorf("service at %s not ready within %v", url, timeout)
}

// WaitForVehicleRegistration waits for vehicles to register with fleet service
func (sm *ServiceManager) WaitForVehicleRegistration(expectedCount int, timeout time.Duration) error {
	client := &http.Client{Timeout: 5 * time.Second}
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		resp, err := client.Get("http://localhost:8080/vehicles")
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		
		if resp.StatusCode == http.StatusOK {
			// Parse response to count vehicles
			var vehicles []interface{}
			if err := parseJSONResponse(resp, &vehicles); err == nil {
				if len(vehicles) >= expectedCount {
					resp.Body.Close()
					return nil
				}
			}
		}
		resp.Body.Close()
		time.Sleep(500 * time.Millisecond)
	}
	
	return fmt.Errorf("expected %d vehicles not registered within %v", expectedCount, timeout)
}
