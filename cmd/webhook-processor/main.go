// ABOUTME: This file is the entrypoint for the webhook processor Lambda function.
// ABOUTME: It reads webhook events from SQS and writes them to DynamoDB.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbadapter "github.com/jduncan/josh-bot/internal/adapters/dynamodb"
	"github.com/jduncan/josh-bot/internal/adapters/sqsprocessor"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		slog.Error("TABLE_NAME environment variable is required")
		os.Exit(1)
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		slog.Error("unable to load AWS config", "error", err)
		os.Exit(1)
	}

	client := dynamodb.NewFromConfig(cfg)
	webhookService := dynamodbadapter.NewWebhookService(client, tableName)
	processor := sqsprocessor.NewProcessor(webhookService)

	lambda.Start(processor.Handle)
}
