variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "project_name" {
  description = "Project name"
  type        = string
  default     = "fleet-orchestration"
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "fleet_service_cpu" {
  description = "CPU units for fleet service"
  type        = number
  default     = 256
}

variable "fleet_service_memory" {
  description = "Memory for fleet service"
  type        = number
  default     = 512
}

variable "job_service_cpu" {
  description = "CPU units for job service"
  type        = number
  default     = 256
}

variable "job_service_memory" {
  description = "Memory for job service"
  type        = number
  default     = 512
}

variable "car_simulator_cpu" {
  description = "CPU units for car simulator"
  type        = number
  default     = 256
}

variable "car_simulator_memory" {
  description = "Memory for car simulator"
  type        = number
  default     = 512
}

variable "desired_count" {
  description = "Desired number of tasks"
  type        = number
  default     = 2
}
