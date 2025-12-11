# Kinesis Data Streams for real-time vehicle telemetry and job events

resource "aws_kinesis_stream" "vehicle_telemetry" {
  name             = "${var.project_name}-vehicle-telemetry"
  shard_count      = 2
  retention_period = 24

  shard_level_metrics = [
    "IncomingRecords",
    "OutgoingRecords",
  ]

  tags = {
    Name = "${var.project_name}-vehicle-telemetry"
  }
}

resource "aws_kinesis_stream" "job_events" {
  name             = "${var.project_name}-job-events"
  shard_count      = 1
  retention_period = 24

  shard_level_metrics = [
    "IncomingRecords",
    "OutgoingRecords",
  ]

  tags = {
    Name = "${var.project_name}-job-events"
  }
}
