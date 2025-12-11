# Use AWS ECR Public Go base image
FROM public.ecr.aws/docker/library/golang:1.21-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Set Go proxy to direct to bypass proxy issues in AWS
RUN go env -w GOPROXY=direct

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fleet-service cmd/main.go

# Final stage - use minimal AWS ECR base image
FROM public.ecr.aws/docker/library/alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh fleet

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/fleet-service .

# Change ownership to non-root user
RUN chown fleet:fleet /app/fleet-service

# Switch to non-root user
USER fleet

# Expose port
EXPOSE 8080

# Run the service
CMD ["./fleet-service"]
