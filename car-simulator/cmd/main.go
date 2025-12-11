package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"car-simulator/internal/simulator"
)

func main() {
	// Setup structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Get configuration from environment variables
	fleetServiceURL := getEnv("FLEET_SERVICE_URL", "http://localhost:8080")
	jobServiceURL := getEnv("JOB_SERVICE_URL", "http://localhost:8081")
	region := getEnv("REGION", "us-west-2")
	vehicleCount := getEnvInt("VEHICLE_COUNT", 1)

	// Get starting area coordinates (default to San Francisco)
	startLat := getEnvFloat("START_LAT", 37.7749)
	startLng := getEnvFloat("START_LNG", -122.4194)

	slog.Info("Starting vehicle simulators",
		"vehicle_count", vehicleCount,
		"region", region,
		"fleet_service_url", fleetServiceURL,
		"job_service_url", jobServiceURL,
		"start_lat", startLat,
		"start_lng", startLng)

	// Wait for fleet service to be ready after system reset
	slog.Info("Waiting for fleet service to initialize", "wait_seconds", 45)
	time.Sleep(45 * time.Second)

	// Create and start vehicles
	var vehicles []*simulator.Vehicle
	for i := 0; i < vehicleCount; i++ {
		vehicleID := fmt.Sprintf("sim-vehicle-%d", i+1)

		// Use predefined spawn location instead of random coordinates
		spawnLocation := simulator.GetRandomSpawnLocation()
		lat := spawnLocation.Lat
		lng := spawnLocation.Lng

		vehicle := simulator.NewVehicle(vehicleID, region, fleetServiceURL, jobServiceURL, lat, lng)

		if err := vehicle.Start(); err != nil {
			slog.Error("Failed to start vehicle", "vehicle_id", vehicleID, "error", err)
			continue
		}

		vehicles = append(vehicles, vehicle)
		slog.Info("Started vehicle",
			"vehicle_id", vehicleID,
			"spawn_location", spawnLocation.Name,
			"lat", lat,
			"lng", lng)

		// Stagger vehicle starts to avoid overwhelming the fleet service
		time.Sleep(100 * time.Millisecond)
	}

	slog.Info("Vehicle startup complete", "started_count", len(vehicles), "requested_count", vehicleCount)

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	slog.Info("Car simulators running, waiting for shutdown signal")
	<-c

	slog.Info("Shutting down car simulators")
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets environment variable as integer with default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvFloat gets environment variable as float64 with default value
func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
