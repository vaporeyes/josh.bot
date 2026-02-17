#!/bin/bash
# ABOUTME: Wrapper script for sync-mem cron job, tracks last sync epoch for incremental syncs.
# ABOUTME: Called by launchd every 15 minutes; logs to ~/Library/Logs/sync-mem.log.

set -euo pipefail

PROJECT_DIR="$HOME/dev/projects/josh.bot"
STATE_FILE="$HOME/.sync-mem-last-epoch"
LOG_FILE="$HOME/Library/Logs/sync-mem.log"

log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >> "$LOG_FILE"
}

# Read last epoch (default 0 for first run)
SINCE=0
if [ -f "$STATE_FILE" ]; then
  SINCE=$(cat "$STATE_FILE")
fi

log "Starting sync-mem (since epoch $SINCE)"

cd "$PROJECT_DIR"

# Run the sync and capture both stdout and stderr
OUTPUT=$(go run cmd/sync-mem/main.go --since "$SINCE" 2>&1) || {
  log "ERROR: sync-mem failed: $OUTPUT"
  exit 1
}

log "$OUTPUT"

# Extract max_epoch from stderr output (format: max_epoch=NNNN)
NEW_EPOCH=$(echo "$OUTPUT" | grep -o 'max_epoch=[0-9]*' | cut -d= -f2)
if [ -n "$NEW_EPOCH" ] && [ "$NEW_EPOCH" -gt "$SINCE" ]; then
  echo "$NEW_EPOCH" > "$STATE_FILE"
  log "Updated last epoch to $NEW_EPOCH"
fi

log "Sync complete"
