# Cognito User Pool for ALB authentication
# Requires authentication on all public ALB endpoints

resource "random_password" "admin" {
  length  = 16
  special = true
}

resource "aws_cognito_user_pool" "main" {
  name = "${var.project_name}-users"

  password_policy {
    minimum_length    = 8
    require_lowercase = true
    require_numbers   = true
    require_symbols   = true
    require_uppercase = true
  }

  admin_create_user_config {
    allow_admin_create_user_only = true
  }
}

resource "aws_cognito_user_pool_client" "alb" {
  name                                 = "${var.project_name}-alb-client"
  user_pool_id                         = aws_cognito_user_pool.main.id
  generate_secret                      = true
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_flows                  = ["code"]
  allowed_oauth_scopes                 = ["openid"]
  callback_urls                        = ["https://${aws_lb.main.dns_name}/oauth2/idpresponse"]
  supported_identity_providers         = ["COGNITO"]
}

resource "aws_cognito_user_pool_domain" "main" {
  domain       = "${var.project_name}-${data.aws_caller_identity.current.account_id}"
  user_pool_id = aws_cognito_user_pool.main.id
}

resource "aws_cognito_user" "admin" {
  user_pool_id = aws_cognito_user_pool.main.id
  username     = "admin"
  password     = random_password.admin.result

  attributes = {
    email          = "admin@example.com"
    email_verified = true
  }
}

# Secondary region Cognito (us-west-1)

resource "aws_cognito_user_pool" "secondary" {
  provider = aws.west1
  name     = "${var.project_name}-users-west1"

  password_policy {
    minimum_length    = 8
    require_lowercase = true
    require_numbers   = true
    require_symbols   = true
    require_uppercase = true
  }

  admin_create_user_config {
    allow_admin_create_user_only = true
  }
}

resource "aws_cognito_user_pool_client" "alb_secondary" {
  provider                             = aws.west1
  name                                 = "${var.project_name}-alb-client-west1"
  user_pool_id                         = aws_cognito_user_pool.secondary.id
  generate_secret                      = true
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_flows                  = ["code"]
  allowed_oauth_scopes                 = ["openid"]
  callback_urls                        = ["https://${aws_lb.secondary.dns_name}/oauth2/idpresponse"]
  supported_identity_providers         = ["COGNITO"]
}

resource "aws_cognito_user_pool_domain" "secondary" {
  provider     = aws.west1
  domain       = "${var.project_name}-${data.aws_caller_identity.current.account_id}-west1"
  user_pool_id = aws_cognito_user_pool.secondary.id
}

resource "aws_cognito_user" "admin_secondary" {
  provider     = aws.west1
  user_pool_id = aws_cognito_user_pool.secondary.id
  username     = "admin"
  password     = random_password.admin.result

  attributes = {
    email          = "admin@example.com"
    email_verified = true
  }
}
