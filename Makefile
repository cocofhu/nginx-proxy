.PHONY: build run test clean docker-build docker-run

# 变量定义
APP_NAME=nginx-proxy
DOCKER_IMAGE=nginx-proxy:latest
GO_VERSION=1.21

# 检查 Go 是否安装
check-go:
	@which go > /dev/null || (echo "Go is not installed or not in PATH. Please install Go 1.21+ from https://golang.org/dl/" && exit 1)
	@echo "Go version: $$(go version)"

# 构建应用
build: check-go
	@mkdir -p bin
	CGO_ENABLED=1 go build -o bin/$(APP_NAME) ./cmd/server

# Docker 构建（用于 Dockerfile 中）
docker-build-binary:
	@mkdir -p bin
	CGO_ENABLED=1 go build -o bin/$(APP_NAME) ./cmd/server

# Alpine 构建（修复 SQLite 兼容性问题）
alpine-build:
	@mkdir -p bin
	CGO_ENABLED=1 go build -tags "sqlite_omit_load_extension" -o bin/$(APP_NAME) ./cmd/server

# 纯 Go 构建（无需 CGO，推荐用于 Docker）
build-no-cgo:
	@mkdir -p bin
	CGO_ENABLED=0 go build -o bin/$(APP_NAME) ./cmd/server

# 运行应用
run: check-go
	go run ./cmd/server/main.go

# 运行测试
test:
	go test -v ./...

# 清理构建文件
clean:
	rm -rf bin/
	rm -f *.db

# 安装依赖
deps: check-go
	go mod tidy
	go mod download

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run

# Docker 构建（Alpine 版本）
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Docker 构建（Debian 版本，解决 SQLite 编译问题）
docker-build-debian:
	docker build -f Dockerfile.debian -t $(DOCKER_IMAGE) .

# Docker 构建（最简单版本，纯 Go 无 CGO）
docker-build-simple:
	docker build -f Dockerfile.simple -t $(DOCKER_IMAGE) .

# Docker 运行
docker-run:
	@mkdir -p data nginx-conf nginx-certs logs config template
	docker run -d \
		--name $(APP_NAME) \
		-p 8080:8080 \
		-v $(PWD)/data:/app/data \
		-v $(PWD)/nginx-conf:/etc/nginx/conf.d \
		-v $(PWD)/nginx-certs:/etc/nginx/certs \
		-v $(PWD)/logs:/var/log/nginx \
		-v $(PWD)/config:/app/config \
		-v $(PWD)/template:/app/template \
		$(DOCKER_IMAGE)

# Docker Compose 启动（生产环境）
compose-up:
	docker-compose up -d

# Docker Compose 停止
compose-down:
	docker-compose down

# Docker 开发环境（无需本地 Go）
dev-docker:
	@echo "启动开发环境（无需本地 Go 环境）..."
	@mkdir -p data nginx-conf nginx-certs logs
	docker-compose -f docker-compose.dev.yml up -d

# 停止开发环境
dev-docker-down:
	docker-compose -f docker-compose.dev.yml down

# 单容器部署（推荐）
docker-single: docker-build
	@echo "创建必要的目录..."
	@mkdir -p data nginx-conf nginx-certs logs
	@echo "启动单容器 nginx-proxy（包含 nginx）..."
	docker run -d \
		--name nginx-proxy-complete \
		-p 80:80 \
		-p 443:443 \
		-p 8080:8080 \
		-v $(PWD)/data:/app/data \
		-v $(PWD)/nginx-conf:/etc/nginx/conf.d \
		-v $(PWD)/nginx-certs:/etc/nginx/certs \
		-v $(PWD)/logs:/var/log/nginx \
		-v $(PWD)/config:/app/config \
		-v $(PWD)/template:/app/template \
		$(DOCKER_IMAGE)
	@echo "nginx-proxy 已启动："
	@echo "  - API: http://localhost:8080"
	@echo "  - HTTP: http://localhost:80"
	@echo "  - HTTPS: https://localhost:443"

# 单容器 Docker Compose
single-compose:
	@mkdir -p data nginx-conf nginx-certs logs
	docker-compose -f docker-compose.single.yml up -d

# 仅使用 Docker 构建和运行（无需本地 Go）- 兼容旧版本
docker-only: docker-single

# 停止并清理 Docker 容器
docker-clean:
	-docker stop nginx-proxy
	-docker rm nginx-proxy

# 创建必要的目录
setup:
	mkdir -p data nginx-conf nginx-certs logs

# 开发环境初始化
dev-setup: setup deps
	@echo "开发环境初始化完成"

# 查看日志
logs:
	docker-compose logs -f nginx-proxy

# 重启服务
restart:
	docker-compose restart nginx-proxy

# 备份数据
backup:
	@echo "备份数据库..."
	cp data/nginx-proxy.db data/nginx-proxy.db.backup.$(shell date +%Y%m%d_%H%M%S)
	@echo "备份完成"

# 健康检查
health:
	curl -f http://localhost:8080/api/rules || exit 1

# CI/CD 相关命令
# 创建发布版本
release:
	@chmod +x scripts/release.sh
	@./scripts/release.sh $(TYPE)

# 生产环境部署（默认）
deploy: docker-build
	docker-compose up -d

# 部署到指定环境
deploy-env:
	@chmod +x scripts/deploy.sh
	@./scripts/deploy.sh $(ENV) $(VERSION)

# 生产环境部署
deploy-prod:
	@chmod +x scripts/deploy.sh
	@./scripts/deploy.sh production $(VERSION)

# 测试环境部署
deploy-staging:
	@chmod +x scripts/deploy.sh
	@./scripts/deploy.sh staging $(VERSION)

# 查看当前版本
version:
	@echo "Current version: $(shell cat VERSION)"

# 卷管理命令
# 创建挂载目录
setup-volumes:
	@echo "创建挂载目录..."
	@mkdir -p volumes/data volumes/nginx-conf volumes/nginx-certs volumes/nginx-logs volumes/templates volumes/config
	@cp config.json volumes/config/ 2>/dev/null || echo "配置文件不存在，跳过复制"
	@cp -r template/* volumes/templates/ 2>/dev/null || echo "模板目录为空，跳过复制"
	@echo "挂载目录创建完成！"

# 清理卷数据
clean-volumes:
	@echo "清理挂载目录..."
	@rm -rf volumes/
	@echo "挂载目录已清理！"

# 备份卷数据
backup-volumes:
	@echo "备份卷数据..."
	@tar -czf nginx-proxy-volumes-backup-$(shell date +%Y%m%d_%H%M%S).tar.gz volumes/
	@echo "备份完成！"

# 查看卷使用情况
volume-info:
	@echo "=== Docker 卷信息 ==="
	@docker volume ls | grep nginx-proxy || echo "没有找到 nginx-proxy 相关卷"
	@echo ""
	@echo "=== 本地挂载目录 ==="
	@if [ -d "volumes" ]; then du -sh volumes/*; else echo "volumes 目录不存在"; fi

# 构建特定版本的镜像
build-version:
	@VERSION=$$(cat VERSION) && \
	docker build -t $(DOCKER_IMAGE):$$VERSION . && \
	docker build -t $(DOCKER_IMAGE):latest .

# 推送镜像到仓库
push-version:
	@VERSION=$$(cat VERSION) && \
	docker push $(DOCKER_IMAGE):$$VERSION && \
	docker push $(DOCKER_IMAGE):latest