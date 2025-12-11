# ECR Cross-Region Replication
# Automatically replicates all ECR repositories from us-west-2 to us-west-1

resource "aws_ecr_replication_configuration" "main" {
  replication_configuration {
    rule {
      destination {
        region      = "us-west-1"
        registry_id = data.aws_caller_identity.current.account_id
      }
    }
  }
}
