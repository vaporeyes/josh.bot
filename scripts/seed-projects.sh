#!/usr/bin/env bash
# ABOUTME: Seeds the josh-bot-data DynamoDB table with initial project data.
# ABOUTME: Run this once after creating the table: ./scripts/seed-projects.sh

set -euo pipefail

TABLE_NAME="${1:-josh-bot-data}"

aws dynamodb put-item \
	--table-name "$TABLE_NAME" \
	--item '{
    "id": {"S": "project#modular-aws-backend"},
    "slug": {"S": "modular-aws-backend"},
    "name": {"S": "Modular AWS Backend"},
    "stack": {"S": "Go, AWS Lambda, DynamoDB, Terraform"},
    "description": {"S": "API-first personal bot with hexagonal architecture and DynamoDB persistence."},
    "url": {"S": "https://github.com/vaporeyes/josh-bot"},
    "status": {"S": "active"},
    "updated_at": {"S": "2026-02-14T00:00:00Z"}
  }'

echo "Seeded project: modular-aws-backend"

aws dynamodb put-item \
	--table-name "$TABLE_NAME" \
	--item '{
    "id": {"S": "project#modernist-cookbot"},
    "slug": {"S": "modernist-cookbot"},
    "name": {"S": "Modernist Cookbot"},
    "stack": {"S": "Python, Anthropic Claude, Sous Vide"},
    "description": {"S": "AI sous-chef for modernist cooking and sous vide techniques."},
    "url": {"S": "https://github.com/vaporeyes/cookbot"},
    "status": {"S": "active"},
    "updated_at": {"S": "2026-02-14T00:00:00Z"}
  }'

echo "Seeded project: modernist-cookbot"
echo "Done. Seeded 2 projects in $TABLE_NAME"
