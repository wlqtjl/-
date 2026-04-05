#!/usr/bin/env bash
#
# deploy.sh — 我在 (WoZai) 700MB 小内存 VPS 一键部署脚本
#
# 使用方式:
#   curl -sSL https://raw.githubusercontent.com/wlqtjl/-/main/deploy.sh | bash
#   或:
#   git clone https://github.com/wlqtjl/- wozai && cd wozai && bash deploy.sh
#
# 支持:
#   - Docker 容器一键部署 (自动安装 Docker)
#   - 裸机直接部署 (无需 Docker，适合极端低内存)
#   - 自动检测环境和内存
#   - 自动生成安全密钥
#   - 自动健康检查
#
set -euo pipefail

# ===== 颜色 =====
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

log()  { echo -e "${GREEN}[✓]${NC} $*"; }
warn() { echo -e "${YELLOW}[!]${NC} $*"; }
err()  { echo -e "${RED}[✗]${NC} $*"; exit 1; }
info() { echo -e "${BLUE}[i]${NC} $*"; }

# ===== 环境检测 =====
TOTAL_MEM_MB=$(awk '/MemTotal/ {printf "%d", $2/1024}' /proc/meminfo 2>/dev/null || echo "0")
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="${SCRIPT_DIR}"

echo ""
echo -e "${BOLD}╔══════════════════════════════════════╗${NC}"
echo -e "${BOLD}║   我在 (WoZai) · 一键部署脚本       ║${NC}"
echo -e "${BOLD}║   为小内存 VPS 优化                  ║${NC}"
echo -e "${BOLD}╚══════════════════════════════════════╝${NC}"
echo ""
info "检测到系统内存: ${TOTAL_MEM_MB}MB"

if [[ "$TOTAL_MEM_MB" -lt 400 ]]; then
  warn "内存低于 400MB，可能无法稳定运行 Docker 部署"
  warn "将建议使用裸机部署模式"
fi

# ===== 检查是否在项目目录中 =====
check_project_files() {
  if [[ -f "${PROJECT_DIR}/docker-compose.lowmem.yml" ]] && [[ -f "${PROJECT_DIR}/Dockerfile" ]]; then
    return 0
  fi
  return 1
}

# ===== 生成随机密钥 =====
gen_secret() {
  local len="${1:-32}"
  head -c "$len" /dev/urandom | base64 | tr -d '\n/+=' | head -c "$len"
}

# ===== 安装 Docker =====
install_docker() {
  if command -v docker &>/dev/null; then
    log "Docker 已安装: $(docker --version 2>/dev/null | head -1)"
    return 0
  fi

  info "正在安装 Docker..."
  if command -v apt-get &>/dev/null; then
    export DEBIAN_FRONTEND=noninteractive
    apt-get update -qq
    apt-get install -y -qq ca-certificates curl gnupg >/dev/null 2>&1

    install -m 0755 -d /etc/apt/keyrings
    if [[ ! -f /etc/apt/keyrings/docker.gpg ]]; then
      curl -fsSL https://download.docker.com/linux/$(. /etc/os-release && echo "$ID")/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
      chmod a+r /etc/apt/keyrings/docker.gpg
    fi

    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$(. /etc/os-release && echo "$ID") \
      $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
      tee /etc/apt/sources.list.d/docker.list > /dev/null

    apt-get update -qq
    apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-compose-plugin >/dev/null 2>&1
  elif command -v dnf &>/dev/null || command -v yum &>/dev/null; then
    # RHEL 兼容系统：CentOS, Rocky Linux, AlmaLinux, OpenCloudOS, TencentOS 等
    local PKG_MGR="yum"
    if command -v dnf &>/dev/null; then
      PKG_MGR="dnf"
    fi
    $PKG_MGR install -y -q yum-utils >/dev/null 2>&1 || warn "yum-utils 安装失败，将尝试手动添加 repo"
    # 使用 CentOS repo（兼容所有 RHEL 系发行版，包括 OpenCloudOS）
    dnf config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo >/dev/null 2>&1 \
      || yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo >/dev/null 2>&1 \
      || {
        # 手动添加 repo 文件（兼容无 config-manager 的系统）
        cat > /etc/yum.repos.d/docker-ce.repo <<'REPOEOF'
[docker-ce-stable]
name=Docker CE Stable
baseurl=https://download.docker.com/linux/centos/$releasever/$basearch/stable
enabled=1
gpgcheck=1
gpgkey=https://download.docker.com/linux/centos/gpg
REPOEOF
        info "已手动添加 Docker CE repo"
      }
    $PKG_MGR install -y -q docker-ce docker-ce-cli containerd.io docker-compose-plugin >/dev/null 2>&1
  else
    # 通用安装脚本
    curl -fsSL https://get.docker.com | sh
  fi

  systemctl start docker
  systemctl enable docker
  log "Docker 安装完成"
}

# ===== 创建/更新 .env 文件 =====
setup_env() {
  local env_file="${PROJECT_DIR}/.env"

  if [[ -f "$env_file" ]]; then
    warn ".env 已存在，跳过覆盖"
    # 检查是否有未填写的密钥
    if grep -q "请填入\|changeme" "$env_file" 2>/dev/null; then
      warn "请确保 .env 中的 API 密钥已正确填写！"
    fi
    return 0
  fi

  local jwt_secret
  jwt_secret=$(gen_secret 48)
  local pg_password
  pg_password=$(gen_secret 24)

  cat > "$env_file" <<EOF
# 我在 (WoZai) - 环境变量 (由 deploy.sh 自动生成)
# 生成时间: $(date '+%Y-%m-%d %H:%M:%S')

# 数据库密码 (自动生成，请勿随意修改)
POSTGRES_PASSWORD=${pg_password}

# JWT 密钥 (自动生成)
JWT_SECRET=${jwt_secret}

# 服务端口 (Docker 模式下映射到宿主机)
LISTEN_PORT=80

# ===== 以下必须手动填写 =====

# DeepSeek AI 对话 API (https://platform.deepseek.com)
DEEPSEEK_API_KEY=请填入你的DeepSeek_API_Key

# SiliconFlow 语音合成 API (https://cloud.siliconflow.cn)
SILICONFLOW_API_KEY=请填入你的SiliconFlow_API_Key

# ===== 以下为可选配置 =====
AI_PROVIDER=deepseek
ENABLE_SENTIMENT=false
ACCESS_TOKEN_TTL=15
REFRESH_TOKEN_TTL=168
EOF

  chmod 600 "$env_file"
  log "已生成 .env 配置文件"
  echo ""
  echo -e "  ${BOLD}自动生成的密钥:${NC}"
  echo "  PostgreSQL 密码: ${pg_password}"
  echo "  JWT 密钥:        ${jwt_secret:0:16}..."
  echo ""
}

# ===== Docker 部署 =====
deploy_docker() {
  info "使用 Docker 容器部署 (推荐)"
  echo ""

  install_docker
  setup_env

  cd "$PROJECT_DIR"

  # 优化 Docker daemon 配置用于低内存
  if [[ ! -f /etc/docker/daemon.json ]]; then
    mkdir -p /etc/docker
    cat > /etc/docker/daemon.json <<'DJSON'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "5m",
    "max-file": "2"
  },
  "storage-driver": "overlay2"
}
DJSON
    systemctl restart docker 2>/dev/null || true
    log "Docker daemon 已优化 (日志限制 5MB)"
  fi

  info "构建并启动容器..."
  docker compose -f docker-compose.lowmem.yml build --no-cache 2>&1 | tail -5
  docker compose -f docker-compose.lowmem.yml up -d

  log "容器已启动"

  # 等待健康检查
  info "等待服务启动..."
  local retries=0
  while [[ $retries -lt 30 ]]; do
    if docker compose -f docker-compose.lowmem.yml ps --format json 2>/dev/null | grep -q '"healthy"'; then
      break
    fi
    # 兼容旧版 docker compose
    if curl -sf http://localhost:80/health >/dev/null 2>&1 || curl -sf http://localhost:8080/health >/dev/null 2>&1; then
      break
    fi
    sleep 2
    retries=$((retries + 1))
  done

  if curl -sf http://localhost:80/health >/dev/null 2>&1; then
    log "健康检查通过!"
  elif curl -sf http://localhost:8080/health >/dev/null 2>&1; then
    log "健康检查通过! (端口 8080)"
  else
    warn "健康检查未通过，可能需要更长启动时间"
    warn "请运行: docker compose -f docker-compose.lowmem.yml logs -f"
  fi

  # 显示内存使用
  echo ""
  info "当前内存使用:"
  docker stats --no-stream --format "  {{.Name}}: {{.MemUsage}}" 2>/dev/null || true
  echo ""
  free -h | head -2
}

# ===== 裸机部署 =====
deploy_bare() {
  info "使用裸机直接部署 (无 Docker，最省内存)"
  echo ""

  if [[ $EUID -ne 0 ]]; then
    err "裸机部署需要 root 权限，请使用: sudo bash deploy.sh"
  fi

  # 使用已有的 install.sh
  if [[ -f "${PROJECT_DIR}/install.sh" ]]; then
    bash "${PROJECT_DIR}/install.sh"
  else
    err "未找到 install.sh，请确保在项目目录中运行"
  fi
}

# ===== 预构建二进制部署 (无需 Docker 构建) =====
deploy_prebuilt() {
  info "使用预编译二进制 + Docker PostgreSQL (混合模式)"
  echo ""

  install_docker
  setup_env

  cd "$PROJECT_DIR"

  # 只启动 PostgreSQL 容器
  cat > /tmp/wozai-pg.yml <<'PGYML'
services:
  db:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      POSTGRES_USER: wozai
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-changeme}
      POSTGRES_DB: wozai
    command: >
      postgres
        -c shared_buffers=32MB
        -c effective_cache_size=128MB
        -c work_mem=2MB
        -c maintenance_work_mem=16MB
        -c max_connections=20
    volumes:
      - pgdata:/var/lib/postgresql/data
    ports:
      - "127.0.0.1:5432:5432"
    deploy:
      resources:
        limits:
          memory: 128M
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U wozai"]
      interval: 10s
      timeout: 3s
      retries: 3
volumes:
  pgdata:
PGYML

  docker compose -f /tmp/wozai-pg.yml --env-file "${PROJECT_DIR}/.env" up -d
  log "PostgreSQL 容器已启动"

  # 检查是否有预编译二进制
  if [[ -f "${PROJECT_DIR}/wozai" ]]; then
    local bin_path="/usr/local/bin/wozai"
    cp "${PROJECT_DIR}/wozai" "$bin_path"
    chmod 755 "$bin_path"
    log "已安装二进制: $bin_path"

    # 从 .env 读取配置
    local pg_pass
    pg_pass=$(grep '^POSTGRES_PASSWORD=' "${PROJECT_DIR}/.env" | cut -d= -f2)
    local jwt_secret
    jwt_secret=$(grep '^JWT_SECRET=' "${PROJECT_DIR}/.env" | cut -d= -f2)

    # 创建 systemd service
    cat > /etc/systemd/system/wozai.service <<SVCEOF
[Unit]
Description=WoZai Digital Soul Platform
After=network.target docker.service

[Service]
Type=simple
EnvironmentFile=${PROJECT_DIR}/.env
Environment=DATABASE_URL=postgres://wozai:${pg_pass}@127.0.0.1:5432/wozai?sslmode=disable
Environment=LISTEN_ADDR=:8080
ExecStart=/usr/local/bin/wozai
Restart=always
RestartSec=5
LimitNOFILE=65535
MemoryMax=64M

[Install]
WantedBy=multi-user.target
SVCEOF

    systemctl daemon-reload
    systemctl enable wozai
    systemctl start wozai
    log "WoZai 服务已启动"
  else
    warn "未找到预编译二进制文件，需要手动编译"
    warn "  make build 或 go build -ldflags=\"-s -w\" -o wozai ./cmd/wozai"
  fi
}

# ===== 状态检查 =====
check_status() {
  echo ""
  echo -e "${BOLD}=== 部署状态 ===${NC}"
  echo ""

  # 检查 Docker 容器
  if command -v docker &>/dev/null; then
    if docker compose -f "${PROJECT_DIR}/docker-compose.lowmem.yml" ps --format "table {{.Name}}\t{{.Status}}" 2>/dev/null | grep -q "wozai"; then
      info "Docker 容器:"
      docker compose -f "${PROJECT_DIR}/docker-compose.lowmem.yml" ps --format "table {{.Name}}\t{{.Status}}" 2>/dev/null || docker compose -f "${PROJECT_DIR}/docker-compose.lowmem.yml" ps 2>/dev/null
    fi
  fi

  # 检查 systemd 服务
  if systemctl is-active wozai &>/dev/null; then
    info "systemd 服务: $(systemctl is-active wozai)"
  fi

  # 检查端口
  if command -v ss &>/dev/null; then
    info "监听端口:"
    ss -tlnp 2>/dev/null | grep -E ':(80|8080|5432)\s' | sed 's/^/  /'
  fi

  # 健康检查
  if curl -sf http://localhost:80/health >/dev/null 2>&1; then
    log "健康检查: http://localhost:80 ✓"
  elif curl -sf http://localhost:8080/health >/dev/null 2>&1; then
    log "健康检查: http://localhost:8080 ✓"
  else
    warn "健康检查: 服务未响应"
  fi

  # 内存使用
  echo ""
  info "系统内存:"
  free -h | head -2 | sed 's/^/  /'
  echo ""
}

# ===== 卸载 =====
uninstall() {
  echo ""
  warn "即将卸载 WoZai..."
  read -rp "确定要卸载吗? (y/N) " confirm
  if [[ "${confirm}" != "y" && "${confirm}" != "Y" ]]; then
    echo "已取消"
    exit 0
  fi

  # 停止 Docker 容器
  if command -v docker &>/dev/null; then
    docker compose -f "${PROJECT_DIR}/docker-compose.lowmem.yml" down 2>/dev/null || true
    docker compose -f /tmp/wozai-pg.yml down 2>/dev/null || true
  fi

  # 停止 systemd 服务
  systemctl stop wozai 2>/dev/null || true
  systemctl disable wozai 2>/dev/null || true
  rm -f /etc/systemd/system/wozai.service
  systemctl daemon-reload 2>/dev/null || true

  rm -f /usr/local/bin/wozai

  log "卸载完成 (数据保留在 Docker volume 中)"
  warn "如需删除数据: docker volume rm wozai_pgdata 或 类似名称"
}

# ===== 主菜单 =====
show_menu() {
  echo ""
  echo -e "${BOLD}请选择部署方式:${NC}"
  echo ""
  echo "  1) Docker 容器部署 (推荐，自动构建，占用 ~200MB)"
  echo "  2) 裸机部署 (无 Docker，最省内存，占用 ~120MB)"
  echo "  3) 混合部署 (预编译二进制 + Docker 数据库，占用 ~140MB)"
  echo "  4) 查看部署状态"
  echo "  5) 卸载"
  echo "  0) 退出"
  echo ""
  read -rp "选择 [1-5, 默认 1]: " choice
  choice="${choice:-1}"

  case "$choice" in
    1) deploy_docker ;;
    2) deploy_bare ;;
    3) deploy_prebuilt ;;
    4) check_status ;;
    5) uninstall ;;
    0) exit 0 ;;
    *) err "无效选择" ;;
  esac
}

# ===== 主入口 =====
# 支持直接传参: bash deploy.sh docker / bare / prebuilt / status / uninstall
case "${1:-}" in
  docker)    deploy_docker ;;
  bare)      deploy_bare ;;
  prebuilt)  deploy_prebuilt ;;
  status)    check_status ;;
  uninstall) uninstall ;;
  *)
    if ! check_project_files; then
      err "未检测到项目文件，请在项目目录中运行此脚本"
    fi
    show_menu
    ;;
esac

# ===== 完成提示 =====
echo ""
echo -e "${BOLD}╔══════════════════════════════════════╗${NC}"
echo -e "${BOLD}║   部署完成!                          ║${NC}"
echo -e "${BOLD}╚══════════════════════════════════════╝${NC}"
echo ""
echo "  下一步操作:"
echo ""
echo "  1. 编辑 .env 填入 API 密钥:"
echo "     nano ${PROJECT_DIR}/.env"
echo ""
echo "  2. 重启服务使配置生效:"
if command -v docker &>/dev/null && docker compose -f "${PROJECT_DIR}/docker-compose.lowmem.yml" ps &>/dev/null 2>&1; then
  echo "     docker compose -f docker-compose.lowmem.yml restart"
else
  echo "     sudo systemctl restart wozai"
fi
echo ""
echo "  3. 访问: http://$(hostname -I 2>/dev/null | awk '{print $1}' || echo '你的服务器IP')"
echo ""
echo "  查看日志:"
if command -v docker &>/dev/null && docker compose -f "${PROJECT_DIR}/docker-compose.lowmem.yml" ps &>/dev/null 2>&1; then
  echo "     docker compose -f docker-compose.lowmem.yml logs -f"
else
  echo "     sudo journalctl -u wozai -f"
fi
echo ""

check_status
