package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"fleet-service/internal/handlers"
	"fleet-service/internal/kinesis"
	"fleet-service/internal/service"
	"fleet-service/internal/storage"

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

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		slog.Error("Failed to load AWS config", "error", err)
		os.Exit(1)
	}

	// Initialize storage based on environment
	var vehicleStorage storage.VehicleStorage
	storageType := os.Getenv("STORAGE_TYPE")

	if storageType == "dynamodb" {
		// Create DynamoDB client
		dynamoClient := dynamodb.NewFromConfig(cfg)
		tableName := os.Getenv("DYNAMODB_VEHICLES_TABLE")
		if tableName == "" {
			slog.Error("DYNAMODB_VEHICLES_TABLE environment variable not set")
			os.Exit(1)
		}

		vehicleStorage = storage.NewDynamoDBVehicleStorage(dynamoClient, tableName)
		slog.Info("Using DynamoDB storage", "table", tableName)
	} else {
		vehicleStorage = storage.NewMemoryVehicleStorage()
		slog.Info("Using in-memory storage")
	}

	// Initialize service
	fleetService := service.NewFleetService(vehicleStorage)

	// Start Kinesis consumer if stream name is provided
	if streamName := os.Getenv("KINESIS_VEHICLE_TELEMETRY_STREAM"); streamName != "" {
		kinesisClient := kinesisService.NewFromConfig(cfg)
		consumer := kinesis.NewConsumer(kinesisClient, streamName, fleetService)
		go consumer.Start(context.Background())
	}

	// Initialize HTTP handlers
	httpHandler := handlers.NewHTTPHandler(fleetService)

	// Setup routes
	router := mux.NewRouter()

	// Use path prefix if running behind load balancer
	pathPrefix := os.Getenv("PATH_PREFIX")
	if pathPrefix != "" {
		fleetRouter := router.PathPrefix(pathPrefix).Subrouter()
		httpHandler.RegisterRoutes(fleetRouter)
	} else {
		httpHandler.RegisterRoutes(router)
	}

	// Add CORS middleware for frontend
	router.Use(corsMiddleware)

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("Fleet Service starting", "port", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		slog.Error("Fleet Service failed to start", "error", err)
		os.Exit(1)
	}
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
