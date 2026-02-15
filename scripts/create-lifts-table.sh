#!/usr/bin/env bash
# ABOUTME: Creates the josh-bot-lifts DynamoDB table for workout data.
# ABOUTME: Run once: ./scripts/create-lifts-table.sh

set -euo pipefail

TABLE_NAME="${1:-josh-bot-lifts}"

aws dynamodb create-table \
	--table-name "$TABLE_NAME" \
	--billing-mode PAY_PER_REQUEST \
	--attribute-definitions \
		AttributeName=id,AttributeType=S \
		AttributeName=date,AttributeType=S \
	--key-schema \
		AttributeName=id,KeyType=HASH \
	--global-secondary-indexes \
		'[{
			"IndexName": "date-index",
			"KeySchema": [
				{"AttributeName": "date", "KeyType": "HASH"}
			],
			"Projection": {"ProjectionType": "ALL"}
		}]'

echo "Created table: $TABLE_NAME"
echo "  Partition key: id (S)"
echo "  GSI: date-index (date S) - for time-range queries (tonnage, recent workouts)"
