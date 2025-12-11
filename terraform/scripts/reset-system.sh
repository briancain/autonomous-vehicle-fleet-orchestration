#!/bin/bash

echo "🧹 Resetting Fleet Orchestration System..."

# Clear DynamoDB tables
echo "Clearing vehicles table..."
aws dynamodb scan --table-name fleet-orchestration-vehicles --region us-west-2 --select "ALL_ATTRIBUTES" --no-cli-pager | \
jq -r '.Items[] | @base64' | \
while read item; do
    echo $item | base64 --decode | \
    jq -r '{"id": {"S": .id.S}}' | \
    aws dynamodb delete-item --table-name fleet-orchestration-vehicles --key file:///dev/stdin --region us-west-2 --no-cli-pager
done

echo "Clearing jobs table..."
aws dynamodb scan --table-name fleet-orchestration-jobs --region us-west-2 --select "ALL_ATTRIBUTES" --no-cli-pager | \
jq -r '.Items[] | @base64' | \
while read item; do
    echo $item | base64 --decode | \
    jq -r '{"id": {"S": .id.S}}' | \
    aws dynamodb delete-item --table-name fleet-orchestration-jobs --key file:///dev/stdin --region us-west-2 --no-cli-pager
done

echo "Waiting for table clearing to complete..."
sleep 3

# Restart ECS services in proper order
echo "Restarting backend services first..."
aws ecs update-service --cluster fleet-orchestration-cluster --service fleet-orchestration-fleet-service --force-new-deployment --region us-west-2 --no-cli-pager > /dev/null
aws ecs update-service --cluster fleet-orchestration-cluster --service fleet-orchestration-job-service --force-new-deployment --region us-west-2 --no-cli-pager > /dev/null
aws ecs update-service --cluster fleet-orchestration-cluster --service fleet-orchestration-dashboard --force-new-deployment --region us-west-2 --no-cli-pager > /dev/null

echo "Waiting for backend services to stabilize..."
sleep 15

echo "Restarting car simulator last..."
aws ecs update-service --cluster fleet-orchestration-cluster --service fleet-orchestration-car-simulator --force-new-deployment --region us-west-2 --no-cli-pager > /dev/null

echo "✅ System reset complete! Services are restarting..."
echo "🚗 Fresh vehicles will register in ~30 seconds"
echo "📋 New jobs will start generating in ~60 seconds"
