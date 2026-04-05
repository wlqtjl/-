GO ?= go
BINARY := wozai

.PHONY: build run test lint clean vet deploy deploy-docker deploy-status

build:
	$(GO) build -ldflags="-s -w" -o $(BINARY) ./cmd/wozai

run:
	$(GO) run ./cmd/wozai

test:
	$(GO) test ./... -v -cover

vet:
	$(GO) vet ./...

lint: vet

clean:
	rm -f $(BINARY)

# === 低内存 VPS 部署命令 ===

# Docker 一键部署 (700MB VPS)
deploy:
	bash deploy.sh docker

deploy-docker:
	docker compose -f docker-compose.lowmem.yml up -d --build

deploy-down:
	docker compose -f docker-compose.lowmem.yml down

deploy-logs:
	docker compose -f docker-compose.lowmem.yml logs -f

deploy-status:
	bash deploy.sh status

# 交叉编译 Linux 二进制
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-s -w" -o $(BINARY) ./cmd/wozai
