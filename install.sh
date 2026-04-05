#!/usr/bin/env bash
#
# install.sh — 我在 (WoZai) 一键部署脚本
# 使用方式: sudo bash install.sh
#
set -euo pipefail

APP_NAME="wozai"
APP_USER="wozai"
APP_DIR="/opt/wozai"
BIN_PATH="/usr/local/bin/wozai"
SERVICE_FILE="/etc/systemd/system/wozai.service"
NGINX_CONF="/etc/nginx/sites-available/wozai"
ENV_FILE="${APP_DIR}/.env"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()  { echo -e "${GREEN}[✓]${NC} $*"; }
warn() { echo -e "${YELLOW}[!]${NC} $*"; }
err()  { echo -e "${RED}[✗]${NC} $*"; exit 1; }

# ---------- Pre-flight checks ----------
if [[ $EUID -ne 0 ]]; then
  err "请使用 sudo 运行此脚本"
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ---------- 1. System dependencies ----------
log "安装系统依赖..."
if command -v apt-get &>/dev/null; then
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -qq
  apt-get install -y -qq postgresql postgresql-client nginx curl >/dev/null 2>&1
elif command -v yum &>/dev/null; then
  yum install -y -q postgresql-server postgresql-contrib nginx curl >/dev/null 2>&1
  postgresql-setup --initdb 2>/dev/null || true
else
  warn "未识别的包管理器，请手动安装: postgresql, nginx"
fi

# ---------- 2. Create app user ----------
if ! id "$APP_USER" &>/dev/null; then
  useradd --system --shell /usr/sbin/nologin --home-dir "$APP_DIR" "$APP_USER"
  log "创建系统用户: $APP_USER"
else
  log "系统用户 $APP_USER 已存在"
fi

# ---------- 3. Create directories ----------
mkdir -p "$APP_DIR"
log "应用目录: $APP_DIR"

# ---------- 4. Copy binary ----------
if [[ -f "${SCRIPT_DIR}/wozai" ]]; then
  cp "${SCRIPT_DIR}/wozai" "$BIN_PATH"
  chmod 755 "$BIN_PATH"
  log "已复制二进制文件到 $BIN_PATH"
else
  err "未找到编译好的二进制文件 ${SCRIPT_DIR}/wozai，请先执行 make build"
fi

# ---------- 5. Setup environment file ----------
if [[ ! -f "$ENV_FILE" ]]; then
  JWT_SECRET=$(head -c 48 /dev/urandom | base64 | tr -d '\n/+=' | head -c 64)
  PG_PASSWORD=$(head -c 24 /dev/urandom | base64 | tr -d '\n/+=' | head -c 32)

  cat > "$ENV_FILE" <<ENVEOF
# 我在 (WoZai) - 环境配置
# 由 install.sh 自动生成于 $(date '+%Y-%m-%d %H:%M:%S')

LISTEN_ADDR=:8080
DATABASE_URL=postgres://wozai:${PG_PASSWORD}@localhost:5432/wozai?sslmode=disable
JWT_SECRET=${JWT_SECRET}

# DeepSeek API（必填）
DEEPSEEK_API_KEY=请填入你的DeepSeek_API_Key
DEEPSEEK_URL=https://api.deepseek.com/v1/chat/completions

# SiliconFlow TTS API（必填）
SILICONFLOW_API_KEY=请填入你的SiliconFlow_API_Key
SILICONFLOW_URL=https://api.siliconflow.cn/v1/audio/speech

ACCESS_TOKEN_TTL=15
REFRESH_TOKEN_TTL=168
ENVEOF

  chmod 600 "$ENV_FILE"
  chown "$APP_USER:$APP_USER" "$ENV_FILE"
  log "已生成环境配置: $ENV_FILE"
  warn "请编辑 $ENV_FILE 填入 DEEPSEEK_API_KEY 和 SILICONFLOW_API_KEY"
  echo ""
  echo "  自动生成的 PostgreSQL 密码: ${PG_PASSWORD}"
  echo "  自动生成的 JWT 密钥已写入配置"
  echo ""
else
  log "环境配置已存在: $ENV_FILE"
fi

# ---------- 6. Setup PostgreSQL ----------
log "配置 PostgreSQL..."

# 提取密码用于创建数据库用户
DB_PASS=$(grep '^DATABASE_URL=' "$ENV_FILE" | sed -E 's|.*://wozai:([^@]+)@.*|\1|')

systemctl start postgresql 2>/dev/null || service postgresql start 2>/dev/null || true
systemctl enable postgresql 2>/dev/null || true

# Create DB user and database
su - postgres -c "psql -tc \"SELECT 1 FROM pg_roles WHERE rolname='wozai'\" | grep -q 1 || psql -c \"CREATE ROLE wozai WITH LOGIN PASSWORD '${DB_PASS}';\"" 2>/dev/null || true
su - postgres -c "psql -tc \"SELECT 1 FROM pg_database WHERE datname='wozai'\" | grep -q 1 || psql -c \"CREATE DATABASE wozai OWNER wozai;\"" 2>/dev/null || true
log "PostgreSQL 数据库 wozai 已就绪"

# ---------- 7. Install systemd service ----------
cp "${SCRIPT_DIR}/deploy/wozai.service" "$SERVICE_FILE" 2>/dev/null || \
cat > "$SERVICE_FILE" <<'SVCEOF'
[Unit]
Description=WoZai Digital Soul Platform
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=wozai
Group=wozai
WorkingDirectory=/opt/wozai
EnvironmentFile=/opt/wozai/.env
ExecStart=/usr/local/bin/wozai
Restart=always
RestartSec=5
LimitNOFILE=65535

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/wozai
PrivateTmp=true

[Install]
WantedBy=multi-user.target
SVCEOF

systemctl daemon-reload
systemctl enable "$APP_NAME"
log "systemd 服务已安装: $SERVICE_FILE"

# ---------- 8. Install nginx config ----------
cp "${SCRIPT_DIR}/deploy/nginx-wozai.conf" "$NGINX_CONF" 2>/dev/null || \
cat > "$NGINX_CONF" <<'NGXEOF'
server {
    listen 80;
    server_name _;

    client_max_body_size 2m;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 10s;
        proxy_read_timeout 60s;
        proxy_send_timeout 60s;
    }
}
NGXEOF

# Enable site
if [[ -d /etc/nginx/sites-enabled ]]; then
  ln -sf "$NGINX_CONF" /etc/nginx/sites-enabled/wozai
  rm -f /etc/nginx/sites-enabled/default 2>/dev/null || true
fi
nginx -t 2>/dev/null && systemctl reload nginx 2>/dev/null || true
log "Nginx 反向代理已配置"

# ---------- 9. Set permissions ----------
chown -R "$APP_USER:$APP_USER" "$APP_DIR"
log "目录权限已设置"

# ---------- Done ----------
echo ""
echo "=========================================="
echo "  我在 (WoZai) 安装完成!"
echo "=========================================="
echo ""
echo "  安装路径:   $APP_DIR"
echo "  二进制文件: $BIN_PATH"
echo "  配置文件:   $ENV_FILE"
echo "  服务文件:   $SERVICE_FILE"
echo ""
echo "  下一步操作:"
echo "  1. 编辑配置文件，填入 API 密钥:"
echo "     sudo nano $ENV_FILE"
echo ""
echo "  2. 启动服务:"
echo "     sudo systemctl start wozai"
echo ""
echo "  3. 查看日志:"
echo "     sudo journalctl -u wozai -f"
echo ""
echo "  4. 访问:"
echo "     http://你的服务器IP"
echo ""
