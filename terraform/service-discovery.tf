# Service Discovery for internal service-to-service communication
# Allows ECS services to communicate via HTTP within the VPC,
# bypassing the ALB (and its self-signed cert) for internal traffic.

# Primary region (us-west-2)
resource "aws_service_discovery_private_dns_namespace" "main" {
  name = "${var.project_name}.local"
  vpc  = aws_vpc.main.id

  tags = {
    Name = "${var.project_name}-sd-namespace"
  }
}

resource "aws_service_discovery_service" "fleet_service" {
  name = "fleet-service"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.main.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }
}

resource "aws_service_discovery_service" "job_service" {
  name = "job-service"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.main.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }
}

# Secondary region (us-west-1)
resource "aws_service_discovery_private_dns_namespace" "secondary" {
  provider = aws.west1
  name     = "${var.project_name}.local"
  vpc      = aws_vpc.secondary.id

  tags = {
    Name = "${var.project_name}-sd-namespace-west1"
  }
}

resource "aws_service_discovery_service" "fleet_service_secondary" {
  provider = aws.west1
  name     = "fleet-service"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.secondary.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }
}

resource "aws_service_discovery_service" "job_service_secondary" {
  provider = aws.west1
  name     = "job-service"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.secondary.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }
}
