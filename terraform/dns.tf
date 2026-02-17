# ABOUTME: ACM certificate and API Gateway custom domain for api.josh.bot.
# ABOUTME: DNS validation and CNAME records are managed in Cloudflare, not Terraform.

# 1. ACM Certificate for api.josh.bot
resource "aws_acm_certificate" "api" {
  domain_name       = "api.josh.bot"
  validation_method = "DNS"

  lifecycle {
    create_before_destroy = true
  }
}

# 2. API Gateway Custom Domain
resource "aws_apigatewayv2_domain_name" "api" {
  domain_name = "api.josh.bot"

  domain_name_configuration {
    certificate_arn = aws_acm_certificate.api.arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
}

# 3. Map the custom domain to the API Gateway stage
resource "aws_apigatewayv2_api_mapping" "api" {
  api_id      = aws_apigatewayv2_api.josh_bot_gw.id
  domain_name = aws_apigatewayv2_domain_name.api.domain_name
  stage       = aws_apigatewayv2_stage.default.id
}

# 4. Outputs for Cloudflare DNS setup
output "api_gateway_target_domain" {
  description = "CNAME target for api.josh.bot in Cloudflare (DNS only, not proxied)"
  value       = aws_apigatewayv2_domain_name.api.domain_name_configuration[0].target_domain_name
}

output "acm_validation_records" {
  description = "Add these CNAME records in Cloudflare to validate the ACM certificate"
  value = {
    for dvo in aws_acm_certificate.api.domain_validation_options : dvo.domain_name => {
      name  = dvo.resource_record_name
      type  = dvo.resource_record_type
      value = dvo.resource_record_value
    }
  }
}
