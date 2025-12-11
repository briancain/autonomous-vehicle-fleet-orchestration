# Autonomous Vehicle Fleet Orchestration

A comprehensive fleet management system for autonomous vehicles with job generation, vehicle tracking, and real-time orchestration using containerized Go microservices.

This application was created as a demo app for resiliency testing.

## Architecture

- **Fleet Service**: Manages vehicle registration and tracking
- **Job Service**: Handles job creation, assignment, and completion
- **Car Simulator**: Simulates autonomous vehicle behavior
- **Dashboard**: Web interface for monitoring fleet operations
- **Infrastructure**: AWS ECS Fargate with DynamoDB Global Tables

```
Primary Region (us-west-2)                    Secondary Region (us-west-1)
┌─────────────────────────────┐              ┌─────────────────────────────┐
│ ECS Services: ACTIVE        │              │ ECS Services: STOPPED       │
│ - Fleet Service (2)         │              │ - Fleet Service (0)         │
│ - Job Service (2)           │              │ - Job Service (0)           │
│ - Car Simulator (2)         │              │ - Car Simulator (0)         │
│ - Dashboard (2)             │              │ - Dashboard (0)             │
│                             │              │                             │
│ DynamoDB Global Tables      │◄────────────►│ DynamoDB Global Tables      │
│ - vehicles (PRIMARY)        │              │ - vehicles (REPLICA)        │
│ - jobs (PRIMARY)            │              │ - jobs (REPLICA)            │
│                             │              │                             │
└─────────────────────────────┘              └─────────────────────────────┘
              │                                            ▲
              ▼                                            │
      Route53: PRIMARY ──────── Health Checks ─────────────┘ # Doesn't exist yet, but would if this deployed to a real domain.

Data Flow: DynamoDB Global Tables with native multi-region replication
RTO: ~5 minutes | RPO: <1 second | Replication Lag: Sub-second
```

## Prerequisites

- AWS CLI configured with appropriate credentials
- Terraform installed
- Go installed
- Docker installed

## Deployment

### Quick Deploy (Recommended)

From the root project directory:
```bash
cd /Volumes/workplace/NGRHAgenticWorkflowTestsAppLibrary/src/NGRHAgenticWorkflowTestsAppLibrary
./deploy-all.sh deploy autonomous-vehicle-fleet-orchestration
```

### Local Deploy

From this project directory:
```bash
./deploy_all.sh
```

### Manual Terraform Deploy

**Important**: Before running Terraform commands, you must initialize the S3 remote state backend:

```bash
cd terraform
./init.sh                    # Sets up S3 remote state backend
terraform plan
terraform apply
```

The `init.sh` script automatically:
- Creates a user-specific S3 backend configuration
- Uses your AWS account ID and username for isolation
- Sets up the backend at: `terraform-state-{ACCOUNT_ID}-us-west-2`
- Configures state encryption

## Features

- Real-time vehicle fleet monitoring
- Automated job generation and assignment
- Vehicle simulation for testing
- Web dashboard for operations
- Scalable containerized microservices
- User-isolated state management
- Multi-region deployment support with DynamoDB Global Tables
- ECR image replication for disaster recovery

## Services

- **Job Service**: Port 8080 - Job management API
- **Fleet Service**: Port 8081 - Vehicle management API  
- **Car Simulator**: Port 8082 - Vehicle simulation
- **Dashboard**: Port 3000 - Web interface

## Utility Scripts

The `terraform/scripts/` directory contains helpful operational scripts:

- **check-services.sh**: Check ECS service health and status
- **view-logs.sh**: View structured logs from ECS services
- **reset-system.sh**: Clear DynamoDB tables and reset system state
- **push-images.sh**: Build and push Docker images to ECR
- **kinesis-stats.sh**: Monitor Kinesis stream statistics and health

Usage examples:
```bash
cd terraform/scripts
./check-services.sh                   # Check all services
./view-logs.sh fleet-service          # View fleet service logs
./reset-system.sh                     # Reset system state
./push-images.sh latest               # Build and push images
./kinesis-stats.sh                    # View Kinesis metrics
```

## State Management

This project uses Terraform remote state with S3 backend for state isolation:

- **State Bucket**: `terraform-state-{ACCOUNT_ID}-us-west-2`
- **State Key**: `autonomous-vehicle-fleet-orchestration/terraform.tfstate`
- **Encryption**: Enabled by default
- **User Isolation**: Each user gets their own state configuration

The `init.sh` script handles backend configuration automatically by:
1. Detecting your AWS account ID and username
2. Generating a user-specific backend configuration
3. Creating the `backend.tf` file with proper settings
4. Running `terraform init` with the new backend

## Recent Fixes

- Fixed job generator counting logic to only count active jobs (pending + assigned)
- Added comprehensive unit tests for job counting functionality
- Resolved issue where job generation stopped after 25 total jobs
- Updated infrastructure to use DynamoDB instead of RDS PostgreSQL
- Migrated to native DynamoDB Global Tables for multi-region replication
- Removed custom Lambda replication in favor of AWS-managed Global Tables
