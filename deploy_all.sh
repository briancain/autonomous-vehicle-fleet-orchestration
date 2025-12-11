#!/bin/bash

# Change to the directory where this script is located
cd "$(dirname "$0")"

# Autonomous Vehicle Fleet Orchestration Deployment Script
set -e

USER=${USER:-$(whoami)}
REGION="us-west-2"
PROJECT_NAME="autonomous-vehicle-fleet-orchestration"
BUCKET_NAME="${PROJECT_NAME}-terraform-state-${USER}"
DYNAMODB_TABLE="${PROJECT_NAME}-terraform-locks-${USER}"

echo "🚗 Autonomous Vehicle Fleet Orchestration Deployment"
echo "===================================================="
echo "User: ${USER}"
echo "Region: ${REGION}"
echo ""

# Check if AWS credentials are configured
if ! aws sts get-caller-identity &> /dev/null; then
    echo "❌ AWS credentials are not configured or have expired."
    echo "Please configure your AWS credentials and try again."
    exit 1
fi

echo "✅ AWS credentials verified"

# Check if required tools are installed
for tool in terraform go docker; do
    if ! command -v $tool &> /dev/null; then
        echo "❌ $tool is not installed. Please install it first."
        exit 1
    fi
done

echo "✅ Required tools verified"

# Build Go services
echo "🔨 Building Go services..."
cd job-service && go build -o ../terraform/job-service . && cd ..
cd fleet-service && go build -o ../terraform/fleet-service . && cd ..
cd car-simulator && go build -o ../terraform/car-simulator . && cd ..

echo "✅ Go services built successfully"

# Build and push Docker images if needed
echo "🐳 Building Docker images..."
(cd terraform/scripts && ./push-images.sh latest)

echo "✅ Docker images built successfully"

# Deploy infrastructure
echo "🏗️  Deploying infrastructure..."
cd terraform

# Initialize Terraform with user-specific backend
if [ ! -d ".terraform" ]; then
    echo "Initializing Terraform..."
    TEMP_BACKEND_CONFIG=$(mktemp)
    cat > "$TEMP_BACKEND_CONFIG" << EOF
bucket = "$BUCKET_NAME"
key    = "$PROJECT_NAME/terraform.tfstate"
region = "$REGION"
encrypt = true
dynamodb_table = "$DYNAMODB_TABLE"
EOF

    terraform init -backend-config="$TEMP_BACKEND_CONFIG"
    rm "$TEMP_BACKEND_CONFIG"
fi

# Plan deployment
echo "Creating Terraform plan..."
terraform plan -out="tfplan-$USER"

# Apply deployment
if [ "${SKIP_CONFIRMATION:-false}" = "true" ]; then
    echo "Applying Terraform plan (auto-approved)..."
    terraform apply -auto-approve "tfplan-$USER"
else
    read -p "Do you want to apply this plan? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Applying Terraform plan..."
        terraform apply "tfplan-$USER"
    else
        echo "Deployment cancelled."
        rm -f "tfplan-$USER"
        exit 0
    fi
fi

# Clean up plan file
rm -f "tfplan-$USER"

echo ""
echo "🎉 Deployment completed successfully!"
echo "State bucket: $BUCKET_NAME"
echo "State key: $PROJECT_NAME/terraform.tfstate"
