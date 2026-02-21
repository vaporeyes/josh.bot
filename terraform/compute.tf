# ABOUTME: Defines the Lambda function, HTTP API Gateway, and API key for josh-bot.
# ABOUTME: API key auth is handled in Go code since HTTP APIs don't support native API keys.

# 1. The Lambda Function
resource "aws_lambda_function" "josh_bot_api" {
  filename         = "function.zip"
  source_code_hash = filebase64sha256("function.zip")
  function_name    = "josh-bot-api"
  role             = aws_iam_role.lambda_exec.arn
  handler          = "bootstrap" # Required for AL2023 Go runtimes
  runtime          = "provided.al2023"
  architectures    = ["arm64"] # Cost-effective and fast

  environment {
    variables = {
      APP_ENV          = "production"
      API_KEY          = random_password.api_key.result
      TABLE_NAME       = aws_dynamodb_table.josh_bot_data.name
      LIFTS_TABLE_NAME = aws_dynamodb_table.josh_bot_lifts.name
      MEM_TABLE_NAME   = aws_dynamodb_table.josh_bot_mem.name
      GITHUB_TOKEN     = var.github_token
      DIARY_REPO_OWNER = var.diary_repo_owner
      DIARY_REPO_NAME  = var.diary_repo_name
      WEBHOOK_SECRET   = var.webhook_secret
    }
  }
}

# 2. API Gateway (HTTP API - cheaper and faster than REST API)
resource "aws_apigatewayv2_api" "josh_bot_gw" {
  name                         = "josh-bot-gateway"
  protocol_type                = "HTTP"
  disable_execute_api_endpoint = true # Force traffic through api.josh.bot
}

resource "aws_apigatewayv2_integration" "lambda_link" {
  api_id                 = aws_apigatewayv2_api.josh_bot_gw.id
  integration_type       = "AWS_PROXY"
  integration_uri        = aws_lambda_function.josh_bot_api.invoke_arn
  payload_format_version = "1.0"
}

# 3. Catch-all Route to let your Go code handle routing
resource "aws_apigatewayv2_route" "default_route" {
  api_id    = aws_apigatewayv2_api.josh_bot_gw.id
  route_key = "ANY /{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_link.id}"
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

  default_route_settings {
    throttling_rate_limit  = 10 # 10 requests per second
    throttling_burst_limit = 20 # 20 concurrent burst
  }
}

# 6. DynamoDB table for bot data (single-table design)
resource "aws_dynamodb_table" "josh_bot_data" {
  name         = "josh-bot-data"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "item_type"
    type = "S"
  }

  attribute {
    name = "created_at"
    type = "S"
  }

  # AIDEV-NOTE: GSI enables Query-based list operations instead of full table Scans.
  global_secondary_index {
    name            = "item-type-index"
    hash_key        = "item_type"
    range_key       = "created_at"
    projection_type = "ALL"
  }

  # AIDEV-NOTE: TTL enables automatic cleanup of idempotency records (idem# prefix, 24h expiry).
  ttl {
    attribute_name = "expires_at"
    enabled        = true
  }
}

# 7. DynamoDB table for lift/workout data (separate from main data table)
resource "aws_dynamodb_table" "josh_bot_lifts" {
  name         = "josh-bot-lifts"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "date"
    type = "S"
  }

  global_secondary_index {
    name            = "date-index"
    hash_key        = "date"
    projection_type = "ALL"
  }
}

# 8. DynamoDB table for claude-mem development context (observations, summaries, prompts)
resource "aws_dynamodb_table" "josh_bot_mem" {
  name         = "josh-bot-mem"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "type"
    type = "S"
  }

  attribute {
    name = "created_at_epoch"
    type = "N"
  }

  global_secondary_index {
    name            = "type-index"
    hash_key        = "type"
    range_key       = "created_at_epoch"
    projection_type = "ALL"
  }
}

# 9. API Key (generated and stored in SSM, passed to Lambda as env var)
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