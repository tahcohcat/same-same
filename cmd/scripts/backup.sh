#!/bin/bash

# Backup script for Same-Same Vector Database
set -e

BACKUP_DIR="./backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="same-same-backup_$TIMESTAMP.json"

echo "💾 Creating backup of Same-Same Vector Database..."

# Create backup directory
mkdir -p $BACKUP_DIR

# Export all vectors
echo "📥 Exporting vectors..."
curl -s http://localhost:8080/api/v1/vectors > "$BACKUP_DIR/$BACKUP_FILE"

if [ $? -eq 0 ]; then
    echo "✅ Backup created successfully: $BACKUP_DIR/$BACKUP_FILE"
    
    # Get vector count for verification
    VECTOR_COUNT=$(curl -s http://localhost:8080/api/v1/vectors/count | jq -r '.count' 2>/dev/null || echo "unknown")
    echo "📊 Backed up $VECTOR_COUNT vectors"
    
    # Compress backup
    gzip "$BACKUP_DIR/$BACKUP_FILE"
    echo "🗜️ Backup compressed: $BACKUP_DIR/$BACKUP_FILE.gz"
    
    # Clean up old backups (keep last 10)
    ls -t $BACKUP_DIR/same-same-backup_*.json.gz | tail -n +11 | xargs -r rm
    echo "🧹 Old backups cleaned up"
else
    echo "❌ Backup failed"
    rm -f "$BACKUP_DIR/$BACKUP_FILE"
    exit 1
fi