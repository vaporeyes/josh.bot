#!/usr/bin/env bash
# ABOUTME: Zips the ArchiveBox data directory and rsyncs it to an NFS backup path.
# ABOUTME: Intended to run as a daily cron job: 0 2 * * * /path/to/backup-archivebox.sh

set -euo pipefail

# --- Configuration ---
ARCHIVEBOX_DIR="/home/jduncan/media-stack/archivebox/data/"
NFS_BACKUP_DIR="/mnt/8tbdata/backups/archivebox/"
KEEP_DAYS=7 # Number of daily backups to retain
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
GZ_NAME="archivebox-${TIMESTAMP}.tar.gz"
TMP_DIR=$(mktemp -d)

trap 'rm -rf "$TMP_DIR"' EXIT

# --- Preflight checks ---
if [ ! -d "$ARCHIVEBOX_DIR" ]; then
	echo "ERROR: ArchiveBox directory not found: $ARCHIVEBOX_DIR" >&2
	exit 1
fi

if ! mountpoint -q "$NFS_BACKUP_DIR" 2>/dev/null; then
	echo "WARNING: $NFS_BACKUP_DIR may not be a mount point, proceeding anyway" >&2
fi

if [ ! -d "$NFS_BACKUP_DIR" ]; then
	echo "ERROR: NFS backup directory not found: $NFS_BACKUP_DIR" >&2
	exit 1
fi

# --- Create tar.gz ---
echo "tar gzipping $ARCHIVEBOX_DIR -> $TMP_DIR/$GZ_NAME"
cd "$ARCHIVEBOX_DIR"
tar -czf "$TMP_DIR/$GZ_NAME" \
	--exclude="./.tmp/*" \
	--exclude="./logs/*" \
	.

FILE_SIZE=$(du -h "$TMP_DIR/$GZ_NAME" | cut -f1)
echo "Created $GZ_NAME ($FILE_SIZE)"

# --- Rsync to NFS ---
echo "Rsyncing to $NFS_BACKUP_DIR/"
rsync -ah --progress "$TMP_DIR/$GZ_NAME" "$NFS_BACKUP_DIR/"
echo "Backup complete: $NFS_BACKUP_DIR/$GZ_NAME"

# --- Prune old backups ---
echo "Pruning backups older than $KEEP_DAYS days"
find "$NFS_BACKUP_DIR" -name "archivebox-*.tar.gz" -mtime +"$KEEP_DAYS" -delete -print | while read -r f; do
	echo "  Removed: $f"
done

echo "Done."
