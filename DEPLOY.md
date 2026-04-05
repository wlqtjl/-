# 我在 (WoZai) — 公网部署完整指南

> 本文档帮你从零开始将 WoZai 部署到公网，让其他人随时随地可以访问。
> 技术安装细节请参考 [INSTALL.md](INSTALL.md)。

## 目录

1. [项目资源需求](#1-项目资源需求)
2. [公有云选择对比](#2-公有云选择对比)
3. [综合推荐](#3-综合推荐)
4. [完整部署流程（腾讯云轻量示例）](#4-完整部署流程)
5. [部署后运维速查](#5-部署后运维速查)
6. [费用总结](#6-费用总结)
7. [安全加固建议](#7-安全加固建议)

---

## 1. 项目资源需求

WoZai 使用 Go 静态编译 + Alpine Docker 镜像，资源需求极低：

| 组件 | 内存 | 说明 |
|------|------|------|
| 系统 + Docker | ~120MB | Alpine Linux 容器极低开销 |
| PostgreSQL | ~80MB | shared_buffers=32MB 调优 |
| WoZai Go 应用 | ~30MB | 静态二进制 + GOMEMLIMIT=50MiB |
| **合计** | **~230MB** | 700MB VPS 即可稳定运行 |

### 最低配置

| 项目 | 要求 |
|------|------|
| CPU | 1 核 |
| 内存 | 512 MB（使用 lowmem 配置） |
| 磁盘 | 5 GB SSD |
| 带宽 | 1 Mbps |

### 推荐配置

| 项目 | 要求 |
|------|------|
| CPU | 1–2 核 |
| 内存 | 1–2 GB |
| 磁盘 | 20 GB SSD |
| 带宽 | 3 Mbps |

### 网络要求

服务器需要能访问以下外部 API（中国大陆 VPS 均可直连）：

- `api.deepseek.com` — AI 对话
- `api.siliconflow.cn` — 语音合成

---

## 2. 公有云选择对比

### 第一梯队：国内高性价比（推荐）

#### 腾讯云 轻量应用服务器 ⭐ 首推

| 配置 | 价格 | 说明 |
|------|------|------|
| 2核/2G/40G SSD/4Mbps | ¥50/年（新用户首年） | 秒杀价，限新用户 |
| 2核/2G/50G SSD/4Mbps | ¥99/年（续费价） | 轻量级首选 |
| 2核/4G/60G SSD/5Mbps | ¥199/年 | 富余空间 |

优势：
- 自带公网 IP + 固定带宽
- 国内节点，访问 DeepSeek/SiliconFlow 延迟极低
- 支持一键 Docker 镜像，开箱即用
- 支持域名备案

官网：https://cloud.tencent.com/product/lighthouse

#### 阿里云 轻量应用服务器

| 配置 | 价格 | 说明 |
|------|------|------|
| 2核/2G/40G SSD/3Mbps | ¥99/年（新用户） | |
| 2核/4G/60G SSD/4Mbps | ¥199/年 | |

优势：
- 生态成熟，文档齐全
- 支持域名备案

注意：价格略高于腾讯云

#### 雨云 / 狗云 / RackNerd（预算极低场景）

| 配置 | 价格 | 说明 |
|------|------|------|
| 1核/1G/20G SSD | ¥25-40/月 | 国内小厂商 |
| 1核/768M/15G SSD | $10-15/年 | RackNerd 海外促销 |

注意：
- 海外节点需考虑国内访问延迟
- 小厂商稳定性不如大厂

### 第二梯队：海外免备案

#### Bandwagon (搬瓦工) / Vultr / DigitalOcean

| 配置 | 价格 | 说明 |
|------|------|------|
| 1核/1G/25G SSD | $5-6/月 | 香港/东京/新加坡节点 |
| 1核/512M/10G SSD | $2.5/月 | Vultr 最低配（够用） |

优势：
- 无需备案，即买即用
- 香港/东京节点国内延迟可接受

注意：中国大陆直接访问可能不稳定

---

## 3. 综合推荐

| 场景 | 推荐方案 | 预算 |
|------|----------|------|
| **国内用户为主** | 腾讯云轻量 2核2G | ¥50-99/年 |
| **不想备案** | Vultr 东京/香港 1核1G | $5/月 |
| **预算极低测试用** | RackNerd 促销款 | $10/年 |
| **已有服务器** | 直接部署 | ¥0 |

**最终推荐：腾讯云轻量应用服务器 2核2G（¥50-99/年）**

理由：国内访问快、价格最低、Docker 预装、域名备案方便、API 调用无延迟。

---

## 4. 完整部署流程

以腾讯云轻量应用服务器为例，其他云服务商步骤类似。

### 第 1 步：购买服务器

1. 访问 https://cloud.tencent.com/product/lighthouse
2. 选择配置：
   - **地域**：选离你最近的（如上海/广州/北京）
   - **镜像**：选择 **Ubuntu 22.04** 或 **Docker 基础镜像**
   - **套餐**：2核/2G/40G SSD/4Mbps
3. 完成支付，等待实例创建（约 1-2 分钟）
4. 在控制台记下 **公网 IP**（如 `123.45.67.89`）

### 第 2 步：配置防火墙

在腾讯云控制台 → 轻量应用服务器 → 防火墙，放行以下端口：

| 端口 | 协议 | 用途 |
|------|------|------|
| 22 | TCP | SSH 远程登录 |
| 80 | TCP | HTTP 访问 |
| 443 | TCP | HTTPS 访问 |

### 第 3 步：SSH 连接服务器

```bash
# 方式一：腾讯云控制台「登录」按钮（网页终端）
# 方式二：本地终端
ssh root@123.45.67.89
# 输入购买时设置的密码
```

### 第 4 步：安装 Docker

```bash
# 检查是否已安装（Docker 镜像通常已预装）
docker --version

# 如果未安装：
curl -fsSL https://get.docker.com | sh
systemctl start docker
systemctl enable docker
```

> **⚠️ OpenCloudOS / TencentOS 或国内 VPS 安装 Docker 报错？**
>
> 如果 `get.docker.com` 或 `download.docker.com` 超时/连接重置，是因为国内网络无法直连 Docker 官方源。
> 请改用国内镜像手动安装：
>
> ```bash
> # 1. 清理之前失败的 repo
> rm -f /etc/yum.repos.d/docker-ce.repo
>
> # 2. 添加阿里云 Docker 镜像源
> cat > /etc/yum.repos.d/docker-ce.repo <<'EOF'
> [docker-ce-stable]
> name=Docker CE Stable - $basearch
> baseurl=https://mirrors.aliyun.com/docker-ce/linux/centos/9/$basearch/stable
> enabled=1
> gpgcheck=1
> gpgkey=https://mirrors.aliyun.com/docker-ce/linux/centos/gpg
> EOF
>
> # 3. 安装
> dnf makecache && dnf install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
>
> # 4. 配置镜像加速 + 启动
> mkdir -p /etc/docker
> cat > /etc/docker/daemon.json <<'EOF'
> {
>   "registry-mirrors": ["https://mirror.ccs.tencentyun.com"],
>   "log-driver": "json-file",
>   "log-opts": {"max-size": "5m", "max-file": "2"}
> }
> EOF
> systemctl start docker && systemctl enable docker
> ```
>
> 或者直接使用项目的 `deploy.sh`，它已自动适配国内镜像源：
> ```bash
> sudo bash deploy.sh docker
> ```

### 第 5 步：拉取代码并部署

```bash
# 1. 克隆项目
git clone https://github.com/wlqtjl/- wozai
cd wozai

# ⚠️ 如果 clone 很慢或超时（国内 VPS 常见），使用 GitHub 加速镜像：
# git clone https://ghfast.top/https://github.com/wlqtjl/- wozai

# 2. 一键部署
sudo bash deploy.sh
# 选择选项 1: Docker 容器部署
#
# 脚本会自动完成：
#   - 安装 Docker（如未安装）
#   - 生成 .env 配置文件（含随机安全密钥）
#   - 构建 Docker 镜像
#   - 启动容器
#   - 执行健康检查
```

### 第 6 步：配置 API 密钥

部署脚本会自动生成数据库密码和 JWT 密钥，但 **AI 服务的 API 密钥必须手动填写**。

```bash
nano .env
```

修改以下两项：

```
DEEPSEEK_API_KEY=sk-你的DeepSeek密钥
SILICONFLOW_API_KEY=sk-你的SiliconFlow密钥
```

保存后重启使配置生效：

```bash
docker compose -f docker-compose.lowmem.yml restart
```

#### API 密钥获取方式

| API | 注册地址 | 步骤 |
|-----|---------|------|
| DeepSeek（AI 对话） | https://platform.deepseek.com | 注册 → API Keys → 创建 API Key |
| SiliconFlow（语音合成） | https://cloud.siliconflow.cn | 注册 → API 密钥 → 创建 |

> 两个平台新用户均有免费额度，按量计费，余额用完即停，无超额风险。

### 第 7 步：验证部署

```bash
# 健康检查
curl http://localhost/health
# 期望返回: {"status":"ok"}

# 查看容器状态
docker compose -f docker-compose.lowmem.yml ps

# 查看内存使用
docker stats --no-stream

# 查看日志
docker compose -f docker-compose.lowmem.yml logs -f app
```

在浏览器访问 `http://你的公网IP`（如 `http://123.45.67.89`），看到登录页面即部署成功。

### 第 8 步：绑定域名（推荐）

1. **购买域名**

   推荐 `.com` 或 `.cn` 后缀。可在腾讯云、阿里云或 Cloudflare 购买，约 ¥30-70/年。

2. **DNS 解析**

   在域名管理后台添加 A 记录：

   | 记录类型 | 主机记录 | 记录值 | TTL |
   |---------|---------|--------|-----|
   | A | @ | 123.45.67.89（你的公网 IP） | 600 |
   | A | www | 123.45.67.89（你的公网 IP） | 600 |

3. **域名备案**（国内服务器必须）
   - 腾讯云控制台 → 备案 → 按指引提交资料
   - 首次备案约 7-15 个工作日
   - 备案期间域名不可使用，但 **IP 直接访问不受影响**

### 第 9 步：配置 HTTPS（强烈推荐）

域名 DNS 解析生效后（备案通过后），配置免费 SSL 证书：

```bash
# 1. 安装 Nginx
apt install -y nginx

# 2. 复制项目提供的 Nginx 配置
cp ~/wozai/nginx-wozai.conf /etc/nginx/sites-available/wozai

# 3. 修改域名
nano /etc/nginx/sites-available/wozai
# 将 server_name _; 改为你的域名，例如:
#   server_name wozai.yourdomain.com;

# 4. 启用配置
ln -sf /etc/nginx/sites-available/wozai /etc/nginx/sites-enabled/wozai
rm -f /etc/nginx/sites-enabled/default
nginx -t && systemctl reload nginx

# 5. 安装 certbot 并自动签发 SSL 证书
apt install -y certbot python3-certbot-nginx
certbot --nginx -d wozai.yourdomain.com
# 按提示操作，建议选择「自动重定向 HTTP → HTTPS」

# 6. 验证自动续期
certbot renew --dry-run
```

完成后访问 `https://wozai.yourdomain.com` 即可。

> **注意**：使用 Docker 部署时，`docker-compose.lowmem.yml` 默认将应用映射到宿主机 80 端口。
> 安装 Nginx 后需将 Docker 端口改为其他端口（如 8080），避免与 Nginx 冲突：
>
> ```bash
> # 编辑 .env，添加或修改：
> LISTEN_PORT=8080
>
> # 重启容器
> docker compose -f docker-compose.lowmem.yml down
> docker compose -f docker-compose.lowmem.yml up -d
> ```
>
> Nginx 会将外部 80/443 端口的请求代理到 `127.0.0.1:8080`。

### 第 10 步：设置自动备份

```bash
# 创建备份目录
mkdir -p /opt/wozai-backups

# 编辑定时任务
crontab -e
```

添加以下行（每天凌晨 3 点备份，保留 30 天）：

```
0 3 * * * cd ~/wozai && docker compose -f docker-compose.lowmem.yml exec -T db pg_dump -U wozai wozai | gzip > /opt/wozai-backups/wozai_$(date +\%Y\%m\%d).sql.gz && find /opt/wozai-backups -name "*.sql.gz" -mtime +30 -delete
```

---

## 5. 部署后运维速查

| 操作 | 命令 |
|------|------|
| 查看状态 | `docker compose -f docker-compose.lowmem.yml ps` |
| 查看日志 | `docker compose -f docker-compose.lowmem.yml logs -f app` |
| 重启服务 | `docker compose -f docker-compose.lowmem.yml restart` |
| 停止服务 | `docker compose -f docker-compose.lowmem.yml down` |
| 查看内存 | `docker stats --no-stream` |
| 更新代码 | `cd ~/wozai && git pull && docker compose -f docker-compose.lowmem.yml up -d --build` |
| 数据库备份 | `docker compose -f docker-compose.lowmem.yml exec -T db pg_dump -U wozai wozai \| gzip > backup.sql.gz` |

更多运维命令请参考 [INSTALL.md](INSTALL.md)。

---

## 6. 费用总结

| 项目 | 年费用 | 说明 |
|------|--------|------|
| 云服务器 | ¥50-99 | 腾讯云轻量 2核2G |
| 域名 | ¥30-70 | 可选 |
| SSL 证书 | ¥0 | Let's Encrypt 免费自动续期 |
| DeepSeek API | ¥30-80/月 | 按量计费（100用户/天估算） |
| SiliconFlow API | ¥15-40/月 | 按量计费（100用户/天估算） |
| **总计** | **¥600-1500/年** | 用户少时费用更低 |

> **最低启动成本**：服务器 ¥50（首年）+ API 免费额度 = **¥50 即可上线**。

### API 费用明细

| 功能 | 单次消耗 | 月均预估（100用户/天） |
|------|---------|----------------------|
| AI 对话 | ~300 token/次 | ¥30-80 |
| 语音合成 | ~100 字符/次 | ¥15-40 |

两家平台均支持预充值，余额用完即停，无超额风险。

---

## 7. 安全加固建议

上线后建议执行以下安全加固措施：

### 7.1 SSH 安全

```bash
# 修改 SSH 端口（避免默认 22 端口被扫描）
nano /etc/ssh/sshd_config
# 修改 Port 22 为其他端口，如 Port 2222

# 禁用密码登录，改用 SSH Key
# 先在本地生成密钥：ssh-keygen -t ed25519
# 将公钥上传到服务器：ssh-copy-id -p 2222 root@你的服务器IP
# 然后修改 sshd_config：
#   PasswordAuthentication no
#   PubkeyAuthentication yes

# 重启 SSH
systemctl restart sshd
```

> 修改 SSH 端口后，记得在云服务商防火墙中放行新端口。

### 7.2 防暴力破解

```bash
apt install -y fail2ban
systemctl enable fail2ban
systemctl start fail2ban
```

### 7.3 系统更新

```bash
# 定期更新系统软件包
apt update && apt upgrade -y
```

### 7.4 可用性监控

- **腾讯云自带**基础监控（CPU、内存、带宽）
- 推荐接入 [UptimeRobot](https://uptimerobot.com)（免费）监控网站可用性
  - 添加 HTTP 监控：`https://你的域名/health`
  - 设置告警通知（邮件/Webhook）

---

## 附录：部署架构图

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

- **浏览器** → Nginx（80/443）→ WoZai Go 应用（8080）→ PostgreSQL（5432）
- WoZai 应用同时调用 DeepSeek（对话）和 SiliconFlow（语音合成）的外部 API
- Docker 模式下 PostgreSQL 不暴露端口到宿主机，仅容器内部通信
