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
	go build -o bin/$(APP_NAME) ./cmd/server

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

# Docker 构建
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Docker 运行
docker-run:
	docker run -d \
		--name $(APP_NAME) \
		-p 8080:8080 \
		-v $(PWD)/data:/app/data \
		-v $(PWD)/nginx-conf:/etc/nginx/conf.d \
		-v $(PWD)/nginx-certs:/etc/nginx/certs \
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

# 仅使用 Docker 构建和运行（无需本地 Go）
docker-only: docker-build
	@echo "创建必要的目录..."
	@mkdir -p data nginx-conf nginx-certs logs
	@echo "启动 nginx-proxy 容器..."
	docker run -d \
		--name nginx-proxy \
		-p 8080:8080 \
		-v $(PWD)/data:/app/data \
		-v $(PWD)/nginx-conf:/etc/nginx/conf.d \
		-v $(PWD)/nginx-certs:/etc/nginx/certs \
		$(DOCKER_IMAGE)
	@echo "nginx-proxy 已启动在 http://localhost:8080"

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

# 生产环境部署
deploy: docker-build
	docker-compose up -d

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