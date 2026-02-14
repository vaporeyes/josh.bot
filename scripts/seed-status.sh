#!/usr/bin/env bash
# ABOUTME: Seeds the josh-bot-data DynamoDB table with initial status data.
# ABOUTME: Run this once after creating the table: ./scripts/seed-status.sh

set -euo pipefail

TABLE_NAME="${1:-josh-bot-data}"

aws dynamodb put-item \
	--table-name "$TABLE_NAME" \
	--item '{
    "id": {"S": "status"},
    "name": {"S": "Josh Duncan"},
    "title": {"S": "Platform Engineer"},
    "bio": {"S": "Builder of systems, lifter of heavy things, maker and whatnot."},
    "current_activity": {"S": "Building josh.bot"},
    "location": {"S": "Clarksville, TN"},
    "availability": {"S": "Open to interesting projects"},
    "status": {"S": "ok"},
    "links": {"M": {
      "github": {"S": "https://github.com/vaporeyes"},
      "linkedin": {"S": "https://www.linkedin.com/in/josh-duncan-919138175/"}
    }},
    "updated_at": {"S": "2026-02-14T00:00:00Z"},
    "interests": {"L": [
	  {"S": "Python"},
      {"S": "Go"},
      {"S": "AWS"},
      {"S": "Sous vide"},
      {"S": "Powerlifting"},
      {"S": "Art Nouveau"},
	  {"S": "Synthwave"},
	  {"S": "Cyberpunk"}
    ]}
  }'

echo "Seeded status item in $TABLE_NAME"
