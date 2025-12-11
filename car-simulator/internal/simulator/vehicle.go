package simulator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"car-simulator/internal/job"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
)

// Vehicle represents a simulated autonomous vehicle
type Vehicle struct {
	ID             string  `json:"id"`
	Region         string  `json:"region"`
	Status         string  `json:"status"` // available, busy, charging, maintenance
	BatteryLevel   float64 `json:"battery_level"`
	BatteryRangeKm float64 `json:"battery_range_km"`
	LocationLat    float64 `json:"location_lat"`
	LocationLng    float64 `json:"location_lng"`
	CurrentJobID   *string `json:"current_job_id,omitempty"`
	VehicleType    string  `json:"vehicle_type"`

	// Simulation state
	fleetServiceURL  string
	jobServiceURL    string
	jobClient        job.JobClient
	targetLat        float64
	targetLng        float64
	isMoving         bool
	batteryDrainRate float64 // km per battery percent
	currentJob       *job.Job
	jobPhase         string // "pickup", "delivery", "idle"

	// Routing state
	routingService *RoutingService
	currentRoute   *Route
	routeIndex     int // current position in route

	// Kinesis streaming (optional)
	kinesisClient *kinesis.Client
	streamName    string
}

// NewVehicle creates a new simulated vehicle
func NewVehicle(id, region, fleetServiceURL, jobServiceURL string, startLat, startLng float64) *Vehicle {
	batteryLevel := rand.Intn(40) + 60 // Start with 60-100% battery
	batteryDrainRate := 4.0            // 4.0km per 1% battery (400km total range)

	v := &Vehicle{
		ID:               id,
		Region:           region,
		Status:           "available",
		BatteryLevel:     float64(batteryLevel),
		BatteryRangeKm:   float64(batteryLevel) * batteryDrainRate, // Calculate range from battery level
		LocationLat:      startLat,
		LocationLng:      startLng,
		VehicleType:      "sedan",
		fleetServiceURL:  fleetServiceURL,
		jobServiceURL:    jobServiceURL,
		jobClient:        job.NewClient(jobServiceURL),
		batteryDrainRate: batteryDrainRate,
		jobPhase:         "idle",
		routingService:   NewRoutingService(),
		routeIndex:       0,
	}

	// Initialize Kinesis client if stream name is provided
	v.initKinesis()
	return v
}

// Start begins the vehicle simulation loop
func (v *Vehicle) Start() error {
	// Register with fleet service with retry logic
	if err := v.registerWithFleetRetry(); err != nil {
		return fmt.Errorf("failed to register with fleet after retries: %v", err)
	}

	// Start simulation loop
	go v.simulationLoop()
	return nil
}

// simulationLoop runs the main vehicle behavior
func (v *Vehicle) simulationLoop() {
	ticker := time.NewTicker(2 * time.Second) // Update every 2 seconds
	defer ticker.Stop()

	for range ticker.C {
		// Log current vehicle status
		v.logVehicleStatus()

		// Check for new job assignments
		v.checkForJobs()

		switch v.Status {
		case "available":
			v.simulateIdleBehavior()
		case "busy":
			v.simulateJobExecution()
		case "charging":
			v.simulateCharging()
		case "maintenance":
			v.simulateMaintenance()
		}

		// Update location and status with fleet service
		v.reportToFleet()

		// Check if battery is low and not busy
		if v.BatteryLevel <= 30 && v.Status == "available" {
			slog.Warn("Vehicle battery low, initiating charging",
				"vehicle_id", v.ID,
				"battery_level", v.BatteryLevel,
				"threshold", 30)
			v.goToCharge()
		}
	}
}

// checkForJobs polls the job service for assigned jobs
func (v *Vehicle) checkForJobs() {
	// Don't accept jobs if not available or already have a job
	if v.Status != "available" || v.currentJob != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	jobs, err := v.jobClient.GetAssignedJobs(ctx, v.ID)
	if err != nil {
		slog.Error("Failed to check for jobs", "vehicle_id", v.ID, "error", err)
		return
	}

	// Find an assigned job that's not completed
	for _, job := range jobs {
		if job.Status == "assigned" {
			v.startJob(job)
			break
		}
	}
}

// startJob begins executing a job
func (v *Vehicle) startJob(job *job.Job) {
	v.currentJob = job
	v.Status = "busy"
	v.jobPhase = "pickup"
	v.setRouteTarget(job.PickupLat, job.PickupLng)

	slog.Info("Vehicle started job",
		"vehicle_id", v.ID,
		"job_type", job.JobType,
		"job_id", job.ID,
		"pickup_lat", job.PickupLat,
		"pickup_lng", job.PickupLng)
}

// simulateIdleBehavior makes the vehicle move randomly when idle
func (v *Vehicle) simulateIdleBehavior() {
	if !v.isMoving {
		// Occasionally start moving to a random nearby location
		if rand.Float64() < 0.1 { // 10% chance every 2 seconds
			v.setRandomTarget(0.01) // Within ~1km
			v.isMoving = true
		}
	} else {
		v.moveTowardsTarget()
	}
}

// simulateJobExecution moves vehicle through job phases
func (v *Vehicle) simulateJobExecution() {
	if v.currentJob == nil {
		v.Status = "available"
		v.isMoving = false
		return
	}

	// Safety check: if battery is critically low during job, go to charge
	if v.BatteryLevel <= 15 { // Emergency threshold higher than normal 30%
		slog.Warn("Vehicle battery critically low during job, abandoning job to charge",
			"vehicle_id", v.ID,
			"battery_level", v.BatteryLevel,
			"job_id", v.currentJob.ID)

		// Abandon current job
		v.currentJob = nil
		v.CurrentJobID = nil
		v.jobPhase = "idle"

		// Go to charge immediately
		v.goToCharge()
		return
	}

	if v.isMoving {
		v.moveAlongRoute()

		// Check if reached current target
		if v.distanceToTarget() < 0.001 { // ~100m
			switch v.jobPhase {
			case "pickup":
				// Reached pickup location, now go to destination
				v.jobPhase = "delivery"
				v.setRouteTarget(v.currentJob.DestinationLat, v.currentJob.DestinationLng)
				slog.Info("Vehicle reached pickup, going to destination",
					"vehicle_id", v.ID,
					"destination_lat", v.currentJob.DestinationLat,
					"destination_lng", v.currentJob.DestinationLng)
			case "delivery":
				// Reached destination, complete job
				v.completeCurrentJob()
			}
		}
	}
}

// completeCurrentJob finishes the current job
func (v *Vehicle) completeCurrentJob() {
	if v.currentJob == nil {
		return
	}

	slog.Info("Vehicle completed job",
		"vehicle_id", v.ID,
		"job_type", v.currentJob.JobType,
		"job_id", v.currentJob.ID)

	// Notify job service
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := v.jobClient.CompleteJob(ctx, v.currentJob.ID); err != nil {
		fmt.Printf("Failed to complete job %s: %v\n", v.currentJob.ID, err)
	}

	// Reset vehicle state
	v.currentJob = nil
	v.CurrentJobID = nil
	v.Status = "available"
	v.isMoving = false
	v.jobPhase = "idle"
}

// simulateMaintenance handles vehicle in maintenance state
func (v *Vehicle) simulateMaintenance() {
	// Vehicle is stranded, simulate recovery after some time
	if v.jobPhase == "stranded" {
		// After 30 seconds, simulate roadside assistance
		// In real system, this would be handled by fleet management
		slog.Info("Vehicle requesting roadside assistance",
			"vehicle_id", v.ID,
			"location_lat", v.LocationLat,
			"location_lng", v.LocationLng)

		// Simulate emergency charging to 20% and go to charging station
		v.BatteryLevel = 20
		v.BatteryRangeKm = v.BatteryLevel * v.batteryDrainRate
		v.Status = "charging"
		v.jobPhase = "going_to_charge"
		v.goToCharge()
	}
}

// simulateCharging handles charging behavior
func (v *Vehicle) simulateCharging() {
	// If still moving to charging station
	if v.isMoving && v.jobPhase == "going_to_charge" {
		v.moveAlongRoute()

		// Check if arrived at charging station
		if v.distanceToTarget() < 0.001 { // ~100m
			v.isMoving = false
			v.jobPhase = "charging"
			slog.Info("Vehicle arrived at charging station",
				"vehicle_id", v.ID,
				"battery_level", v.BatteryLevel)
		}
		return
	}

	// Actually charging
	if v.jobPhase == "charging" {
		if v.BatteryLevel < 95 {
			oldBattery := v.BatteryLevel
			v.BatteryLevel += 2 // Charge 2% every 2 seconds
			v.BatteryRangeKm = v.BatteryLevel * v.batteryDrainRate
			slog.Info("Vehicle charging progress",
				"vehicle_id", v.ID,
				"battery_level", v.BatteryLevel,
				"previous_level", oldBattery,
				"range_km", v.BatteryRangeKm)
		} else {
			// Fully charged, become available
			v.Status = "available"
			v.isMoving = false
			v.jobPhase = "idle"
			slog.Info("Vehicle fully charged, returning to service",
				"vehicle_id", v.ID,
				"battery_level", v.BatteryLevel,
				"range_km", v.BatteryRangeKm)
		}
	}
}

// setRouteTarget calculates a route to the target and starts following it
func (v *Vehicle) setRouteTarget(targetLat, targetLng float64) {
	v.targetLat = targetLat
	v.targetLng = targetLng

	// Get route from routing service
	route, err := v.routingService.GetRoute(v.LocationLat, v.LocationLng, targetLat, targetLng)
	if err != nil {
		fmt.Printf("Failed to get route for vehicle %s: %v\n", v.ID, err)
		// Fallback to direct movement
		v.currentRoute = nil
		v.isMoving = true
		return
	}

	v.currentRoute = route
	v.routeIndex = 0
	v.isMoving = true

	fmt.Printf("Vehicle %s calculated route with %d waypoints (%.1fkm, %.1f min)\n",
		v.ID, len(route.Points), route.Distance/1000, route.Duration/60)
}

// moveAlongRoute moves the vehicle along the calculated route
func (v *Vehicle) moveAlongRoute() {
	// Don't move if battery is depleted
	if v.BatteryLevel <= 0 {
		v.handleBatteryDepletion()
		return
	}

	if v.currentRoute == nil || len(v.currentRoute.Points) == 0 {
		// Fallback to direct movement
		v.moveTowardsTarget()
		return
	}

	// Check if we've reached the end of the route
	if v.routeIndex >= len(v.currentRoute.Points)-1 {
		// Reached destination
		v.LocationLat = v.targetLat
		v.LocationLng = v.targetLng
		v.isMoving = false
		v.currentRoute = nil
		v.routeIndex = 0
		return
	}

	// Store previous position for distance calculation
	prevLat := v.LocationLat
	prevLng := v.LocationLng

	// Move towards next waypoint in route
	nextPoint := v.currentRoute.Points[v.routeIndex+1]

	// Calculate movement step for city driving
	stepSize := v.getMovementSpeed() // Configurable driving speed

	latDiff := nextPoint.Lat - v.LocationLat
	lngDiff := nextPoint.Lng - v.LocationLng
	distance := math.Sqrt(latDiff*latDiff + lngDiff*lngDiff)

	if distance < stepSize {
		// Reached this waypoint, move to next
		v.routeIndex++
		v.LocationLat = nextPoint.Lat
		v.LocationLng = nextPoint.Lng
	} else {
		// Move towards waypoint
		v.LocationLat += (latDiff / distance) * stepSize
		v.LocationLng += (lngDiff / distance) * stepSize
	}

	// Drain battery based on actual distance moved using haversine
	kmTraveled := haversineDistance(prevLat, prevLng, v.LocationLat, v.LocationLng)
	v.drainBattery(kmTraveled)
}

// moveTowardsTarget moves the vehicle towards its target location
func (v *Vehicle) moveTowardsTarget() {
	// Don't move if battery is depleted
	if v.BatteryLevel <= 0 {
		v.handleBatteryDepletion()
		return
	}

	// Store previous position for distance calculation
	prevLat := v.LocationLat
	prevLng := v.LocationLng

	distance := v.distanceToTarget()
	if distance < 0.001 { // Close enough (~100m)
		v.LocationLat = v.targetLat
		v.LocationLng = v.targetLng
		v.isMoving = false
		return
	}

	// Move at configurable city speed
	speed := v.getMovementSpeed() // Configurable driving speed

	// Calculate direction
	latDiff := v.targetLat - v.LocationLat
	lngDiff := v.targetLng - v.LocationLng

	// Normalize and apply speed
	factor := speed / distance
	v.LocationLat += latDiff * factor
	v.LocationLng += lngDiff * factor

	// Drain battery based on actual distance moved using haversine
	kmTraveled := haversineDistance(prevLat, prevLng, v.LocationLat, v.LocationLng)
	v.drainBattery(kmTraveled)
}

// setRandomTarget sets a random target within the specified radius
func (v *Vehicle) setRandomTarget(radiusDegrees float64) {
	angle := rand.Float64() * 2 * math.Pi
	radius := rand.Float64() * radiusDegrees

	v.targetLat = v.LocationLat + radius*math.Cos(angle)
	v.targetLng = v.LocationLng + radius*math.Sin(angle)
}

// distanceToTarget calculates distance to current target
func (v *Vehicle) distanceToTarget() float64 {
	latDiff := v.targetLat - v.LocationLat
	lngDiff := v.targetLng - v.LocationLng
	return math.Sqrt(latDiff*latDiff + lngDiff*lngDiff)
}

// drainBattery reduces battery level and range
func (v *Vehicle) drainBattery(kmTraveled float64) {
	if kmTraveled <= 0 {
		return
	}

	oldBatteryLevel := v.BatteryLevel
	batteryUsedPercent := kmTraveled / v.batteryDrainRate
	newBatteryLevel := math.Max(0, v.BatteryLevel-batteryUsedPercent)
	v.BatteryLevel = newBatteryLevel
	v.BatteryRangeKm = v.BatteryLevel * v.batteryDrainRate

	// Debug logging for battery drain analysis
	slog.Info("Battery drain details",
		"vehicle_id", v.ID,
		"distance_traveled_km", kmTraveled,
		"battery_drained_percent", batteryUsedPercent,
		"battery_before", int(oldBatteryLevel),
		"battery_after", int(v.BatteryLevel),
		"drain_rate_km_per_percent", v.batteryDrainRate,
		"efficiency_actual", kmTraveled/batteryUsedPercent)

	// Handle complete battery depletion
	if v.BatteryLevel == 0 {
		v.handleBatteryDepletion()
	}
}

// handleBatteryDepletion handles when vehicle runs out of battery
func (v *Vehicle) handleBatteryDepletion() {
	slog.Error("Vehicle battery depleted",
		"vehicle_id", v.ID,
		"location_lat", v.LocationLat,
		"location_lng", v.LocationLng,
		"status", v.Status,
		"job_phase", v.jobPhase)

	// If vehicle was going to charge, teleport to nearest charging station
	if v.Status == "charging" && v.jobPhase == "going_to_charge" {
		chargingStation := FindNearestChargingStation(v.LocationLat, v.LocationLng, v.Region)
		v.LocationLat = chargingStation.Lat
		v.LocationLng = chargingStation.Lng
		v.isMoving = false
		v.jobPhase = "charging"
		v.BatteryLevel = 5 // Give minimal charge to start charging process
		v.BatteryRangeKm = v.BatteryLevel * v.batteryDrainRate

		slog.Info("Vehicle teleported to charging station due to battery depletion",
			"vehicle_id", v.ID,
			"station_id", chargingStation.ID,
			"station_lat", chargingStation.Lat,
			"station_lng", chargingStation.Lng)
		return
	}

	// For other cases, stop and set to maintenance
	v.isMoving = false
	v.Status = "maintenance"
	v.jobPhase = "stranded"

	// If had a job, abandon it
	if v.currentJob != nil {
		slog.Warn("Abandoning job due to battery depletion",
			"vehicle_id", v.ID,
			"job_id", v.currentJob.ID)
		v.currentJob = nil
		v.CurrentJobID = nil
	}
}

// goToCharge sets vehicle to charging status and moves to charging station
func (v *Vehicle) goToCharge() {
	// Find nearest charging station
	chargingStation := FindNearestChargingStation(v.LocationLat, v.LocationLng, v.Region)

	v.Status = "charging"
	v.setRouteTarget(chargingStation.Lat, chargingStation.Lng)
	v.jobPhase = "going_to_charge"

	slog.Info("Vehicle going to charging station",
		"vehicle_id", v.ID,
		"charging_station_id", chargingStation.ID,
		"battery_level", v.BatteryLevel,
		"station_lat", chargingStation.Lat,
		"station_lng", chargingStation.Lng,
		"distance_to_station", haversineDistance(v.LocationLat, v.LocationLng, chargingStation.Lat, chargingStation.Lng))
}

// registerWithFleetRetry attempts to register with exponential backoff
func (v *Vehicle) registerWithFleetRetry() error {
	const (
		maxRetries = 10
		baseDelay  = 1 * time.Second
		maxDelay   = 30 * time.Second
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	slog.Info("Starting vehicle registration",
		"vehicle_id", v.ID,
		"max_retries", maxRetries,
		"timeout", "5m")

	var lastErr error
	delay := baseDelay

	for attempt := 0; attempt < maxRetries; attempt++ {
		if err := v.registerWithFleet(); err == nil {
			slog.Info("Vehicle registration successful",
				"vehicle_id", v.ID,
				"attempt", attempt+1)
			return nil // Success
		} else {
			lastErr = err
			slog.Warn("Vehicle registration attempt failed",
				"vehicle_id", v.ID,
				"attempt", attempt+1,
				"max_retries", maxRetries,
				"error", err,
				"next_retry_delay", delay)
		}

		if attempt == maxRetries-1 {
			break // Don't wait after last attempt
		}

		// Use context-aware delay
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			slog.Error("Vehicle registration timeout",
				"vehicle_id", v.ID,
				"context_error", ctx.Err())
			return fmt.Errorf("registration timeout: %v", ctx.Err())
		case <-timer.C:
			// Continue to next attempt
		}

		// Exponential backoff with jitter
		delay = time.Duration(float64(delay) * 1.5)
		if delay > maxDelay {
			delay = maxDelay
		}
	}

	slog.Error("Vehicle registration failed after all retries",
		"vehicle_id", v.ID,
		"max_retries", maxRetries,
		"final_error", lastErr)
	return fmt.Errorf("registration failed after %d attempts: %v", maxRetries, lastErr)
}

// registerWithFleet registers this vehicle with the fleet service
func (v *Vehicle) registerWithFleet() error {
	jsonData, _ := json.Marshal(v)
	url := fmt.Sprintf("%s/vehicles", v.fleetServiceURL)

	slog.Debug("Attempting vehicle registration",
		"vehicle_id", v.ID,
		"fleet_url", url)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error("HTTP request failed during registration",
			"vehicle_id", v.ID,
			"url", url,
			"error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		slog.Error("Vehicle registration rejected",
			"vehicle_id", v.ID,
			"status_code", resp.StatusCode,
			"url", url)
		return fmt.Errorf("failed to register vehicle, status: %d", resp.StatusCode)
	}

	slog.Info("Vehicle registered with fleet service", "vehicle_id", v.ID)
	return nil
}

// reportToFleet sends location update to fleet service
func (v *Vehicle) reportToFleet() {
	locationUpdate := struct {
		Lat    float64 `json:"lat"`
		Lng    float64 `json:"lng"`
		Status string  `json:"status"`
	}{
		Lat:    v.LocationLat,
		Lng:    v.LocationLng,
		Status: v.Status,
	}

	jsonData, _ := json.Marshal(locationUpdate)
	url := fmt.Sprintf("%s/vehicles/%s/location", v.fleetServiceURL, v.ID)

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to report location to fleet service",
			"vehicle_id", v.ID,
			"fleet_url", url,
			"error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("Fleet service location update returned non-OK status",
			"vehicle_id", v.ID,
			"status_code", resp.StatusCode,
			"url", url)
	} else {
		slog.Debug("Successfully reported location to fleet service",
			"vehicle_id", v.ID,
			"lat", v.LocationLat,
			"lng", v.LocationLng)
	}

	// NEW: Also stream to Kinesis (supplemental analytics)
	v.streamVehicleData()
}

// getMovementSpeed returns the movement speed based on environment configuration
func (v *Vehicle) getMovementSpeed() float64 {
	if demoSpeed := os.Getenv("DEMO_SPEED"); demoSpeed != "" {
		if speed, err := strconv.ParseFloat(demoSpeed, 64); err == nil {
			return speed
		}
	}
	return 0.00035 // Default realistic city driving speed (~35 km/h)
}

// logVehicleStatus logs the current status of the vehicle for monitoring
func (v *Vehicle) logVehicleStatus() {
	logData := slog.Group("vehicle_status",
		"vehicle_id", v.ID,
		"status", v.Status,
		"battery_level", v.BatteryLevel,
		"battery_range_km", v.BatteryRangeKm,
		"location_lat", v.LocationLat,
		"location_lng", v.LocationLng,
		"region", v.Region,
		"vehicle_type", v.VehicleType,
		"battery_drain_rate", v.batteryDrainRate,
		"is_moving", v.isMoving,
	)

	// Add job-specific information if busy
	if v.Status == "busy" && v.currentJob != nil {
		logData = slog.Group("vehicle_status",
			"vehicle_id", v.ID,
			"status", v.Status,
			"battery_level", v.BatteryLevel,
			"battery_range_km", v.BatteryRangeKm,
			"location_lat", v.LocationLat,
			"location_lng", v.LocationLng,
			"region", v.Region,
			"vehicle_type", v.VehicleType,
			"battery_drain_rate", v.batteryDrainRate,
			"is_moving", v.isMoving,
			"current_job_id", v.currentJob.ID,
			"job_type", v.currentJob.JobType,
			"job_phase", v.jobPhase,
			"pickup_lat", v.currentJob.PickupLat,
			"pickup_lng", v.currentJob.PickupLng,
			"destination_lat", v.currentJob.DestinationLat,
			"destination_lng", v.currentJob.DestinationLng,
		)
	}

	// Add route information if moving
	if v.isMoving && v.targetLat != 0 && v.targetLng != 0 {
		slog.Info("Vehicle status update", logData,
			"route_target_lat", v.targetLat,
			"route_target_lng", v.targetLng,
			"distance_to_target", v.distanceToTarget(),
		)
	} else {
		slog.Info("Vehicle status update", logData)
	}
}

// initKinesis initializes the Kinesis client if stream name is provided
func (v *Vehicle) initKinesis() {
	streamName := os.Getenv("KINESIS_VEHICLE_TELEMETRY_STREAM")
	if streamName == "" {
		return // Kinesis disabled
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		slog.Warn("Failed to load AWS config for Kinesis", "error", err)
		return
	}

	v.kinesisClient = kinesis.NewFromConfig(cfg)
	v.streamName = streamName
	slog.Info("Kinesis streaming enabled", "vehicle_id", v.ID, "stream", streamName)
}

// streamVehicleData sends vehicle telemetry to Kinesis (supplemental to HTTP API)
func (v *Vehicle) streamVehicleData() {
	if v.kinesisClient == nil {
		return // Kinesis not enabled
	}

	record := map[string]interface{}{
		"vehicle_id": v.ID,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"latitude":   v.LocationLat,
		"longitude":  v.LocationLng,
		"status":     v.Status,
		"battery":    v.BatteryLevel,
	}

	if v.CurrentJobID != nil {
		record["job_id"] = *v.CurrentJobID
	}

	data, err := json.Marshal(record)
	if err != nil {
		slog.Error("Failed to marshal Kinesis record", "vehicle_id", v.ID, "error", err)
		return
	}

	_, err = v.kinesisClient.PutRecord(context.TODO(), &kinesis.PutRecordInput{
		StreamName:   &v.streamName,
		Data:         data,
		PartitionKey: &v.ID,
	})

	if err != nil {
		slog.Error("Failed to send data to Kinesis", "vehicle_id", v.ID, "error", err)
	}
}
