#!/bin/bash

# View structured logs from ECS services
# Usage: ./view-logs.sh [service] [region]

REGION=${2:-us-west-2}
SERVICE=${1}

if [ -z "$SERVICE" ]; then
    echo "📋 Available services:"
    echo "  fleet-service"
    echo "  job-service" 
    echo "  car-simulator"
    echo "  dashboard"
    echo ""
    echo "Usage: ./view-logs.sh <service> [region]"
    echo "Example: ./view-logs.sh car-simulator"
    exit 1
fi

LOG_GROUP="/ecs/fleet-orchestration-${SERVICE}"

echo "📊 Viewing logs for: $SERVICE"
echo "Log Group: $LOG_GROUP"
echo "Region: $REGION"
echo "========================================"

aws logs tail "$LOG_GROUP" \
    --region "$REGION" \
    --follow \
    --format short
