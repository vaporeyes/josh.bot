#!/usr/bin/env bash
# ABOUTME: Seeds the josh-bot-data DynamoDB table with initial bookmark data.
# ABOUTME: Run this once after creating the table: ./scripts/seed-links.sh

set -euo pipefail

TABLE_NAME="${1:-josh-bot-data}"

aws dynamodb put-item \
	--table-name "$TABLE_NAME" \
	--item '{
    "id": {"S": "link#d1e2f3a4b5c6"},
    "url": {"S": "https://go.dev/blog/"},
    "title": {"S": "The Go Blog"},
    "tags": {"L": [{"S": "go"}, {"S": "programming"}]},
    "created_at": {"S": "2026-02-14T00:00:00Z"},
    "updated_at": {"S": "2026-02-14T00:00:00Z"}
  }'

echo "Seeded link: The Go Blog"

aws dynamodb put-item \
	--table-name "$TABLE_NAME" \
	--item '{
    "id": {"S": "link#a2b3c4d5e6f7"},
    "url": {"S": "https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/bp-general-nosql-design.html"},
    "title": {"S": "DynamoDB Single-Table Design Best Practices"},
    "tags": {"L": [{"S": "aws"}, {"S": "dynamodb"}, {"S": "databases"}]},
    "created_at": {"S": "2026-02-14T00:00:00Z"},
    "updated_at": {"S": "2026-02-14T00:00:00Z"}
  }'

echo "Seeded link: DynamoDB Single-Table Design"

aws dynamodb put-item \
	--table-name "$TABLE_NAME" \
	--item '{
    "id": {"S": "link#b3c4d5e6f7a8"},
    "url": {"S": "https://www.alexedwards.net/blog/organising-database-access"},
    "title": {"S": "Organising Database Access in Go"},
    "tags": {"L": [{"S": "go"}, {"S": "databases"}, {"S": "architecture"}]},
    "created_at": {"S": "2026-02-14T00:00:00Z"},
    "updated_at": {"S": "2026-02-14T00:00:00Z"}
  }'

echo "Seeded link: Organising Database Access in Go"
echo "Done. Seeded 3 links in $TABLE_NAME"
