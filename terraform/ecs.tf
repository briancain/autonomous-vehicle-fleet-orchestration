resource "aws_ecs_cluster" "main" {
  name = "${var.project_name}-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = {
    Name = "${var.project_name}-cluster"
  }
}

resource "aws_cloudwatch_log_group" "fleet_service" {
  name              = "/ecs/${var.project_name}/fleet-service"
  retention_in_days = 7

  tags = {
    Name = "${var.project_name}-fleet-service-logs"
  }
}

resource "aws_cloudwatch_log_group" "job_service" {
  name              = "/ecs/${var.project_name}/job-service"
  retention_in_days = 7

  tags = {
    Name = "${var.project_name}-job-service-logs"
  }
}

resource "aws_cloudwatch_log_group" "car_simulator" {
  name              = "/ecs/${var.project_name}/car-simulator"
  retention_in_days = 7

  tags = {
    Name = "${var.project_name}-car-simulator-logs"
  }
}

resource "aws_cloudwatch_log_group" "dashboard" {
  name              = "/ecs/${var.project_name}/dashboard"
  retention_in_days = 7

  tags = {
    Name = "${var.project_name}-dashboard-logs"
  }
}

resource "aws_ecs_task_definition" "fleet_service" {
  family                   = "${var.project_name}-fleet-service"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.fleet_service_cpu
  memory                   = var.fleet_service_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([
    {
      name  = "fleet-service"
      image = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/${var.project_name}-fleet-service:latest"

      portMappings = [
        {
          containerPort = 8080
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "PORT"
          value = "8080"
        },
        {
          name  = "STORAGE_TYPE"
          value = "dynamodb"
        },
        {
          name  = "DYNAMODB_VEHICLES_TABLE"
          value = aws_dynamodb_table.vehicles.name
        },
        {
          name  = "AWS_REGION"
          value = var.aws_region
        },
        {
          name  = "PATH_PREFIX"
          value = "/fleet"
        },
        {
          name  = "KINESIS_VEHICLE_TELEMETRY_STREAM"
          value = aws_kinesis_stream.vehicle_telemetry.name
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.fleet_service.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "ecs"
        }
      }

      essential = true
    }
  ])

  tags = {
    Name = "${var.project_name}-fleet-service-task"
  }
}

resource "aws_ecs_task_definition" "job_service" {
  family                   = "${var.project_name}-job-service"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.job_service_cpu
  memory                   = var.job_service_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([
    {
      name  = "job-service"
      image = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/${var.project_name}-job-service:latest"

      portMappings = [
        {
          containerPort = 8081
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "PORT"
          value = "8081"
        },
        {
          name  = "STORAGE_TYPE"
          value = "dynamodb"
        },
        {
          name  = "DYNAMODB_JOBS_TABLE"
          value = aws_dynamodb_table.jobs.name
        },
        {
          name  = "FLEET_SERVICE_URL"
          value = "http://${aws_lb.main.dns_name}/fleet"
        },
        {
          name  = "AWS_REGION"
          value = var.aws_region
        },
        {
          name  = "PATH_PREFIX"
          value = "/jobs"
        },
        {
          name  = "DEMO_MODE"
          value = "true"
        },
        {
          name  = "DEMO_INTERVAL"
          value = "30s"
        },
        {
          name  = "KINESIS_JOB_EVENTS_STREAM"
          value = aws_kinesis_stream.job_events.name
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.job_service.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "ecs"
        }
      }

      essential = true
    }
  ])

  tags = {
    Name = "${var.project_name}-job-service-task"
  }
}

resource "aws_ecs_task_definition" "car_simulator" {
  family                   = "${var.project_name}-car-simulator"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.car_simulator_cpu
  memory                   = var.car_simulator_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([
    {
      name  = "car-simulator"
      image = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/${var.project_name}-car-simulator:latest"

      environment = [
        {
          name  = "FLEET_SERVICE_URL"
          value = "http://${aws_lb.main.dns_name}/fleet"
        },
        {
          name  = "JOB_SERVICE_URL"
          value = "http://${aws_lb.main.dns_name}/jobs"
        },
        {
          name  = "REGION"
          value = var.aws_region
        },
        {
          name  = "VEHICLE_COUNT"
          value = "10"
        },
        {
          name  = "START_LAT"
          value = "45.5152"
        },
        {
          name  = "START_LNG"
          value = "-122.6784"
        },
        {
          name  = "DEMO_SPEED"
          value = "0.0007"
        },
        {
          name  = "KINESIS_VEHICLE_TELEMETRY_STREAM"
          value = aws_kinesis_stream.vehicle_telemetry.name
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.car_simulator.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "ecs"
        }
      }

      essential = true
    }
  ])

  tags = {
    Name = "${var.project_name}-car-simulator-task"
  }
}

resource "aws_ecs_task_definition" "dashboard" {
  family                   = "${var.project_name}-dashboard"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([
    {
      name  = "dashboard"
      image = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/${var.project_name}-dashboard:latest"

      portMappings = [
        {
          containerPort = 3000
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "PORT"
          value = "3000"
        },
        {
          name  = "FLEET_SERVICE_URL"
          value = "http://${aws_lb.main.dns_name}/fleet"
        },
        {
          name  = "JOB_SERVICE_URL"
          value = "http://${aws_lb.main.dns_name}/jobs"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.dashboard.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "ecs"
        }
      }

      essential = true
    }
  ])

  tags = {
    Name = "${var.project_name}-dashboard-task"
  }
}

resource "aws_ecs_service" "fleet_service" {
  name            = "${var.project_name}-fleet-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.fleet_service.arn
  desired_count   = 2 # Scaled to 2 for high availability and resilience
  launch_type     = "FARGATE"

  network_configuration {
    security_groups  = [aws_security_group.ecs_tasks.id]
    subnets          = aws_subnet.private[*].id
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.fleet_service.arn
    container_name   = "fleet-service"
    container_port   = 8080
  }

  depends_on = [aws_lb_listener.main]

  tags = {
    Name = "${var.project_name}-fleet-service"
  }
}

resource "aws_ecs_service" "job_service" {
  name            = "${var.project_name}-job-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.job_service.arn
  desired_count   = 2 # Scaled to 2 for high availability and resilience
  launch_type     = "FARGATE"

  network_configuration {
    security_groups  = [aws_security_group.ecs_tasks.id]
    subnets          = aws_subnet.private[*].id
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.job_service.arn
    container_name   = "job-service"
    container_port   = 8081
  }

  depends_on = [aws_lb_listener.main]

  tags = {
    Name = "${var.project_name}-job-service"
  }
}

resource "aws_ecs_service" "car_simulator" {
  name            = "${var.project_name}-car-simulator"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.car_simulator.arn
  desired_count   = 1 # Single instance to avoid split brain with vehicle IDs
  launch_type     = "FARGATE"

  network_configuration {
    security_groups  = [aws_security_group.ecs_tasks.id]
    subnets          = aws_subnet.private[*].id
    assign_public_ip = false
  }

  tags = {
    Name = "${var.project_name}-car-simulator"
  }
}

resource "aws_ecs_service" "dashboard" {
  name            = "${var.project_name}-dashboard"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.dashboard.arn
  desired_count   = 2 # Scaled to 2 for high availability and resilience
  launch_type     = "FARGATE"

  network_configuration {
    security_groups  = [aws_security_group.ecs_tasks.id]
    subnets          = aws_subnet.private[*].id
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.dashboard.arn
    container_name   = "dashboard"
    container_port   = 3000
  }

  depends_on = [aws_lb_listener.main]

  tags = {
    Name = "${var.project_name}-dashboard"
  }
}

# Auto Scaling Configuration
resource "aws_appautoscaling_target" "fleet_service" {
  max_capacity       = 10
  min_capacity       = 2
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.fleet_service.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"

  tags = {
    Name = "${var.project_name}-fleet-service-scaling-target"
  }
}

resource "aws_appautoscaling_policy" "fleet_service_cpu" {
  name               = "${var.project_name}-fleet-service-cpu-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.fleet_service.resource_id
  scalable_dimension = aws_appautoscaling_target.fleet_service.scalable_dimension
  service_namespace  = aws_appautoscaling_target.fleet_service.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value = 70.0
  }
}

resource "aws_appautoscaling_target" "job_service" {
  max_capacity       = 10
  min_capacity       = 2
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.job_service.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"

  tags = {
    Name = "${var.project_name}-job-service-scaling-target"
  }
}

resource "aws_appautoscaling_policy" "job_service_cpu" {
  name               = "${var.project_name}-job-service-cpu-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.job_service.resource_id
  scalable_dimension = aws_appautoscaling_target.job_service.scalable_dimension
  service_namespace  = aws_appautoscaling_target.job_service.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value = 70.0
  }
}

resource "aws_appautoscaling_target" "dashboard" {
  max_capacity       = 6
  min_capacity       = 2
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.dashboard.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"

  tags = {
    Name = "${var.project_name}-dashboard-scaling-target"
  }
}

resource "aws_appautoscaling_policy" "dashboard_cpu" {
  name               = "${var.project_name}-dashboard-cpu-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.dashboard.resource_id
  scalable_dimension = aws_appautoscaling_target.dashboard.scalable_dimension
  service_namespace  = aws_appautoscaling_target.dashboard.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value = 70.0
  }
}
