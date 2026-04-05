#!/usr/bin/env bash
#
# init-db.sh — 初始化 PostgreSQL 数据库
# 用于不使用 Docker 的裸机部署场景
#
set -euo pipefail

DB_NAME="${1:-wozai}"
DB_USER="${2:-wozai}"
DB_PASS="${3:-}"

if [[ -z "$DB_PASS" ]]; then
  DB_PASS=$(head -c 24 /dev/urandom | base64 | tr -d '\n/+=' | head -c 32)
  echo "自动生成数据库密码: $DB_PASS"
fi

echo "=== 初始化 PostgreSQL 数据库 ==="
echo "数据库: $DB_NAME"
echo "用户:   $DB_USER"
echo ""

sudo -u postgres psql <<SQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = '${DB_USER}') THEN
    CREATE ROLE ${DB_USER} WITH LOGIN PASSWORD '${DB_PASS}';
  ELSE
    ALTER ROLE ${DB_USER} WITH PASSWORD '${DB_PASS}';
  END IF;
END
\$\$;

SELECT 'CREATE DATABASE ${DB_NAME} OWNER ${DB_USER}'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '${DB_NAME}');
\gexec

GRANT ALL PRIVILEGES ON DATABASE ${DB_NAME} TO ${DB_USER};
SQL

echo ""
echo "数据库初始化完成。"
echo "连接字符串: postgres://${DB_USER}:${DB_PASS}@localhost:5432/${DB_NAME}?sslmode=disable"
echo ""
echo "请将此连接字符串写入 /opt/wozai/.env 的 DATABASE_URL 中"
