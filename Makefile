.PHONY: build-all test clean fmt deps build-linux build-images dev help dashboard demo

# Build all services
build-all:
	@echo "Building Fleet Service..."
	cd fleet-service && make build
	@echo "Building Job Service..."
	cd job-service && make build
	@echo "Building Car Simulator..."
	cd car-simulator && make build
	@echo "Building Dashboard..."
	cd dashboard && npm run build
	@echo "All services built successfully!"

# Build and push Docker images to ECR
build-images:
	@echo "Building and pushing Docker images..."
	cd terraform/scripts && ./push-images.sh latest
	@echo "Docker images built and pushed successfully!"

# Test all services
test:
	@echo "Testing Fleet Service..."
	cd fleet-service && make test
	@echo "Testing Job Service..."
	cd job-service && make test
	@echo "Testing Car Simulator..."
	cd car-simulator && make test
	@echo "All tests passed!"

# Clean all build artifacts
clean:
	@echo "Cleaning Fleet Service..."
	cd fleet-service && make clean
	@echo "Cleaning Job Service..."
	cd job-service && make clean
	@echo "Cleaning Car Simulator..."
	cd car-simulator && make clean
	@echo "All build artifacts cleaned!"

# Format all code
fmt:
	@echo "Formatting Fleet Service..."
	cd fleet-service && make fmt
	@echo "Formatting Job Service..."
	cd job-service && make fmt
	@echo "Formatting Car Simulator..."
	cd car-simulator && make fmt
	@echo "Formatting Terraform..."
	terraform fmt -recursive terraform/
	@echo "All code formatted!"

# Install dependencies for all services
deps:
	@echo "Installing Fleet Service dependencies..."
	cd fleet-service && make deps
	@echo "Installing Job Service dependencies..."
	cd job-service && make deps
	@echo "Installing Car Simulator dependencies..."
	cd car-simulator && make deps
	@echo "All dependencies installed!"

# Build for Linux (useful for Docker)
build-linux:
	@echo "Building Fleet Service for Linux..."
	cd fleet-service && make build-linux
	@echo "Building Job Service for Linux..."
	cd job-service && make build-linux
	@echo "Building Car Simulator for Linux..."
	cd car-simulator && make build-linux
	@echo "All Linux binaries built!"

# Quick development workflow
dev: deps fmt test build-all
	@echo "Validating Terraform configuration..."
	cd terraform && terraform validate
	@echo "Development build complete!"

# Build and setup dashboard
dashboard:
	@echo "Setting up dashboard..."
	cd dashboard && make dev
	@echo "Dashboard ready!"

# Start all services and open dashboard (for local testing)
demo: build-all dashboard
	@echo "Starting Fleet Orchestration Demo..."
	@echo "Starting Fleet Service on port 8080..."
	cd fleet-service && nohup ./bin/fleet-service > fleet-service.log 2>&1 &
	@sleep 3
	@echo "Starting Job Service on port 8081..."
	cd job-service && nohup env DEMO_MODE=true DEMO_INTERVAL=30s PORT=8081 ./bin/job-service > job-service.log 2>&1 &
	@sleep 3
	@echo "Starting Car Simulator in Portland..."
	cd car-simulator && nohup env START_LAT=45.5152 START_LNG=-122.6784 VEHICLE_COUNT=10 REGION=us-west-2 DEMO_SPEED=0.0007 ./bin/car-simulator > car-simulator.log 2>&1 &
	@sleep 3
	@echo "Starting Dashboard Server on port 3000..."
	cd dashboard && nohup node server.js > dashboard.log 2>&1 &
	@sleep 3
	@echo "Waiting for services to be ready..."
	@for i in 1 2 3 4 5; do \
		if curl -s http://localhost:8080/health > /dev/null 2>&1; then \
			echo "Fleet service is ready!"; \
			break; \
		fi; \
		echo "Waiting for fleet service... ($$i/5)"; \
		sleep 2; \
	done
	@echo "Opening dashboard in browser..."
	open http://localhost:3000 || xdg-open http://localhost:3000 || start http://localhost:3000
	@echo ""
	@echo "Demo started! Services running in background."
	@echo "Location: Portland, OR with 10 vehicles"
	@echo "Dashboard: http://localhost:3000"
	@echo "Logs: fleet-service.log, job-service.log, car-simulator.log, dashboard.log"
	@echo "To stop all services: make stop-demo"

# Stop all demo services
stop-demo:
	@echo "Stopping all services..."
	@pkill -f fleet-service || true
	@pkill -f job-service || true
	@pkill -f car-simulator || true
	@pkill -f "node server.js" || true
	@echo "Cleaning up log files..."
	@rm -f fleet-service/fleet-service.log job-service/job-service.log car-simulator/car-simulator.log dashboard/dashboard.log
	@echo "All services stopped and logs cleaned."

# Show help
help:
	@echo "Autonomous Vehicle Fleet Orchestration - Available targets:"
	@echo ""
	@echo "Main targets:"
	@echo "  build-all     - Build all services (Fleet, Job, Car Simulator)"
	@echo "  build-images  - Build and push Docker images to ECR"
	@echo "  test          - Run all unit tests"
	@echo "  clean         - Clean all build artifacts"
	@echo "  fmt           - Format all Go code"
	@echo "  deps          - Install all dependencies"
	@echo "  dev           - Complete development workflow (deps + fmt + test + build)"
	@echo ""
	@echo "Demo & Dashboard:"
	@echo "  dashboard     - Build and setup dashboard"
	@echo "  demo          - Start all services and open dashboard in browser"
	@echo "  stop-demo     - Stop all running services"
	@echo ""
	@echo "Platform-specific:"
	@echo "  build-linux   - Build all services for Linux"
	@echo ""
	@echo "Individual services:"
	@echo "  cd fleet-service && make help    - Fleet Service commands"
	@echo "  cd job-service && make help      - Job Service commands"
	@echo "  cd car-simulator && make help    - Car Simulator commands"
	@echo ""
	@echo "Integration tests:"
	@echo "  cd integration-tests && make help - Integration test commands"
	@echo ""
	@echo "  help          - Show this help message"
