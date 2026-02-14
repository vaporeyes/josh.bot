# ABOUTME: Defines the Lambda function, HTTP API Gateway, and API key for josh-bot.
# ABOUTME: API key auth is handled in Go code since HTTP APIs don't support native API keys.

# 1. The Lambda Function
resource "aws_lambda_function" "josh_bot_api" {
  filename      = "function.zip"
  function_name = "josh-bot-api"
  role          = aws_iam_role.lambda_exec.arn
  handler       = "bootstrap" # Required for AL2023 Go runtimes
  runtime       = "provided.al2023"
  architectures = ["arm64"] # Cost-effective and fast

  environment {
    variables = {
      APP_ENV = "production"
      API_KEY = random_password.api_key.result
    }
  }
}

# 2. API Gateway (HTTP API - cheaper and faster than REST API)
resource "aws_apigatewayv2_api" "josh_bot_gw" {
  name          = "josh-bot-gateway"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_integration" "lambda_link" {
  api_id           = aws_apigatewayv2_api.josh_bot_gw.id
  integration_type = "AWS_PROXY"
  integration_uri  = aws_lambda_function.josh_bot_api.invoke_arn
}

# 3. Catch-all Route to let your Go code handle routing
resource "aws_apigatewayv2_route" "default_route" {
  api_id           = aws_apigatewayv2_api.josh_bot_gw.id
  route_key        = "ANY /{proxy+}"
  target = "integrations/${aws_apigatewayv2_integration.lambda_link.id}"
}

# 4. Lambda Permission to allow API Gateway calls
resource "aws_lambda_permission" "api_gw" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.josh_bot_api.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.josh_bot_gw.execution_arn}/*/*"
}

# 5. API Gateway Stage
resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.josh_bot_gw.id
  name        = "$default"
  auto_deploy = true
}

# 6. API Key (generated and stored in SSM, passed to Lambda as env var)
resource "random_password" "api_key" {
  length  = 40
  special = false
}

resource "aws_ssm_parameter" "api_key" {
  name  = "/josh-bot/api-key"
  type  = "SecureString"
  value = random_password.api_key.result
}

output "api_key_value" {
  value     = random_password.api_key.result
  sensitive = true
}