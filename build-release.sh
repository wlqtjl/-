#!/usr/bin/env bash
#
# build-release.sh — 交叉编译 Linux amd64 二进制并打包为安装包
# 使用方式: bash deploy/build-release.sh
#
set -euo pipefail

GO="${GO:-go}"
VERSION="${VERSION:-$(date +%Y%m%d)}"
DIST_DIR="dist"
PKG_NAME="wozai-${VERSION}-linux-amd64"
PKG_DIR="${DIST_DIR}/${PKG_NAME}"

echo "=== 构建发布包 v${VERSION} ==="

# Clean
rm -rf "$DIST_DIR"

# Cross compile
echo "编译 linux/amd64 ..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $GO build -ldflags="-s -w" -o "${DIST_DIR}/wozai" ./cmd/wozai

# Assemble package
echo "组装安装包..."
mkdir -p "$PKG_DIR/deploy"

# Binary
cp "${DIST_DIR}/wozai" "$PKG_DIR/"

# Deploy files
cp deploy/install.sh    "$PKG_DIR/deploy/"
cp deploy/init-db.sh    "$PKG_DIR/deploy/"
cp deploy/wozai.service "$PKG_DIR/deploy/"
cp deploy/nginx-wozai.conf "$PKG_DIR/deploy/"

# Config template
cp .env.example "$PKG_DIR/"

# Docker files
cp Dockerfile         "$PKG_DIR/"
cp docker-compose.yml "$PKG_DIR/"
cp .dockerignore      "$PKG_DIR/"

# Source (for Docker build)
cp -r cmd internal web go.mod go.sum Makefile "$PKG_DIR/"

# Docs
cp INSTALL.md "$PKG_DIR/" 2>/dev/null || true

# Create tarball
echo "打包..."
cd "$DIST_DIR"
tar czf "${PKG_NAME}.tar.gz" "$PKG_NAME"
cd ..

SIZE=$(du -h "${DIST_DIR}/${PKG_NAME}.tar.gz" | cut -f1)
echo ""
echo "=== 构建完成 ==="
echo "安装包: ${DIST_DIR}/${PKG_NAME}.tar.gz (${SIZE})"
echo ""
echo "部署到目标服务器:"
echo "  scp ${DIST_DIR}/${PKG_NAME}.tar.gz root@your-server:/tmp/"
echo "  ssh root@your-server"
echo "  cd /tmp && tar xzf ${PKG_NAME}.tar.gz"
echo "  cd ${PKG_NAME} && sudo bash deploy/install.sh"
