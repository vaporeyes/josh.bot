#!/bin/bash
# ABOUTME: Processes CSV files dropped into ~/media/lifts/ by importing them into DynamoDB.
# ABOUTME: Called by fswatch on file change; moves processed files to ~/media/lifts/done/.

set -euo pipefail

PROJECT_DIR="$HOME/dev/projects/josh.bot"
WATCH_DIR="$HOME/media/lifts"
DONE_DIR="$WATCH_DIR/done"
LOG_FILE="$HOME/Library/Logs/import-lifts.log"

log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >> "$LOG_FILE"
}

mkdir -p "$DONE_DIR"

# Find all CSV files in the watch folder (not in done/)
found=0
for csv in "$WATCH_DIR"/*.csv; do
  [ -f "$csv" ] || continue
  found=1

  filename=$(basename "$csv")
  log "Importing $filename"

  cd "$PROJECT_DIR"
  OUTPUT=$(TABLE_NAME=josh-bot-lifts go run cmd/import-lifts/main.go "$csv" 2>&1) || {
    log "ERROR importing $filename: $OUTPUT"
    continue
  }

  log "$OUTPUT"

  # Move to done folder with timestamp to avoid collisions
  mv "$csv" "$DONE_DIR/${filename%.csv}_$(date '+%Y%m%d%H%M%S').csv"
  log "Moved $filename to done/"
done

if [ "$found" -eq 0 ]; then
  log "Watch triggered but no CSV files found (may have been a subdirectory change)"
fi
