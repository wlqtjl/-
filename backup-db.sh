#!/usr/bin/env bash
#
# backup-db.sh — 备份 PostgreSQL 数据库
# 使用方式: bash deploy/backup-db.sh
#
set -euo pipefail

DB_NAME="${DB_NAME:-wozai}"
DB_USER="${DB_USER:-wozai}"
BACKUP_DIR="${BACKUP_DIR:-/opt/wozai/backups}"
KEEP_DAYS="${KEEP_DAYS:-30}"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.sql.gz"

mkdir -p "$BACKUP_DIR"

echo "备份数据库 ${DB_NAME}..."
pg_dump -U "$DB_USER" "$DB_NAME" | gzip > "$BACKUP_FILE"
chmod 600 "$BACKUP_FILE"

echo "备份完成: $BACKUP_FILE ($(du -h "$BACKUP_FILE" | cut -f1))"

# 清理旧备份
find "$BACKUP_DIR" -name "${DB_NAME}_*.sql.gz" -mtime +"$KEEP_DAYS" -delete
echo "已清理 ${KEEP_DAYS} 天前的旧备份"
