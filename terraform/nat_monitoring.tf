# Health checks for primary NAT Gateways
resource "aws_cloudwatch_metric_alarm" "nat_gateway_health" {
  count = 2

  alarm_name          = "${var.project_name}-nat-gateway-${count.index + 1}-health"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "ErrorPortAllocation"
  namespace           = "AWS/NATGateway"
  period              = "300"
  statistic           = "Sum"
  threshold           = "0"
  alarm_description   = "NAT Gateway ${count.index + 1} port allocation errors"
  treat_missing_data  = "notBreaching"

  dimensions = {
    NatGatewayId = aws_nat_gateway.main[count.index].id
  }

  alarm_actions = [aws_sns_topic.nat_alerts.arn]

  tags = {
    Name = "${var.project_name}-nat-gateway-${count.index + 1}-health"
  }
}

resource "aws_sns_topic" "nat_alerts" {
  name              = "${var.project_name}-nat-alerts"
  kms_master_key_id = "alias/aws/sns"

  tags = {
    Name = "${var.project_name}-nat-alerts"
  }
}
