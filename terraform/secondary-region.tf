# Secondary Region Infrastructure (us-west-1)
# Identical to primary but with ECS services at desired_count = 0

# VPC for secondary region
resource "aws_vpc" "secondary" {
  provider             = aws.west1
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "${var.project_name}-vpc-west1"
  }
}

resource "aws_internet_gateway" "secondary" {
  provider = aws.west1
  vpc_id   = aws_vpc.secondary.id

  tags = {
    Name = "${var.project_name}-igw-west1"
  }
}

resource "aws_subnet" "public_secondary" {
  provider                = aws.west1
  count                   = 2
  vpc_id                  = aws_vpc.secondary.id
  cidr_block              = "10.0.${count.index + 1}.0/24"
  availability_zone       = data.aws_availability_zones.secondary.names[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.project_name}-public-subnet-${count.index + 1}-west1"
  }
}

resource "aws_subnet" "private_secondary" {
  provider          = aws.west1
  count             = 2
  vpc_id            = aws_vpc.secondary.id
  cidr_block        = "10.0.${count.index + 10}.0/24"
  availability_zone = data.aws_availability_zones.secondary.names[count.index]

  tags = {
    Name = "${var.project_name}-private-subnet-${count.index + 1}-west1"
  }
}

resource "aws_route_table" "public_secondary" {
  provider = aws.west1
  vpc_id   = aws_vpc.secondary.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.secondary.id
  }

  tags = {
    Name = "${var.project_name}-public-rt-west1"
  }
}

resource "aws_route_table_association" "public_secondary" {
  provider       = aws.west1
  count          = 2
  subnet_id      = aws_subnet.public_secondary[count.index].id
  route_table_id = aws_route_table.public_secondary.id
}

# NAT Gateways for secondary region
resource "aws_eip" "nat_secondary" {
  provider = aws.west1
  count    = 2
  domain   = "vpc"

  tags = {
    Name = "${var.project_name}-nat-eip-${count.index + 1}-west1"
  }

  depends_on = [aws_internet_gateway.secondary]
}

resource "aws_nat_gateway" "secondary" {
  provider      = aws.west1
  count         = 2
  allocation_id = aws_eip.nat_secondary[count.index].id
  subnet_id     = aws_subnet.public_secondary[count.index].id

  tags = {
    Name = "${var.project_name}-nat-${count.index + 1}-west1"
  }

  depends_on = [aws_internet_gateway.secondary]
}

# Private route tables for secondary region
resource "aws_route_table" "private_secondary" {
  provider = aws.west1
  count    = 2
  vpc_id   = aws_vpc.secondary.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.secondary[count.index].id
  }

  tags = {
    Name = "${var.project_name}-private-rt-${count.index + 1}-west1"
  }
}

resource "aws_route_table_association" "private_secondary" {
  provider       = aws.west1
  count          = 2
  subnet_id      = aws_subnet.private_secondary[count.index].id
  route_table_id = aws_route_table.private_secondary[count.index].id
}

data "aws_availability_zones" "secondary" {
  provider = aws.west1
  state    = "available"
}

# ECS Cluster for secondary region
resource "aws_ecs_cluster" "secondary" {
  provider = aws.west1
  name     = "${var.project_name}-cluster-west1"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

# ALB for secondary region
resource "aws_lb" "secondary" {
  provider           = aws.west1
  name               = "${var.project_name}-alb-west1"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb_secondary.id]
  subnets            = aws_subnet.public_secondary[*].id

  enable_deletion_protection = false
}

# Security group for secondary ALB
resource "aws_security_group" "alb_secondary" {
  provider    = aws.west1
  name_prefix = "${var.project_name}-alb-west1-"
  vpc_id      = aws_vpc.secondary.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# Target groups for secondary region services
resource "aws_lb_target_group" "fleet_service_secondary" {
  provider    = aws.west1
  name        = "${var.project_name}-fleet-west1"
  port        = 8081
  protocol    = "HTTP"
  vpc_id      = aws_vpc.secondary.id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = "/health"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }
}

resource "aws_lb_target_group" "job_service_secondary" {
  provider    = aws.west1
  name        = "${var.project_name}-job-west1"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = aws_vpc.secondary.id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = "/health"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }
}

resource "aws_lb_target_group" "dashboard_secondary" {
  provider    = aws.west1
  name        = "${var.project_name}-dash-west1"
  port        = 3000
  protocol    = "HTTP"
  vpc_id      = aws_vpc.secondary.id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = "/"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }
}

# ALB Listener for secondary region
resource "aws_lb_listener" "secondary" {
  provider          = aws.west1
  load_balancer_arn = aws_lb.secondary.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.dashboard_secondary.arn
  }
}

# ALB Listener rules for secondary region
resource "aws_lb_listener_rule" "fleet_service_secondary" {
  provider     = aws.west1
  listener_arn = aws_lb_listener.secondary.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.fleet_service_secondary.arn
  }

  condition {
    path_pattern {
      values = ["/api/fleet/*"]
    }
  }
}

resource "aws_lb_listener_rule" "job_service_secondary" {
  provider     = aws.west1
  listener_arn = aws_lb_listener.secondary.arn
  priority     = 200

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.job_service_secondary.arn
  }

  condition {
    path_pattern {
      values = ["/api/jobs/*"]
    }
  }
}
