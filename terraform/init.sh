#!/bin/bash

# Terraform initialization script with user-specific state file
set -e

# Get the current username and AWS account ID
USERNAME=$(whoami)
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
PROJECT_NAME="autonomous-vehicle-fleet-orchestration"
REGION="us-west-2"

echo "Initializing Terraform for user: $USERNAME"
echo "AWS Account ID: $ACCOUNT_ID"

# Create backend.tf with dynamic values
cat > backend.tf << EOF
terraform {
  backend "s3" {
    bucket = "terraform-state-${ACCOUNT_ID}-${REGION}"
    key    = "${PROJECT_NAME}/terraform.tfstate"
    region = "${REGION}"
    encrypt = true
  }
}
EOF

echo "Created backend.tf:"
cat backend.tf

# Initialize Terraform
terraform init

echo "Terraform initialized successfully with S3 backend"
echo "State bucket: terraform-state-${ACCOUNT_ID}-${REGION}"
echo "State key: ${PROJECT_NAME}/terraform.tfstate"
