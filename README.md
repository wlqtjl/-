# 我在 (WoZai) — 数字灵魂对话平台

与你的「数字灵魂」对话、倾听 TA 的声音。

基于 Go + PostgreSQL + DeepSeek AI + SiliconFlow TTS 构建，专为小内存 VPS 优化（700MB 内存即可运行）。

## 快速开始

```bash
git clone https://github.com/wlqtjl/- wozai
cd wozai
sudo bash deploy.sh
```

详细安装方式请参考 **[INSTALL.md](INSTALL.md)**（Docker / 裸机 / 手动安装）。

## 公网部署

想把 WoZai 部署到公网让其他人访问？请参考 **[DEPLOY.md](DEPLOY.md)**，包含：

- 云服务器选型对比与推荐
- 从零到上线的完整 10 步流程
- HTTPS 配置、域名绑定、自动备份
- 费用估算与安全加固

## 文档

| 文档 | 说明 |
|------|------|
| [INSTALL.md](INSTALL.md) | 安装部署文档（系统要求、三种安装方式、配置说明、运维、故障排查） |
| [DEPLOY.md](DEPLOY.md) | 公网部署指南（云服务器选型、完整部署流程、费用与安全加固） |

## 许可证

[LICENSE](LICENSE)