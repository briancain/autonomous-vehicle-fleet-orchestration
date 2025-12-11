# ECS Services in Secondary Region (us-west-1) - STANDBY MODE
# All services have desired_count = 0 until failover

# Data sources for existing ECR repositories
data "aws_ecr_repository" "fleet_service" {
  name = "fleet-orchestration-fleet-service"
}

data "aws_ecr_repository" "job_service" {
  name = "fleet-orchestration-job-service"
}

data "aws_ecr_repository" "car_simulator" {
  name = "fleet-orchestration-car-simulator"
}

data "aws_ecr_repository" "dashboard" {
  name = "fleet-orchestration-dashboard"
}

# CloudWatch log groups for secondary region
resource "aws_cloudwatch_log_group" "fleet_service_secondary" {
  provider          = aws.west1
  name              = "/ecs/${var.project_name}/fleet-service-west1"
  retention_in_days = 7
}

resource "aws_cloudwatch_log_group" "job_service_secondary" {
  provider          = aws.west1
  name              = "/ecs/${var.project_name}/job-service-west1"
  retention_in_days = 7
}

resource "aws_cloudwatch_log_group" "car_simulator_secondary" {
  provider          = aws.west1
  name              = "/ecs/${var.project_name}/car-simulator-west1"
  retention_in_days = 7
}

resource "aws_cloudwatch_log_group" "dashboard_secondary" {
  provider          = aws.west1
  name              = "/ecs/${var.project_name}/dashboard-west1"
  retention_in_days = 7
}

# Task definitions for secondary region
resource "aws_ecs_task_definition" "fleet_service_secondary" {
  provider                 = aws.west1
  family                   = "${var.project_name}-fleet-service-west1"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([
    {
      name  = "fleet-service"
      image = "${data.aws_ecr_repository.fleet_service.repository_url}:latest"

      portMappings = [
        {
          containerPort = 8081
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "AWS_REGION"
          value = "us-west-1"
        },
        {
          name  = "DYNAMODB_TABLE_VEHICLES"
          value = aws_dynamodb_table.vehicles.name
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.fleet_service_secondary.name
          "awslogs-region"        = "us-west-1"
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])
}

resource "aws_ecs_task_definition" "job_service_secondary" {
  provider                 = aws.west1
  family                   = "${var.project_name}-job-service-west1"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([
    {
      name  = "job-service"
      image = "${data.aws_ecr_repository.job_service.repository_url}:latest"

      portMappings = [
        {
          containerPort = 8080
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "AWS_REGION"
          value = "us-west-1"
        },
        {
          name  = "DYNAMODB_TABLE_JOBS"
          value = aws_dynamodb_table.jobs.name
        },
        {
          name  = "DYNAMODB_TABLE_VEHICLES"
          value = aws_dynamodb_table.vehicles.name
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.job_service_secondary.name
          "awslogs-region"        = "us-west-1"
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])
}

resource "aws_ecs_task_definition" "car_simulator_secondary" {
  provider                 = aws.west1
  family                   = "${var.project_name}-car-simulator-west1"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([
    {
      name  = "car-simulator"
      image = "${data.aws_ecr_repository.car_simulator.repository_url}:latest"

      environment = [
        {
          name  = "AWS_REGION"
          value = "us-west-1"
        },
        {
          name  = "FLEET_SERVICE_URL"
          value = "http://${aws_lb.secondary.dns_name}/api/fleet"
        },
        {
          name  = "JOB_SERVICE_URL"
          value = "http://${aws_lb.secondary.dns_name}/api/jobs"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.car_simulator_secondary.name
          "awslogs-region"        = "us-west-1"
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])
}

resource "aws_ecs_task_definition" "dashboard_secondary" {
  provider                 = aws.west1
  family                   = "${var.project_name}-dashboard-west1"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([
    {
      name  = "dashboard"
      image = "${data.aws_ecr_repository.dashboard.repository_url}:latest"

      portMappings = [
        {
          containerPort = 3000
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "REACT_APP_FLEET_SERVICE_URL"
          value = "http://${aws_lb.secondary.dns_name}/api/fleet"
        },
        {
          name  = "REACT_APP_JOB_SERVICE_URL"
          value = "http://${aws_lb.secondary.dns_name}/api/jobs"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.dashboard_secondary.name
          "awslogs-region"        = "us-west-1"
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])
}

# Security group for ECS tasks in secondary region
resource "aws_security_group" "ecs_tasks_secondary" {
  provider    = aws.west1
  name_prefix = "${var.project_name}-ecs-tasks-west1-"
  vpc_id      = aws_vpc.secondary.id

  ingress {
    from_port       = 0
    to_port         = 65535
    protocol        = "tcp"
    security_groups = [aws_security_group.alb_secondary.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-ecs-tasks-sg-west1"
  }
}

# Fleet Service in Secondary Region
resource "aws_ecs_service" "fleet_service_secondary" {
  provider        = aws.west1
  name            = "${var.project_name}-fleet-service"
  cluster         = aws_ecs_cluster.secondary.id
  task_definition = aws_ecs_task_definition.fleet_service_secondary.arn
  desired_count   = 0 # STANDBY MODE
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.private_secondary[*].id
    security_groups  = [aws_security_group.ecs_tasks_secondary.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.fleet_service_secondary.arn
    container_name   = "fleet-service"
    container_port   = 8081
  }

  depends_on = [aws_lb_listener_rule.fleet_service_secondary]
}

# Job Service in Secondary Region
resource "aws_ecs_service" "job_service_secondary" {
  provider        = aws.west1
  name            = "${var.project_name}-job-service"
  cluster         = aws_ecs_cluster.secondary.id
  task_definition = aws_ecs_task_definition.job_service_secondary.arn
  desired_count   = 0 # STANDBY MODE
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.private_secondary[*].id
    security_groups  = [aws_security_group.ecs_tasks_secondary.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.job_service_secondary.arn
    container_name   = "job-service"
    container_port   = 8080
  }

  depends_on = [aws_lb_listener_rule.job_service_secondary]
}

# Car Simulator in Secondary Region
resource "aws_ecs_service" "car_simulator_secondary" {
  provider        = aws.west1
  name            = "${var.project_name}-car-simulator"
  cluster         = aws_ecs_cluster.secondary.id
  task_definition = aws_ecs_task_definition.car_simulator_secondary.arn
  desired_count   = 0 # STANDBY MODE
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.private_secondary[*].id
    security_groups  = [aws_security_group.ecs_tasks_secondary.id]
    assign_public_ip = false
  }
}

# Dashboard in Secondary Region
resource "aws_ecs_service" "dashboard_secondary" {
  provider        = aws.west1
  name            = "${var.project_name}-dashboard"
  cluster         = aws_ecs_cluster.secondary.id
  task_definition = aws_ecs_task_definition.dashboard_secondary.arn
  desired_count   = 0 # STANDBY MODE
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.private_secondary[*].id
    security_groups  = [aws_security_group.ecs_tasks_secondary.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.dashboard_secondary.arn
    container_name   = "dashboard"
    container_port   = 3000
  }
}
