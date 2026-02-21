# ABOUTME: Defines the SQS queue and DLQ for async webhook event processing.
# ABOUTME: Events are published by the API Lambda and consumed by the webhook processor Lambda.

# 1. Dead Letter Queue for failed webhook processing
resource "aws_sqs_queue" "webhook_dlq" {
  name                      = "josh-bot-webhook-dlq"
  message_retention_seconds = 1209600 # 14 days
}

# 2. Main webhook processing queue
resource "aws_sqs_queue" "webhook_queue" {
  name                       = "josh-bot-webhook-queue"
  visibility_timeout_seconds = 30

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.webhook_dlq.arn
    maxReceiveCount     = 3
  })
}

# 3. SQS event source mapping to trigger the processor Lambda
resource "aws_lambda_event_source_mapping" "webhook_sqs" {
  event_source_arn                   = aws_sqs_queue.webhook_queue.arn
  function_name                      = aws_lambda_function.webhook_processor.arn
  batch_size                         = 10
  function_response_types            = ["ReportBatchItemFailures"]
  maximum_batching_window_in_seconds = 5
}
