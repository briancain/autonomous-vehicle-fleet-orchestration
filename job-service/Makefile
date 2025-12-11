.PHONY: build test clean run build-linux build-darwin fmt lint deps help

# Build the job service binary
build:
	mkdir -p bin
	go build -o bin/job-service cmd/main.go

# Run tests
test:
	go test ./... -v

# Clean build artifacts
clean:
	rm -rf bin/

# Run the service locally
run: build
	./bin/job-service

# Build for different platforms
build-linux:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/job-service-linux cmd/main.go

build-darwin:
	mkdir -p bin
	GOOS=darwin GOARCH=amd64 go build -o bin/job-service-darwin cmd/main.go

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod tidy
	go mod download

# Show help
help:
	@echo "Job Service - Available targets:"
	@echo "  build         - Build the job service binary"
	@echo "  test          - Run unit tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  run           - Build and run the service locally"
	@echo "  build-linux   - Build for Linux platform"
	@echo "  build-darwin  - Build for macOS platform"
	@echo "  fmt           - Format Go code"
	@echo "  lint          - Run linter"
	@echo "  deps          - Install dependencies"
	@echo "  help          - Show this help message"
