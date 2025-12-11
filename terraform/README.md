# Fleet Orchestration Terraform Infrastructure

This Terraform configuration deploys the complete autonomous vehicle fleet orchestration system to AWS with multi-region disaster recovery capabilities.

## Architecture

- **Multi-Region**: Primary (us-west-2) and Secondary (us-west-1) regions
- **VPC**: Multi-AZ setup with public/private subnets in both regions
- **ECS Fargate**: Container orchestration for services
- **Application Load Balancer**: HTTP routing and load balancing in both regions
- **DynamoDB Global Tables**: Native multi-region replication with sub-second latency
- **ECR Replication**: Automatic cross-region container image replication
- **CloudWatch**: Logging and monitoring in both regions
- **Kinesis Streams**: Real-time data streaming (primary region)

## Multi-Region Strategy

**Primary Region (us-west-2)**: Active deployment
- All ECS services running with desired task counts
- DynamoDB tables in PRIMARY mode
- Kinesis streams for real-time data

**Secondary Region (us-west-1)**: Standby deployment
- ECS services deployed but scaled to 0 tasks
- DynamoDB tables in REPLICA mode (auto-synced)
- Infrastructure ready for immediate failover

**Disaster Recovery Metrics**:
- RTO: ~5 minutes (time to scale up secondary services)
- RPO: <1 second (DynamoDB Global Tables replication)

## Prerequisites

1. AWS CLI configured with appropriate credentials
2. Terraform >= 1.0 installed
3. Finch (or Docker) installed

## Deployment

1. **Push Docker images to ECR:**
   ```bash
   ./scripts/push-images.sh
   ```

2. **Initialize Terraform with user-specific state:**
   ```bash
   ./init.sh
   ```
   This creates a user-isolated S3 backend configuration.

3. **Create terraform.tfvars (optional):**
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   # Edit terraform.tfvars with your values
   ```

4. **Plan deployment:**
   ```bash
   terraform plan
   ```

5. **Deploy infrastructure:**
   ```bash
   terraform apply
   ```

## What Gets Deployed

### Primary Region (us-west-2)
- VPC with 2 public and 2 private subnets
- Application Load Balancer
- ECS Cluster with 4 services (Fleet, Job, Dashboard, Car Simulator)
- DynamoDB Global Tables (PRIMARY)
- NAT Gateways with CloudWatch monitoring
- Kinesis Streams for telemetry
- ECR repositories with replication enabled

### Secondary Region (us-west-1)
- VPC with 2 public and 2 private subnets
- Application Load Balancer
- ECS Cluster with 4 services (scaled to 0)
- DynamoDB Global Tables (REPLICA - auto-synced)
- NAT Gateways

### Global Resources
- DynamoDB Global Tables (vehicles, jobs)
- ECR replication configuration
- IAM roles and policies

## Manual ECR Setup (Alternative)

If you prefer manual setup:

```bash
# Create ECR repositories
aws ecr create-repository --repository-name fleet-orchestration-fleet-service
aws ecr create-repository --repository-name fleet-orchestration-job-service  
aws ecr create-repository --repository-name fleet-orchestration-car-simulator
aws ecr create-repository --repository-name fleet-orchestration-dashboard

# Get login token
aws ecr get-login-password --region us-west-2 | finch login --username AWS --password-stdin <account-id>.dkr.ecr.us-west-2.amazonaws.com

# Build and push images
finch build --platform linux/amd64 -t fleet-service ../fleet-service
finch tag fleet-service:latest <account-id>.dkr.ecr.us-west-2.amazonaws.com/fleet-orchestration-fleet-service:latest
finch push <account-id>.dkr.ecr.us-west-2.amazonaws.com/fleet-orchestration-fleet-service:latest

# Repeat for job-service, car-simulator, and dashboard
```

## Outputs

After deployment, Terraform will output:
- Primary region load balancer DNS name
- Service URLs (dashboard, fleet, job)
- DynamoDB Global Table names
- Kinesis stream names
- VPC and subnet IDs

## Failover Process

To failover to secondary region:

1. **Scale up secondary services:**
   ```bash
   aws ecs update-service --region us-west-1 --cluster fleet-orchestration-cluster-west1 \
     --service fleet-orchestration-fleet-service --desired-count 2
   
   aws ecs update-service --region us-west-1 --cluster fleet-orchestration-cluster-west1 \
     --service fleet-orchestration-job-service --desired-count 2
   
   aws ecs update-service --region us-west-1 --cluster fleet-orchestration-cluster-west1 \
     --service fleet-orchestration-dashboard --desired-count 2
   
   aws ecs update-service --region us-west-1 --cluster fleet-orchestration-cluster-west1 \
     --service fleet-orchestration-car-simulator --desired-count 1
   ```

2. **Update DNS** (if using Route53 - not currently configured)

3. **Verify services are healthy:**
   ```bash
   ./scripts/check-services.sh
   ```

## Utility Scripts

The `scripts/` directory contains operational tools:

- **check-services.sh**: Check ECS service health and status
- **view-logs.sh**: View structured logs from ECS services
- **reset-system.sh**: Clear DynamoDB tables and reset system state
- **push-images.sh**: Build and push Docker images to ECR
- **kinesis-stats.sh**: Monitor Kinesis stream statistics and health

## State Management

This project uses Terraform remote state with S3 backend:

- **State Bucket**: `terraform-state-{ACCOUNT_ID}-us-west-2`
- **State Key**: `autonomous-vehicle-fleet-orchestration/terraform.tfstate`
- **Encryption**: Enabled by default
- **User Isolation**: Each user gets their own state configuration

The `init.sh` script handles backend configuration automatically.

## Cleanup

```bash
terraform destroy
```

Note: This will destroy resources in both regions.

## Configuration

Key variables in `terraform.tfvars`:
- `aws_region`: Primary AWS region (default: us-west-2)
- `environment`: Environment name (default: dev)
- `project_name`: Project identifier (default: fleet-orchestration)
- `vpc_cidr`: VPC CIDR block (default: 10.0.0.0/16)

## Architecture Decisions

**Why DynamoDB Global Tables?**
- Native AWS-managed multi-region replication
- Sub-second replication latency
- No custom Lambda code to maintain
- Production-ready and battle-tested
- Automatic conflict resolution
- Active-active capable (currently configured as active-standby)

**Why Active-Standby vs Active-Active?**
- Simpler operational model
- Avoids potential write conflicts
- Lower cost (secondary services scaled to 0)
- Fast failover when needed (~5 minutes RTO)
