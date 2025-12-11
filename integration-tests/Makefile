.PHONY: test test-verbose clean deps build-services

# Run all integration tests
test: build-services
	@echo "Running integration tests..."
	go test ./tests -v -timeout=120s

# Run integration tests with extra verbose output
test-verbose: build-services
	@echo "Running integration tests with verbose output..."
	go test ./tests -v -timeout=120s -args -test.v

# Run specific test
test-end-to-end: build-services
	@echo "Running end-to-end integration test..."
	go test ./tests -v -timeout=120s -run TestEndToEndWorkflow

test-job-assignment: build-services
	@echo "Running job assignment integration test..."
	go test ./tests -v -timeout=120s -run TestJobAssignmentLogic

test-vehicle-lifecycle: build-services
	@echo "Running vehicle lifecycle integration test..."
	go test ./tests -v -timeout=120s -run TestVehicleLifecycle

# Build all service binaries required for integration tests
build-services:
	@echo "Building service binaries for integration tests..."
	@cd ../fleet-service && make build
	@cd ../job-service && make build
	@cd ../car-simulator && make build
	@echo "All service binaries built successfully!"

# Install test dependencies
deps:
	go mod tidy
	go mod download

# Clean test artifacts
clean:
	go clean -testcache
	rm -f integration-tests.log

# Run tests with coverage
test-coverage: build-services
	@echo "Running integration tests with coverage..."
	go test ./tests -v -timeout=120s -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Quick smoke test - just verify services can start
smoke-test: build-services
	@echo "Running smoke test..."
	go test ./tests -v -timeout=60s -run TestEndToEndWorkflow/VehiclesRegistered

# Help
help:
	@echo "Available targets:"
	@echo "  test              - Run all integration tests"
	@echo "  test-verbose      - Run tests with extra verbose output"
	@echo "  test-end-to-end   - Run end-to-end workflow test"
	@echo "  test-job-assignment - Run job assignment logic test"
	@echo "  test-vehicle-lifecycle - Run vehicle lifecycle test"
	@echo "  smoke-test        - Quick test to verify services start"
	@echo "  build-services    - Build all required service binaries"
	@echo "  deps              - Install test dependencies"
	@echo "  clean             - Clean test artifacts"
	@echo "  test-coverage     - Run tests with coverage report"
