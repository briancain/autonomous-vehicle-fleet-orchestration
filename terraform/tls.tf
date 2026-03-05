# Self-signed TLS certificate for demo purposes (no custom domain required)
# This allows HTTPS on the ALB without a registered domain.
# Browsers will show a certificate warning, but security automation
# won't kill the listener since it's serving HTTPS.

resource "tls_private_key" "alb" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_self_signed_cert" "alb" {
  private_key_pem = tls_private_key.alb.private_key_pem

  subject {
    common_name  = "fleet-orchestration-demo"
    organization = "Demo"
  }

  validity_period_hours = 8760 # 1 year

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_acm_certificate" "alb" {
  private_key      = tls_private_key.alb.private_key_pem
  certificate_body = tls_self_signed_cert.alb.cert_pem
}

# Secondary region cert
resource "tls_private_key" "alb_secondary" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_self_signed_cert" "alb_secondary" {
  private_key_pem = tls_private_key.alb_secondary.private_key_pem

  subject {
    common_name  = "fleet-orchestration-demo-west1"
    organization = "Demo"
  }

  validity_period_hours = 8760

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_acm_certificate" "alb_secondary" {
  provider         = aws.west1
  private_key      = tls_private_key.alb_secondary.private_key_pem
  certificate_body = tls_self_signed_cert.alb_secondary.cert_pem
}
