package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"job-service/internal/fleet"
	"job-service/internal/handlers"
	"job-service/internal/kinesis"
	"job-service/internal/service"
	"job-service/internal/storage"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	kinesisService "github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/gorilla/mux"
)

func main() {
	// Setup structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Get configuration from environment
	fleetServiceURL := getEnv("FLEET_SERVICE_URL", "http://localhost:8080")
	port := getEnv("PORT", "8081")
	demoMode := getEnv("DEMO_MODE", "false") == "true"
	demoInterval := getEnvDuration("DEMO_INTERVAL", "15s")
	storageType := getEnv("STORAGE_TYPE", "memory")

	// Initialize storage based on configuration
	var jobStorage storage.JobStorage
	switch storageType {
	case "dynamodb":
		tableName := getEnv("DYNAMODB_JOBS_TABLE", "fleet-jobs")
		region := getEnv("AWS_REGION", "us-west-2")

		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
		if err != nil {
			slog.Error("Failed to load AWS config", "error", err)
			os.Exit(1)
		}

		dynamoClient := dynamodb.NewFromConfig(cfg)
		jobStorage = storage.NewDynamoDBJobStorage(dynamoClient, tableName)
		slog.Info("Using DynamoDB storage", "table_name", tableName)
	default:
		jobStorage = storage.NewMemoryJobStorage()
		slog.Info("Using in-memory storage")
	}

	// Initialize fleet client
	fleetClient := fleet.NewClient(fleetServiceURL)

	// Initialize service
	jobService := service.NewJobService(jobStorage, fleetClient)

	// Initialize Kinesis streamer if stream name is provided
	if streamName := getEnv("KINESIS_JOB_EVENTS_STREAM", ""); streamName != "" {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			slog.Warn("Failed to load AWS config for Kinesis", "error", err)
		} else {
			kinesisClient := kinesisService.NewFromConfig(cfg)
			streamer := kinesis.NewStreamer(kinesisClient, streamName)
			jobService.SetKinesisStreamer(streamer)
			slog.Info("Kinesis job event streaming enabled", "stream", streamName)
		}
	}

	// Initialize background job processor
	jobProcessor := service.NewJobProcessor(jobService)
	jobProcessor.Start()
	defer jobProcessor.Stop()

	// Initialize demo job generator
	var demoGenerator *service.DemoJobGenerator
	var demoHandler *handlers.DemoHandler

	if demoMode {
		demoGenerator = service.NewDemoJobGenerator(jobService, demoInterval)
		demoHandler = handlers.NewDemoHandler(demoGenerator)
		demoGenerator.Start() // Auto-start in demo mode
		slog.Info("Demo mode enabled", "job_generation_interval", demoInterval)
	}

	// Initialize HTTP handlers
	httpHandler := handlers.NewHTTPHandler(jobService)

	// Setup routes
	router := mux.NewRouter()

	// Use path prefix if running behind load balancer
	pathPrefix := os.Getenv("PATH_PREFIX")
	if pathPrefix != "" {
		jobsRouter := router.PathPrefix(pathPrefix).Subrouter()
		httpHandler.RegisterRoutes(jobsRouter)
	} else {
		httpHandler.RegisterRoutes(router)
	}

	// Add demo routes if demo mode is available
	if demoHandler != nil {
		router.HandleFunc("/demo/start", demoHandler.StartDemo).Methods("POST")
		router.HandleFunc("/demo/stop", demoHandler.StopDemo).Methods("POST")
		router.HandleFunc("/demo/status", demoHandler.GetDemoStatus).Methods("GET")
	}

	// Add CORS middleware for frontend
	router.Use(corsMiddleware)

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		slog.Info("Job Service starting", "port", port, "fleet_service_url", fleetServiceURL)
		if err := http.ListenAndServe(":"+port, router); err != nil {
			slog.Error("Job Service failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-c
	slog.Info("Job Service shutting down")
	if demoGenerator != nil {
		demoGenerator.Stop()
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvDuration gets duration from environment variable
func getEnvDuration(key, defaultValue string) time.Duration {
	value := getEnv(key, defaultValue)
	duration, err := time.ParseDuration(value)
	if err != nil {
		slog.Warn("Invalid duration, using default", "provided", value, "default", defaultValue, "error", err)
		duration, _ = time.ParseDuration(defaultValue)
	}
	return duration
}

// corsMiddleware adds CORS headers for frontend access
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
