# 我在 (WoZai) — 腾讯云部署完整指南

**域名**: `https://www.wozai.org/`
**推荐配置**: 腾讯云轻量应用服务器 2核/2GB/50GB SSD (~¥45-65/月)
**最低配置**: 1核/1GB (~¥30/月，使用 lowmem 模式)

---

## 目录

1. [第一阶段：准备工作](#第一阶段准备工作)
2. [第二阶段：服务器环境配置](#第二阶段服务器环境配置)
3. [第三阶段：部署应用](#第三阶段部署应用)
4. [第四阶段：域名 + HTTPS 配置](#第四阶段域名--https-配置)
5. [第五阶段：运维与安全加固](#第五阶段运维与安全加固)
6. [日常运维命令速查](#日常运维命令速查)
7. [费用总览](#费用总览)
8. [部署架构总览](#部署架构总览)
9. [关键注意事项](#关键注意事项)

---

## 第一阶段：准备工作

### 步骤 1：申请外部 AI API 密钥（10分钟）

> **这两个 API 是必须的，没有它们服务无法工作。**

| API | 用途 | 申请地址 | 费用 |
|-----|------|---------|------|
| DeepSeek | AI 对话 | https://platform.deepseek.com | 新用户免费额度，之后按 token 计费 |
| SiliconFlow | 语音合成 TTS | https://cloud.siliconflow.cn | 新用户免费额度，之后按字符计费 |

**DeepSeek 申请步骤**:
1. 打开 https://platform.deepseek.com → 注册账号
2. 登录后进入「API Keys」页面
3. 点击「创建 API Key」
4. 复制 `sk-` 开头的密钥，**记下来**

**SiliconFlow 申请步骤**:
1. 打开 https://cloud.siliconflow.cn → 注册账号
2. 进入「API 密钥」页面
3. 创建新的 API Key
4. 复制 `sk-` 开头的密钥，**记下来**

---

### 步骤 2：购买腾讯云服务器（5分钟）

1. 登录腾讯云控制台：https://console.cloud.tencent.com
2. 进入「轻量应用服务器」（Lighthouse）
   - 路径：控制台 → 产品 → 轻量应用服务器 → 购买
3. 选择配置：
   - **地域**：选离你最近的（如上海/北京/广州）
   - **镜像**：选「系统镜像」→ **Ubuntu 22.04 LTS**
   - **配置**：
     - 推荐：2核 / 2GB / 50GB SSD（约 ¥45-65/月）
     - 最低：1核 / 1GB / 40GB（约 ¥30/月，需要用 lowmem 配置）
   - **购买时长**：按需
4. 点击「立即购买」→ 完成支付
5. **记下服务器的公网 IP 地址**（例如 `43.xxx.xxx.xxx`）

---

### 步骤 3：配置服务器安全组/防火墙（3分钟）

腾讯云轻量应用服务器使用「防火墙」而不是传统安全组：

1. 进入轻量应用服务器控制台 → 选中你的实例
2. 点击「防火墙」选项卡
3. 添加以下规则：

| 协议 | 端口 | 来源 | 策略 | 用途 |
|------|------|------|------|------|
| TCP | 22 | 0.0.0.0/0 | 允许 | SSH 远程管理 |
| TCP | 80 | 0.0.0.0/0 | 允许 | HTTP 访问 |
| TCP | 443 | 0.0.0.0/0 | 允许 | HTTPS 访问 |

> ⚠️ **不要**开放 5432 (PostgreSQL) 和 8080 端口到公网。

---

### 步骤 4：域名 DNS 解析配置（5分钟）

> 你的域名 `wozai.org` 需要先做 ICP 备案才能指向国内服务器（见步骤 5）。但 DNS 解析可以先配好。

1. 登录你的域名管理面板（如果域名在腾讯云，进入「域名注册」→「我的域名」→「解析」）
2. 添加以下 DNS 记录：

| 主机记录 | 类型 | 记录值 | TTL |
|---------|------|--------|-----|
| `@` | A | `你的服务器公网IP` | 600 |
| `www` | A | `你的服务器公网IP` | 600 |

3. 等待 DNS 生效（通常 5-10 分钟，最长 48 小时）

**验证 DNS 是否生效**：
```bash
# 在本地电脑执行
ping wozai.org
ping www.wozai.org
# 应该解析到你的服务器 IP
```

---

### 步骤 5：ICP 备案（重要！1-20个工作日）

> ⚠️ **在中国大陆，域名指向境内服务器必须完成 ICP 备案，否则网站会被封禁。**

1. 登录腾讯云控制台 → 搜索「备案」→ 进入「网站备案」
2. 路径：https://console.cloud.tencent.com/beian
3. 按流程提交：
   - **主体信息**：个人备案填身份信息，企业备案填营业执照
   - **网站信息**：
     - 网站名称：`我在`（不能含"博客""论坛"等敏感词）
     - 域名：`wozai.org`
     - 网站首页：`www.wozai.org`
   - **上传材料**：身份证正反面照片、腾讯云授权书（系统自动生成）
4. 提交后腾讯云初审（1-2 个工作日），通过后提交管局审核（3-20 个工作日）
5. 备案通过后会收到「备案号」（如：沪ICP备XXXXXXXX号）

> 💡 **在等待备案期间，可以先用 IP 直接访问（http://你的IP），提前完成部署和测试。**

---

## 第二阶段：服务器环境配置

### 步骤 6：SSH 登录服务器（2分钟）

```bash
# 在你的本地电脑执行
ssh root@你的服务器公网IP

# 如果腾讯云使用密钥登录:
ssh -i /path/to/your-key.pem root@你的服务器公网IP
```

首次登录后，建议先更新系统：
```bash
apt update && apt upgrade -y
```

---

### 步骤 7：安装 Docker（3分钟）

> 项目自带 deploy.sh 会自动安装 Docker，但手动安装更可控。

```bash
# 安装 Docker (腾讯云源更快)
curl -fsSL https://mirrors.cloud.tencent.com/docker-ce/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg

echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://mirrors.cloud.tencent.com/docker-ce/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

apt update
apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# 验证
docker --version
docker compose version

# 设置开机启动
systemctl enable docker
```

---

### 步骤 8：获取项目代码（2分钟）

```bash
# 安装 git
apt install -y git

# 克隆项目
cd /opt
git clone https://github.com/wlqtjl/- wozai
cd /opt/wozai
```

---

### 步骤 9：配置环境变量（5分钟）

```bash
# 复制配置模板
cp .env.example .env

# 生成安全密钥
JWT_SECRET=$(openssl rand -base64 48)
PG_PASSWORD=$(openssl rand -base64 24)

echo "生成的 JWT 密钥: $JWT_SECRET"
echo "生成的数据库密码: $PG_PASSWORD"

# 编辑配置文件
nano .env
```

**在 `.env` 中修改以下内容**（其余保持默认即可）：

```bash
# ===== 数据库 =====
POSTGRES_PASSWORD=粘贴上面生成的数据库密码

# ===== 安全 =====
JWT_SECRET=粘贴上面生成的JWT密钥

# ===== AI 对话 (必填) =====
DEEPSEEK_API_KEY=sk-你在步骤1申请的DeepSeek密钥

# ===== 语音合成 (必填) =====
SILICONFLOW_API_KEY=sk-你在步骤1申请的SiliconFlow密钥

# ===== 多模型 (可选，Gemma免费) =====
# AI_PROVIDER=deepseek
# GEMMA_API_KEY=你的Google_AI_Studio_API_Key
```

保存退出（Ctrl+X → Y → Enter）。

设置文件权限：
```bash
chmod 600 .env
```

---

## 第三阶段：部署应用

### 步骤 10：使用 Docker 一键部署（5分钟）

**方法 A — 使用自带部署脚本（推荐）**：
```bash
cd /opt/wozai
sudo bash deploy.sh docker
```

**方法 B — 手动 docker compose**：
```bash
cd /opt/wozai

# 内存 ≤ 2GB 使用低内存配置（推荐）
docker compose -f docker-compose.lowmem.yml up -d --build

# 内存 > 2GB 使用标准配置
# docker compose up -d --build
```

### 步骤 11：验证服务运行（2分钟）

```bash
# 查看容器状态（应显示 healthy）
docker compose -f docker-compose.lowmem.yml ps

# 查看应用日志
docker compose -f docker-compose.lowmem.yml logs -f app

# 健康检查
curl http://localhost:80/health
# 期望返回: {"status":"ok"} 或类似

# 查看内存占用
docker stats --no-stream
free -h
```

### 步骤 12：用 IP 测试访问（1分钟）

在浏览器中打开：
```
http://你的服务器公网IP
```

你应该能看到「我在」的界面，可以注册账号并测试对话功能。

---

## 第四阶段：域名 + HTTPS 配置

> ⚠️ **这一步需要 ICP 备案通过后才能进行。备案未通过前，只能用 IP 访问。**

### 步骤 13：安装 Nginx 反向代理（3分钟）

首先修改 Docker 端口映射，避免与 Nginx 冲突：

```bash
cd /opt/wozai

# 在 .env 中设置应用端口为 8080（避免和 Nginx 的 80 端口冲突）
echo "LISTEN_PORT=8080" >> .env

# 重启容器使端口映射生效
docker compose -f docker-compose.lowmem.yml up -d
```

然后安装并配置 Nginx：

```bash
apt install -y nginx

# 复制项目自带的 Nginx 配置
cp /opt/wozai/nginx-wozai.conf /etc/nginx/sites-available/wozai

# 启用配置
ln -sf /etc/nginx/sites-available/wozai /etc/nginx/sites-enabled/wozai
rm -f /etc/nginx/sites-enabled/default

# 测试配置
nginx -t

# 重启 Nginx
systemctl restart nginx
systemctl enable nginx
```

> 项目自带的 `nginx-wozai.conf` 已预配置 `wozai.org` 和 `www.wozai.org` 域名，
> 包含 certbot ACME 验证路径和 HTTP→HTTPS 重定向支持。

---

### 步骤 14：申请 HTTPS SSL 证书（3分钟）

```bash
# 安装 certbot
apt install -y certbot python3-certbot-nginx

# 申请证书（自动配置 Nginx）
certbot --nginx -d wozai.org -d www.wozai.org

# 交互步骤：
# 1. 输入邮箱（用于续期通知）
# 2. 同意服务条款 → Y
# 3. 是否重定向 HTTP 到 HTTPS → 选 2 (Redirect)
```

Certbot 会自动：
- 申请 Let's Encrypt 免费 SSL 证书
- 修改 Nginx 配置添加 SSL
- 设置自动续期（每 90 天自动续期）

**验证 HTTPS**：
```bash
# 测试自动续期
certbot renew --dry-run

# 浏览器访问
# https://www.wozai.org
```

---

### 步骤 15：最终验证（2分钟）

在浏览器中测试：

1. ✅ 打开 `https://www.wozai.org` → 应看到「我在」首页
2. ✅ 打开 `http://www.wozai.org` → 应自动跳转到 HTTPS
3. ✅ 打开 `https://wozai.org` → 应正常访问（或跳转到 www）
4. ✅ 注册一个新账号 → 应成功
5. ✅ 创建一个数字灵魂 → 应成功
6. ✅ 发送一条消息测试对话 → 应收到 AI 回复
7. ✅ 测试语音合成 → 应能播放语音

---

## 第五阶段：运维与安全加固

### 步骤 16：设置自动备份（2分钟）

```bash
# 创建备份目录
mkdir -p /opt/wozai/backups

# 添加每日自动备份 (凌晨3点)
crontab -e
# 添加以下行：
0 3 * * * docker compose -f /opt/wozai/docker-compose.lowmem.yml exec -T db pg_dump -U wozai wozai | gzip > /opt/wozai/backups/wozai_$(date +\%Y\%m\%d).sql.gz && find /opt/wozai/backups -name "*.sql.gz" -mtime +30 -delete
```

### 步骤 17：设置系统防火墙（1分钟）

```bash
ufw allow 22/tcp    # SSH
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS
ufw --force enable
ufw status
```

### 步骤 18：设置 Swap（可选，小内存服务器推荐）

```bash
# 创建 1GB swap
fallocate -l 1G /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile

# 永久生效
echo '/swapfile none swap sw 0 0' >> /etc/fstab

# 验证
free -h
```

### 步骤 19：配置日志轮转（防止磁盘写满）

Docker daemon 日志已在 deploy.sh 中限制为 5MB。额外确认：
```bash
# 检查 Docker 日志配置
cat /etc/docker/daemon.json
# 应包含: "max-size": "5m", "max-file": "2"

# 如果没有，创建它：
mkdir -p /etc/docker
cat > /etc/docker/daemon.json <<'EOF'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "5m",
    "max-file": "2"
  }
}
EOF
systemctl restart docker
```

---

## 日常运维命令速查

```bash
# ===== 查看状态 =====
docker compose -f /opt/wozai/docker-compose.lowmem.yml ps
docker stats --no-stream
free -h

# ===== 查看日志 =====
docker compose -f /opt/wozai/docker-compose.lowmem.yml logs -f app
docker compose -f /opt/wozai/docker-compose.lowmem.yml logs -f db

# ===== 重启服务 =====
docker compose -f /opt/wozai/docker-compose.lowmem.yml restart

# ===== 更新代码 =====
cd /opt/wozai
git pull
docker compose -f docker-compose.lowmem.yml up -d --build

# ===== 手动备份数据库 =====
docker compose -f /opt/wozai/docker-compose.lowmem.yml exec -T db pg_dump -U wozai wozai | gzip > /opt/wozai/backups/manual_$(date +%Y%m%d_%H%M).sql.gz

# ===== 恢复数据库 =====
gunzip < /opt/wozai/backups/wozai_20260405.sql.gz | docker compose -f /opt/wozai/docker-compose.lowmem.yml exec -T db psql -U wozai wozai

# ===== 停止服务 =====
docker compose -f /opt/wozai/docker-compose.lowmem.yml down

# ===== 完全重建 =====
docker compose -f /opt/wozai/docker-compose.lowmem.yml down
docker compose -f /opt/wozai/docker-compose.lowmem.yml up -d --build --force-recreate
```

---

## 费用总览

| 项目 | 月费用（预估） | 说明 |
|------|--------------|------|
| 腾讯云轻量服务器 (2核/2G) | ¥45-65 | 按月付费，年付更优惠 |
| 域名 wozai.org | ¥60-100/年 | 已购买 |
| SSL 证书 | 免费 | Let's Encrypt 自动续期 |
| DeepSeek API | ¥30-80 | 按用量，新用户有免费额度 |
| SiliconFlow API | ¥15-40 | 按用量，新用户有免费额度 |
| ICP 备案 | 免费 | 腾讯云免费代提交 |
| **合计** | **¥90-185/月** | 用户量小时费用更低 |

---

## 部署架构总览

```
用户浏览器
    │
    ▼
https://www.wozai.org
    │
    ▼ (DNS 解析)
┌──────────────────────────────────────┐
│        腾讯云轻量应用服务器            │
│        (Ubuntu 22.04)                │
│                                      │
│  ┌─────────┐    ┌─────────────────┐  │
│  │  Nginx   │───▶│ WoZai App      │  │
│  │  :443    │    │ (Docker :8080)  │  │
│  │  SSL终端 │    │                 │  │
│  └─────────┘    └────────┬────────┘  │
│                          │           │
│                 ┌────────▼────────┐  │
│                 │  PostgreSQL     │  │
│                 │  (Docker :5432) │  │
│                 │  仅内部访问      │  │
│                 └─────────────────┘  │
└──────────────────────────────────────┘
         │                    │
         ▼                    ▼
  ┌──────────────┐  ┌────────────────┐
  │ DeepSeek API │  │ SiliconFlow API│
  │ (AI 对话)     │  │ (语音合成 TTS) │
  └──────────────┘  └────────────────┘
```

---

## 关键注意事项

1. **ICP 备案是最关键的环节** — 没有备案号，域名指向国内服务器会被工信部封禁。**建议第一天就提交备案申请**，等待期间用 IP 访问测试。

2. **API 密钥安全** — `.env` 文件已设为 `chmod 600`，不要提交到 Git。

3. **定期更新** — 定期 `git pull` + 重新构建，获取最新功能和安全修复。

4. **监控磁盘** — `df -h` 定期检查磁盘使用，主要是数据库和日志。

5. **如果使用 Gemma（免费模型）** — 在 `.env` 中设置 `AI_PROVIDER=gemma` 和 `GEMMA_API_KEY`，可以降低 AI 对话的 API 成本。
