#!/bin/bash

# Build and push all fleet orchestration images to ECR
# Usage: ./push-images.sh [tag]

set -e

# Configuration
REGION="us-west-2"
TAG=${1:-latest}
PROJECT_NAME="fleet-orchestration"

echo "🚀 Building and pushing fleet orchestration images to ECR..."
echo "Region: $REGION"
echo "Tag: $TAG"

# Get AWS account ID
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text --no-cli-pager)
if [ -z "$ACCOUNT_ID" ]; then
    echo "❌ Error: Could not get AWS account ID"
    exit 1
fi

echo "Account ID: $ACCOUNT_ID"

# Services to build and push
SERVICES=("fleet-service" "job-service" "car-simulator" "dashboard")

# Create ECR repositories if they don't exist
echo "🏗️  Creating ECR repositories..."
for service in "${SERVICES[@]}"; do
    repo_name="${PROJECT_NAME}-${service}"
    echo "Creating repository: $repo_name"
    aws ecr create-repository --repository-name "$repo_name" --region "$REGION" --no-cli-pager 2>/dev/null || echo "Repository $repo_name already exists"
done

# Authenticate with ECR
echo "🔐 Authenticating with ECR..."
aws ecr get-login-password --region "$REGION" --no-cli-pager | finch login --username AWS --password-stdin "${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com"

# Build and push each service
cd "$(dirname "$0")/../.."

for service in "${SERVICES[@]}"; do
    echo ""
    echo "🔨 Building $service..."
    
    # Build image for amd64 architecture
    finch build --platform linux/amd64 -t "${service}:${TAG}" "./${service}/"
    
    # Tag for ECR
    ecr_url="${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${PROJECT_NAME}-${service}"
    echo "🏷️  Tagging $service for ECR..."
    finch tag "${service}:${TAG}" "${ecr_url}:${TAG}"
    
    # Push to ECR
    echo "📤 Pushing $service to ECR..."
    finch push "${ecr_url}:${TAG}"
    
    echo "✅ $service pushed successfully!"
    echo "Image: ${ecr_url}:${TAG}"
done

echo ""
echo "🎉 All images pushed successfully!"
echo "Ready to deploy with: terraform apply"
