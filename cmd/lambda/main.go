// ABOUTME: This file is the main entrypoint for the josh.bot Lambda function.
// ABOUTME: It wires up the domain service with the Lambda adapter and starts the handler.
package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	lambdaadapter "github.com/jduncan/josh-bot/internal/adapters/lambda"
	"github.com/jduncan/josh-bot/internal/adapters/mock"
)

func main() {
	service := mock.NewBotService()
	adapter := lambdaadapter.NewAdapter(service)
	lambda.Start(adapter.Router)
}
