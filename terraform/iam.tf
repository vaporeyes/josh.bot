# ABOUTME: This file defines the IAM roles and policies for the application.
# ABOUTME: It manages permissions for Lambda execution and CI/CD.

resource "aws_iam_role" "lambda_exec" {
  name = "josh-bot-lambda-exec-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_policy" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "lambda_dynamodb" {
  name = "josh-bot-lambda-dynamodb"
  role = aws_iam_role.lambda_exec.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "dynamodb:GetItem",
          "dynamodb:UpdateItem",
          "dynamodb:Scan",
          "dynamodb:Query",
          "dynamodb:PutItem",
          "dynamodb:DeleteItem",
        ]
        Effect = "Allow"
        Resource = [
          aws_dynamodb_table.josh_bot_data.arn,
          "${aws_dynamodb_table.josh_bot_data.arn}/index/*",
          aws_dynamodb_table.josh_bot_lifts.arn,
          aws_dynamodb_table.josh_bot_mem.arn,
          "${aws_dynamodb_table.josh_bot_mem.arn}/index/*",
        ]
      }
    ]
  })
}

# SQS send permission for the API Lambda (publish webhook events to queue)
resource "aws_iam_role_policy" "lambda_sqs_send" {
  name = "josh-bot-lambda-sqs-send"
  role = aws_iam_role.lambda_exec.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["sqs:SendMessage"]
        Effect   = "Allow"
        Resource = [aws_sqs_queue.webhook_queue.arn]
      }
    ]
  })
}

# Webhook processor Lambda IAM role (separate for least privilege)
resource "aws_iam_role" "webhook_processor_exec" {
  name = "josh-bot-webhook-processor-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "webhook_processor_logs" {
  role       = aws_iam_role.webhook_processor_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "webhook_processor_sqs" {
  name = "josh-bot-webhook-processor-sqs"
  role = aws_iam_role.webhook_processor_exec.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
        ]
        Effect   = "Allow"
        Resource = [aws_sqs_queue.webhook_queue.arn]
      }
    ]
  })
}

resource "aws_iam_role_policy" "webhook_processor_dynamodb" {
  name = "josh-bot-webhook-processor-dynamodb"
  role = aws_iam_role.webhook_processor_exec.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["dynamodb:PutItem"]
        Effect   = "Allow"
        Resource = [aws_dynamodb_table.josh_bot_data.arn]
      }
    ]
  })
}

resource "aws_iam_policy" "github_actions" {
  name = "josh-bot-github-actions-policy"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "iam:PassRole"
        ]
        Effect = "Allow"
        Resource = [
          aws_iam_role.lambda_exec.arn,
          aws_iam_role.webhook_processor_exec.arn,
        ]
      },
      {
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:ListBucket"
        ]
        Effect = "Allow"
        Resource = [
          "arn:aws:s3:::jduncanz-terraform-state-bucket",
          "arn:aws:s3:::jduncanz-terraform-state-bucket/*"
        ]
      },
      {
        Action = [
          "lambda:UpdateFunctionCode",
          "lambda:GetFunctionConfiguration"
        ]
        Effect = "Allow"
        Resource = [
          aws_lambda_function.josh_bot_api.arn,
          aws_lambda_function.webhook_processor.arn,
        ]
      }
    ]
  })
}
