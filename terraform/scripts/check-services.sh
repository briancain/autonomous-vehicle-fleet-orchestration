#!/bin/bash

# ECS Service Status Checker
# Usage: ./check-services.sh [region]

REGION=${1:-us-west-2}
CLUSTER="fleet-orchestration-cluster"

SERVICES=(
    "fleet-orchestration-fleet-service"
    "fleet-orchestration-job-service" 
    "fleet-orchestration-car-simulator"
    "fleet-orchestration-dashboard"
)

echo "🚀 Fleet Orchestration Services Status"
echo "Cluster: $CLUSTER | Region: $REGION"
echo "========================================"

for service in "${SERVICES[@]}"; do
    echo -n "$(echo $service | sed 's/fleet-orchestration-//' | tr '[:lower:]' '[:upper:]'): "
    
    result=$(aws ecs describe-services \
        --cluster "$CLUSTER" \
        --services "$service" \
        --region "$REGION" \
        --query 'services[0]' \
        --output json 2>/dev/null)
    
    if [ $? -eq 0 ] && [ "$result" != "null" ]; then
        running=$(echo "$result" | jq -r '.runningCount')
        desired=$(echo "$result" | jq -r '.desiredCount') 
        pending=$(echo "$result" | jq -r '.pendingCount')
        rollout=$(echo "$result" | jq -r '.deployments[0].rolloutState')
        
        if [ "$rollout" = "COMPLETED" ] && [ "$running" = "$desired" ] && [ "$pending" = "0" ]; then
            echo "✅ HEALTHY ($running/$desired running)"
        elif [ "$rollout" = "IN_PROGRESS" ]; then
            echo "🔄 DEPLOYING ($running/$desired running, $pending pending)"
        else
            echo "❌ UNHEALTHY ($running/$desired running, $pending pending, $rollout)"
        fi
    else
        echo "❓ NOT FOUND"
    fi
done

echo ""
echo "💡 For detailed logs: aws logs tail /ecs/SERVICE_NAME --region $REGION --follow"
