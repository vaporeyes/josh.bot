# ABOUTME: Manages the josh.bot domain - Route53 hosted zone, ACM certificate, and API Gateway custom domain.
# ABOUTME: After apply, update nameservers at your registrar with `terraform output nameservers`.

# 1. Route53 Hosted Zone for josh.bot
resource "aws_route53_zone" "josh_bot" {
  name = "josh.bot"
}

# 2. ACM Certificate for api.josh.bot
resource "aws_acm_certificate" "api" {
  domain_name       = "api.josh.bot"
  validation_method = "DNS"

  lifecycle {
    create_before_destroy = true
  }
}

# 3. DNS validation record for ACM
resource "aws_route53_record" "acm_validation" {
  for_each = { for dvo in aws_acm_certificate.api.domain_validation_options : dvo.domain_name => dvo }

  zone_id = aws_route53_zone.josh_bot.zone_id
  name    = each.value.resource_record_name
  type    = each.value.resource_record_type
  records = [each.value.resource_record_value]
  ttl     = 300
}

resource "aws_acm_certificate_validation" "api" {
  certificate_arn         = aws_acm_certificate.api.arn
  validation_record_fqdns = [for r in aws_route53_record.acm_validation : r.fqdn]
}

# 4. API Gateway Custom Domain
resource "aws_apigatewayv2_domain_name" "api" {
  domain_name = "api.josh.bot"

  domain_name_configuration {
    certificate_arn = aws_acm_certificate_validation.api.certificate_arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
}

resource "aws_apigatewayv2_api_mapping" "api" {
  api_id      = aws_apigatewayv2_api.josh_bot_gw.id
  domain_name = aws_apigatewayv2_domain_name.api.domain_name
  stage       = aws_apigatewayv2_stage.default.id
}

# 5. Route53 A record pointing api.josh.bot to API Gateway
resource "aws_route53_record" "api" {
  zone_id = aws_route53_zone.josh_bot.zone_id
  name    = "api.josh.bot"
  type    = "A"

  alias {
    name                   = aws_apigatewayv2_domain_name.api.domain_name_configuration[0].target_domain_name
    zone_id                = aws_apigatewayv2_domain_name.api.domain_name_configuration[0].hosted_zone_id
    evaluate_target_health = false
  }
}

# 6. Outputs
output "nameservers" {
  description = "Update these NS records at your domain registrar for josh.bot"
  value       = aws_route53_zone.josh_bot.name_servers
}

output "api_url" {
  value = "https://api.josh.bot"
}
