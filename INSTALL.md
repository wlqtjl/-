# 我在 (WoZai) — 安装部署文档

> **想部署到公网？** 请先阅读 [DEPLOY.md](DEPLOY.md)（云服务器选型 + 完整上线流程）。

## 目录

1. [🔥 700MB 小内存 VPS 一键部署](#0-700mb-小内存-vps-一键部署)
2. [系统要求](#1-系统要求)
3. [项目文件结构](#2-项目文件结构)
4. [方式一：一键安装（推荐）](#3-方式一一键安装推荐)
5. [方式二：Docker Compose 部署](#4-方式二docker-compose-部署)
6. [方式三：手动安装](#5-方式三手动安装)
7. [外部依赖 API 申请指南](#6-外部依赖-api-申请指南)
8. [配置说明](#7-配置说明)
9. [HTTPS / 域名配置](#8-https--域名配置)
10. [运维操作](#9-运维操作)
11. [数据备份与恢复](#10-数据备份与恢复)
12. [故障排查](#11-故障排查)
13. [架构说明](#12-架构说明)

---

## 0. 700MB 小内存 VPS 一键部署

> 专为 **700MB 内存 VPS** 优化，整体运行时内存 < 200MB。

### 内存预算

| 组件             | 内存占用    | 说明                        |
|-----------------|------------|----------------------------|
| 系统 + Docker    | ~120MB     | Alpine Linux 容器极低开销    |
| PostgreSQL       | ~80MB      | shared_buffers=32MB 调优    |
| WoZai Go 应用    | ~30MB      | 静态二进制 + GOMEMLIMIT 限制 |
| **合计**         | **~230MB** | 剩余 470MB 给系统缓存        |

### 一键部署 (3 步)

```bash
# 1. 获取代码
git clone https://github.com/wlqtjl/- wozai
cd wozai

# 2. 一键部署 (自动安装 Docker + 构建 + 启动)
sudo bash deploy.sh

# 3. 编辑配置填入 API 密钥
nano .env
# 修改 DEEPSEEK_API_KEY 和 SILICONFLOW_API_KEY
# 然后重启:
docker compose -f docker-compose.lowmem.yml restart
```

### 也可以直接命令行指定模式

```bash
# Docker 容器部署 (推荐)
sudo bash deploy.sh docker

# 裸机部署 (无 Docker，最省内存)
sudo bash deploy.sh bare

# 查看状态
bash deploy.sh status
```

### 或使用 Make 命令

```bash
# Docker 一键部署
make deploy

# 查看日志
make deploy-logs

# 查看状态
make deploy-status

# 停止
make deploy-down
```

### 常用运维命令

```bash
# 查看容器状态和内存
docker compose -f docker-compose.lowmem.yml ps
docker stats --no-stream

# 查看日志
docker compose -f docker-compose.lowmem.yml logs -f app

# 重启
docker compose -f docker-compose.lowmem.yml restart

# 停止
docker compose -f docker-compose.lowmem.yml down

# 更新
git pull
docker compose -f docker-compose.lowmem.yml up -d --build

# 备份数据库
docker compose -f docker-compose.lowmem.yml exec db pg_dump -U wozai wozai | gzip > backup_$(date +%Y%m%d).sql.gz
```

### 关键优化点

- **PostgreSQL**: `shared_buffers=32MB`, `work_mem=2MB`, `max_connections=20`
- **Go 应用**: `GOMEMLIMIT=50MiB`, `GOGC=50` (更积极的 GC)
- **Docker**: `memory limit 128MB` (PG) + `64MB` (App), 日志限制 5MB
- **镜像**: Alpine 基础，运行镜像仅 ~15MB
- **网络**: 不暴露 PostgreSQL 端口，仅内部通信
- **健康检查**: 自动重启不健康容器

---

## 1. 系统要求

### 最低配置 (700MB VPS ✓)

| 项目       | 要求                          |
|------------|-------------------------------|
| 操作系统   | Ubuntu 20.04+ / Debian 11+ / CentOS 8+ / Rocky Linux 8+ / OpenCloudOS / TencentOS |
| CPU        | 1 核                          |
| 内存       | **512 MB** (使用 lowmem 配置)  |
| 磁盘       | 5 GB                          |
| 网络       | 可访问 api.deepseek.com 和 api.siliconflow.cn |

### 推荐配置

| 项目       | 要求                          |
|------------|-------------------------------|
| CPU        | 2 核                          |
| 内存       | 2 GB                          |
| 磁盘       | 20 GB SSD                     |

### 软件依赖

#### 方式一（裸机一键安装）
- PostgreSQL 14+
- Nginx（用于反向代理）
- 脚本自动安装以上依赖

#### 方式二（Docker）
- Docker 20.10+
- Docker Compose v2+
- 无需安装 PostgreSQL（容器自带）

#### 方式三（手动源码编译）
- Go 1.22+
- PostgreSQL 14+
- Nginx（可选）

---

## 2. 项目文件结构

```
wozai-master/
├── cmd/wozai/main.go          # 程序入口
├── internal/                   # 核心业务代码
│   ├── config/                 # 配置加载
│   ├── handler/                # HTTP 处理器
│   ├── middleware/             # 中间件（认证、安全头、限流）
│   ├── migrations/sql/         # 数据库迁移文件（自动执行）
│   ├── model/                  # 数据模型
│   ├── repo/                   # 数据访问层
│   ├── router/                 # 路由
│   └── service/                # 业务逻辑
├── web/                        # 前端静态文件（编译嵌入二进制）
│   ├── index.html
│   ├── style.css
│   ├── app.js
│   └── embed.go
├── deploy/                     # 部署套件
│   ├── install.sh              # 一键安装脚本
│   ├── init-db.sh              # 数据库初始化脚本
│   ├── build-release.sh        # 构建发布包脚本
│   ├── backup-db.sh            # 数据库备份脚本
│   ├── wozai.service           # systemd 服务文件
│   └── nginx-wozai.conf        # Nginx 配置文件
├── Dockerfile                  # Docker 构建文件
├── docker-compose.yml          # Docker Compose 编排
├── .env.example                # 环境变量模板
├── Makefile                    # 构建命令
├── go.mod / go.sum             # Go 依赖
└── INSTALL.md                  # 本文档
```

---

## 3. 方式一：一键安装（推荐）

适合 IDC 裸机 Linux 服务器。安装脚本会自动完成：
- 安装 PostgreSQL + Nginx
- 创建系统用户和数据库
- 部署二进制文件
- 配置 systemd 服务
- 配置 Nginx 反向代理
- 生成安全密钥

### 3.1 在本地构建安装包

在你的开发机器（macOS/Linux）执行：

```bash
cd wozai-master

# 安装 Go 依赖
go mod download

# 交叉编译 + 打包（自动生成 dist/wozai-YYYYMMDD-linux-amd64.tar.gz）
bash deploy/build-release.sh
```

如果你的 `go` 不在 PATH 中：
```bash
GO=~/go/bin/go bash deploy/build-release.sh
```

### 3.2 上传到服务器

```bash
scp dist/wozai-*-linux-amd64.tar.gz root@你的服务器IP:/tmp/
```

### 3.3 在服务器执行安装

```bash
ssh root@你的服务器IP

cd /tmp
tar xzf wozai-*-linux-amd64.tar.gz
cd wozai-*-linux-amd64

sudo bash deploy/install.sh
```

### 3.4 填写 API 密钥

```bash
sudo nano /opt/wozai/.env
```

**必须修改以下两项**（参见 [第6节](#6-外部依赖-api-申请指南) 获取密钥）：
```
DEEPSEEK_API_KEY=sk-xxxxxxxxxxxx
SILICONFLOW_API_KEY=sk-xxxxxxxxxxxx
```

### 3.5 启动服务

```bash
sudo systemctl start wozai
sudo systemctl status wozai
```

### 3.6 验证

```bash
# 健康检查
curl http://localhost:8080/health

# 浏览器访问
http://你的服务器IP
```

---

## 4. 方式二：Docker Compose 部署

适合已有 Docker 环境的服务器。

### 4.1 上传项目文件

将整个 `wozai-master` 目录上传到服务器：
```bash
scp -r wozai-master root@你的服务器IP:/opt/wozai-src
```

### 4.2 创建环境配置

```bash
ssh root@你的服务器IP
cd /opt/wozai-src

cp .env.example .env
```

编辑 `.env`，至少修改以下项：
```bash
# 安全：必须更改！
POSTGRES_PASSWORD=你的数据库密码（建议32位随机字符串）
JWT_SECRET=你的JWT密钥（至少32位随机字符串）

# API 密钥：必须填写！
DEEPSEEK_API_KEY=sk-xxxxxxxxxxxx
SILICONFLOW_API_KEY=sk-xxxxxxxxxxxx
```

快速生成随机密码：
```bash
# JWT 密钥
openssl rand -base64 48

# 数据库密码
openssl rand -base64 24
```

### 4.3 启动

```bash
docker compose up -d
```

### 4.4 验证

```bash
# 查看容器状态
docker compose ps

# 查看日志
docker compose logs -f app

# 健康检查
curl http://localhost:8080/health

# 浏览器访问
http://你的服务器IP:8080
```

### 4.5 停止 / 重启

```bash
docker compose stop       # 停止
docker compose restart    # 重启
docker compose down       # 停止并移除容器（数据保留在 volume）
docker compose down -v    # ⚠️ 停止并删除所有数据
```

---

## 5. 方式三：手动安装

### 5.1 安装 Go

```bash
# Ubuntu/Debian
wget https://go.dev/dl/go1.24.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

### 5.2 安装 PostgreSQL

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y postgresql postgresql-client

# 启动
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

### 5.3 初始化数据库

```bash
# 使用提供的脚本
sudo bash deploy/init-db.sh wozai wozai 你的数据库密码

# 或手动执行
sudo -u postgres psql
CREATE ROLE wozai WITH LOGIN PASSWORD 'your_password';
CREATE DATABASE wozai OWNER wozai;
\q
```

### 5.4 编译

```bash
cd wozai-master
go mod download
CGO_ENABLED=0 go build -ldflags="-s -w" -o wozai ./cmd/wozai
```

### 5.5 配置

```bash
sudo mkdir -p /opt/wozai
sudo cp wozai /usr/local/bin/
sudo cp .env.example /opt/wozai/.env
sudo chmod 600 /opt/wozai/.env
sudo nano /opt/wozai/.env   # 填入所有配置
```

### 5.6 安装 systemd 服务

```bash
sudo cp deploy/wozai.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable wozai
sudo systemctl start wozai
```

### 5.7 配置 Nginx（可选）

```bash
sudo apt install -y nginx
sudo cp deploy/nginx-wozai.conf /etc/nginx/sites-available/wozai
sudo ln -sf /etc/nginx/sites-available/wozai /etc/nginx/sites-enabled/wozai
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t && sudo systemctl reload nginx
```

---

## 6. 外部依赖 API 申请指南

本平台依赖两个 AI 服务的 API，**必须申请后才能使用对话和语音功能**。

### 6.1 DeepSeek API（AI 对话）

| 项目     | 说明                                          |
|----------|-----------------------------------------------|
| 官网     | https://platform.deepseek.com                 |
| 用途     | 驱动数字灵魂的对话能力                        |
| 模型     | deepseek-chat                                 |
| 计费     | 按 token 计费，新用户有免费额度                |

**申请步骤：**
1. 访问 https://platform.deepseek.com 注册账号
2. 进入「API Keys」页面
3. 点击「创建 API Key」
4. 复制生成的 `sk-` 开头的密钥
5. 填入配置文件的 `DEEPSEEK_API_KEY`

### 6.2 SiliconFlow API（语音合成）

| 项目     | 说明                                          |
|----------|-----------------------------------------------|
| 官网     | https://siliconflow.cn                        |
| 用途     | 将灵魂回复文字转为语音                        |
| 模型     | FishAudio/fish-speech-1.5                     |
| 计费     | 按字符计费，新用户有免费额度                   |

**申请步骤：**
1. 访问 https://cloud.siliconflow.cn 注册账号
2. 进入「API 密钥」页面
3. 创建新的 API Key
4. 复制生成的 `sk-` 开头的密钥
5. 填入配置文件的 `SILICONFLOW_API_KEY`

### 6.3 费用预估

| 功能     | 单次消耗            | 月均预估（100用户/天） |
|----------|---------------------|------------------------|
| 对话     | ~300 token/次       | ¥30-80                 |
| 语音合成 | ~100 字符/次        | ¥15-40                 |

> 实际费用取决于使用量。两家平台均支持预充值，余额用完即停，无超额风险。

---

## 7. 配置说明

所有配置通过环境变量管理，配置文件位于 `/opt/wozai/.env`（裸机）或项目根目录 `.env`（Docker）。

| 变量名              | 必填 | 默认值                                          | 说明                   |
|---------------------|------|------------------------------------------------|------------------------|
| `LISTEN_ADDR`       | 否   | `:8080`                                        | 服务监听地址            |
| `DATABASE_URL`      | 是   | —                                              | PostgreSQL 连接字符串   |
| `JWT_SECRET`        | 是   | —                                              | JWT 签名密钥（≥32字符） |
| `DEEPSEEK_API_KEY`  | 是   | —                                              | DeepSeek API 密钥       |
| `DEEPSEEK_URL`      | 否   | `https://api.deepseek.com/v1/chat/completions` | DeepSeek API 地址       |
| `SILICONFLOW_API_KEY`| 是  | —                                              | SiliconFlow API 密钥    |
| `SILICONFLOW_URL`   | 否   | `https://api.siliconflow.cn/v1/audio/speech`   | SiliconFlow API 地址    |
| `ACCESS_TOKEN_TTL`  | 否   | `15`                                           | 访问令牌有效期（分钟）   |
| `REFRESH_TOKEN_TTL` | 否   | `168`                                          | 刷新令牌有效期（小时）   |

### DATABASE_URL 格式

```
postgres://用户名:密码@主机:端口/数据库名?sslmode=disable
```

示例：
```
postgres://wozai:mypassword@localhost:5432/wozai?sslmode=disable
```

---

## 8. HTTPS / 域名配置

### 使用 Let's Encrypt 自动签发

```bash
# 安装 certbot
sudo apt install -y certbot python3-certbot-nginx

# 先确保域名 DNS 已解析到服务器 IP，然后：
sudo certbot --nginx -d wozai.example.com
```

Certbot 会自动修改 Nginx 配置并设置自动续期。

### 手动 HTTPS

参考 `deploy/nginx-wozai.conf` 中注释掉的 HTTPS 配置段。

---

## 9. 运维操作

### 服务管理

```bash
# 启动
sudo systemctl start wozai

# 停止
sudo systemctl stop wozai

# 重启
sudo systemctl restart wozai

# 查看状态
sudo systemctl status wozai

# 查看日志（实时跟踪）
sudo journalctl -u wozai -f

# 查看最近100行日志
sudo journalctl -u wozai -n 100

# 开机自启
sudo systemctl enable wozai

# 取消开机自启
sudo systemctl disable wozai
```

### 更新部署

```bash
# 1. 在开发机重新构建
GO=~/go/bin/go bash deploy/build-release.sh

# 2. 上传新的二进制
scp dist/wozai-*-linux-amd64.tar.gz root@服务器:/tmp/

# 3. 在服务器替换
ssh root@服务器
cd /tmp && tar xzf wozai-*-linux-amd64.tar.gz
sudo systemctl stop wozai
sudo cp wozai-*-linux-amd64/wozai /usr/local/bin/wozai
sudo systemctl start wozai
```

### Docker 更新

```bash
docker compose down
docker compose build --no-cache
docker compose up -d
```

---

## 10. 数据备份与恢复

### 备份

```bash
# 使用提供的脚本
sudo bash deploy/backup-db.sh

# 备份存放在 /opt/wozai/backups/
```

### 手动备份

```bash
pg_dump -U wozai wozai | gzip > wozai_backup_$(date +%Y%m%d).sql.gz
```

### 恢复

```bash
# 从备份恢复
gunzip < wozai_backup_20260405.sql.gz | psql -U wozai wozai
```

### 自动备份（cron）

```bash
sudo crontab -e

# 每天凌晨 3 点备份，保留 30 天
0 3 * * * /opt/wozai-src/deploy/backup-db.sh >> /var/log/wozai-backup.log 2>&1
```

### Docker 备份

```bash
docker compose exec db pg_dump -U wozai wozai | gzip > wozai_backup_$(date +%Y%m%d).sql.gz
```

---

## 11. 故障排查

### 常见问题

| 现象 | 可能原因 | 解决方式 |
|------|----------|----------|
| `docker: command not found` (OpenCloudOS) | `get.docker.com` 不支持 OpenCloudOS | 使用 `sudo bash deploy.sh docker` 自动安装（已适配国内镜像），见下方详细说明 |
| Docker 安装时 `SSL connect error` / `Connection reset` | 国内 VPS 无法访问 `download.docker.com` | 改用国内镜像源，见下方详细说明 |
| `No match for argument: yum-utils` (OpenCloudOS) | OpenCloudOS 无此包 | 可忽略，直接手动添加 repo 文件即可安装 Docker |
| 启动失败 `config: DATABASE_URL is required` | .env 未配置 | 检查 `/opt/wozai/.env` |
| 启动失败 `db ping: dial error` | PostgreSQL 未启动或密码错 | `sudo systemctl start postgresql` 或检查密码 |
| 注册时提示 `服务异常` | 数据库表未创建 | 程序首次启动会自动迁移，检查日志 |
| 对话超时无响应 | DeepSeek API 不可达 | 检查服务器能否访问 `api.deepseek.com` |
| 语音生成失败 | SiliconFlow API Key 无效 | 检查 `SILICONFLOW_API_KEY` |
| 访问 80 端口无响应 | Nginx 未启动 | `sudo systemctl start nginx` |
| 429 Too Many Requests | 触发限流 | 正常安全机制，等待几秒后重试 |

#### 国内 VPS 安装 Docker 详细步骤（OpenCloudOS / TencentOS）

如果 `download.docker.com` 超时或被重置，使用以下命令（阿里云镜像源）：

```bash
# 清理之前失败的 repo
rm -f /etc/yum.repos.d/docker-ce.repo

# 添加阿里云 Docker 镜像源
cat > /etc/yum.repos.d/docker-ce.repo <<'EOF'
[docker-ce-stable]
name=Docker CE Stable - $basearch
baseurl=https://mirrors.aliyun.com/docker-ce/linux/centos/9/$basearch/stable
enabled=1
gpgcheck=1
gpgkey=https://mirrors.aliyun.com/docker-ce/linux/centos/gpg
EOF

# 安装 Docker
dnf makecache
dnf install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# 配置镜像加速 + 启动
mkdir -p /etc/docker
cat > /etc/docker/daemon.json <<'EOF'
{
  "registry-mirrors": ["https://mirror.ccs.tencentyun.com", "https://docker.mirrors.ustc.edu.cn"],
  "log-driver": "json-file",
  "log-opts": {"max-size": "5m", "max-file": "2"}
}
EOF
systemctl start docker && systemctl enable docker
docker --version
```

> 项目的 `deploy.sh` 已自动适配：检测到 `download.docker.com` 不可达时会自动切换到阿里云/腾讯云镜像。
> 直接执行 `sudo bash deploy.sh docker` 即可。

### 日志排查

```bash
# 应用日志
sudo journalctl -u wozai --since "10 minutes ago"

# Nginx 日志
sudo tail -50 /var/log/nginx/wozai_error.log

# PostgreSQL 日志
sudo tail -50 /var/log/postgresql/postgresql-*-main.log

# Docker 日志
docker compose logs --tail=50 app
docker compose logs --tail=50 db
```

### 端口检查

```bash
# 检查 8080 端口是否在监听
ss -tlnp | grep 8080

# 检查 80 端口（Nginx）
ss -tlnp | grep :80

# 检查 5432 端口（PostgreSQL）
ss -tlnp | grep 5432
```

### 防火墙

```bash
# Ubuntu (ufw)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# CentOS (firewalld)
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

---

## 12. 架构说明

```
┌──────────┐     ┌──────────┐     ┌──────────────────┐     ┌────────────────┐
│  浏览器   │────▶│  Nginx   │────▶│  WoZai (Go:8080) │────▶│ PostgreSQL:5432│
│          │◀────│  :80/443 │◀────│                  │◀────│                │
└──────────┘     └──────────┘     └────────┬─────────┘     └────────────────┘
                                           │
                                    ┌──────┴──────┐
                                    ▼             ▼
                              ┌──────────┐  ┌──────────────┐
                              │ DeepSeek │  │ SiliconFlow  │
                              │  对话 AI  │  │  语音合成 TTS │
                              └──────────┘  └──────────────┘
```

### 数据库表

| 表名       | 说明                   |
|------------|------------------------|
| `users`    | 用户账号（邮箱+密码哈希） |
| `souls`    | 数字灵魂（名称、性格、记忆）|
| `messages` | 聊天记录（角色+内容）   |

### API 端点

| 方法   | 路径                              | 说明         | 认证 |
|--------|-----------------------------------|-------------|------|
| GET    | `/health`                         | 健康检查     | 否   |
| POST   | `/api/v1/auth/register`           | 注册         | 否   |
| POST   | `/api/v1/auth/login`              | 登录         | 否   |
| POST   | `/api/v1/auth/refresh`            | 刷新令牌     | 否   |
| POST   | `/api/v1/souls`                   | 创建灵魂     | 是   |
| GET    | `/api/v1/souls`                   | 灵魂列表     | 是   |
| GET    | `/api/v1/souls/{id}`              | 灵魂详情     | 是   |
| PUT    | `/api/v1/souls/{id}`              | 更新灵魂     | 是   |
| DELETE | `/api/v1/souls/{id}`              | 删除灵魂     | 是   |
| POST   | `/api/v1/souls/{id}/chat`         | 发送消息     | 是   |
| GET    | `/api/v1/souls/{id}/messages`     | 聊天历史     | 是   |
| POST   | `/api/v1/souls/{id}/speak`        | 语音合成     | 是   |

### 安全措施

- JWT 认证 (HS256 + 自动刷新)
- bcrypt 密码哈希 (cost=12)
- CSP / X-Frame-Options / nosniff 安全头
- IP 令牌桶限流 (60 burst / 10 per sec)
- 1MB 请求体限制
- 非 root 运行 (systemd + Docker)
- PostgreSQL 参数化查询（防 SQL 注入）
