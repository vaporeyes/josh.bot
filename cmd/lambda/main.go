// ABOUTME: This file is the main entrypoint for the josh.bot Lambda function.
// ABOUTME: It wires up the DynamoDB-backed service with the Lambda adapter.
package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbadapter "github.com/jduncan/josh-bot/internal/adapters/dynamodb"
	ghclient "github.com/jduncan/josh-bot/internal/adapters/github"
	lambdaadapter "github.com/jduncan/josh-bot/internal/adapters/lambda"
	diarysvc "github.com/jduncan/josh-bot/internal/service"
)

func main() {
	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		log.Fatal("TABLE_NAME environment variable is required")
	}

	liftsTableName := os.Getenv("LIFTS_TABLE_NAME")
	if liftsTableName == "" {
		log.Fatal("LIFTS_TABLE_NAME environment variable is required")
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}

	memTableName := os.Getenv("MEM_TABLE_NAME")
	if memTableName == "" {
		memTableName = "josh-bot-mem"
	}

	client := dynamodb.NewFromConfig(cfg)
	service := dynamodbadapter.NewBotService(client, tableName)
	memService := dynamodbadapter.NewMemService(client, memTableName)
	metricsService := dynamodbadapter.NewMetricsService(client, liftsTableName, tableName, memService)
	adapter := lambdaadapter.NewAdapter(service, metricsService, memService)

	// Wire up webhook service if secret is configured
	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	webhookService := dynamodbadapter.NewWebhookService(client, tableName)
	adapter.SetWebhookService(webhookService, webhookSecret)

	// Wire up diary service with GitHub publishing if configured
	ghToken := os.Getenv("GITHUB_TOKEN")
	ghOwner := os.Getenv("DIARY_REPO_OWNER")
	ghRepo := os.Getenv("DIARY_REPO_NAME")
	if ghToken != "" && ghOwner != "" && ghRepo != "" {
		publisher := ghclient.NewClient(ghToken, ghOwner, ghRepo)
		diarySvc := diarysvc.NewDiaryService(service, publisher)
		adapter.SetDiaryService(diarySvc)
	}

	lambda.Start(adapter.Router)
}
