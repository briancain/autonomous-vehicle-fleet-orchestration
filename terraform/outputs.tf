output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.main.id
}

output "private_subnet_ids" {
  description = "IDs of the private subnets"
  value       = aws_subnet.private[*].id
}

output "public_subnet_ids" {
  description = "IDs of the public subnets"
  value       = aws_subnet.public[*].id
}

output "alb_dns_name" {
  description = "DNS name of the load balancer"
  value       = aws_lb.main.dns_name
}

output "alb_zone_id" {
  description = "Zone ID of the load balancer"
  value       = aws_lb.main.zone_id
}

output "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  value       = aws_ecs_cluster.main.name
}

output "dynamodb_vehicles_table" {
  description = "Name of the vehicles DynamoDB table"
  value       = aws_dynamodb_table.vehicles.name
}

output "dynamodb_jobs_table" {
  description = "Name of the jobs DynamoDB table"
  value       = aws_dynamodb_table.jobs.name
}

output "dashboard_url" {
  description = "URL for the dashboard"
  value       = "http://${aws_lb.main.dns_name}"
}

output "fleet_service_url" {
  description = "URL for the fleet service"
  value       = "http://${aws_lb.main.dns_name}/fleet"
}

output "job_service_url" {
  description = "URL for the job service"
  value       = "http://${aws_lb.main.dns_name}/jobs"
}

output "kinesis_vehicle_telemetry_stream" {
  description = "Name of the vehicle telemetry Kinesis stream"
  value       = aws_kinesis_stream.vehicle_telemetry.name
}

output "kinesis_job_events_stream" {
  description = "Name of the job events Kinesis stream"
  value       = aws_kinesis_stream.job_events.name
}
