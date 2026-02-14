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
        ]
        Effect   = "Allow"
        Resource = aws_dynamodb_table.josh_bot_data.arn
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
          aws_iam_role.lambda_exec.arn
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
          aws_lambda_function.josh_bot_api.arn
        ]
      }
    ]
  })
}
