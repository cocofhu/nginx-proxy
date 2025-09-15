# 安装指南

## 方式一：安装 Go 环境（推荐）

### macOS 安装 Go

#### 使用 Homebrew（推荐）
```bash
# 安装 Homebrew（如果还没有安装）
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 安装 Go
brew install go

# 验证安装
go version
```

#### 手动安装
1. 访问 [Go 官网](https://golang.org/dl/)
2. 下载适合 macOS 的安装包（如 go1.21.x.darwin-amd64.pkg 或 go1.21.x.darwin-arm64.pkg）
3. 双击安装包进行安装
4. 重启终端，验证安装：`go version`

### 配置环境变量
如果 `go` 命令仍然找不到，请添加到 PATH：

```bash
# 编辑 shell 配置文件
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc
# 或者对于 bash
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bash_profile

# 重新加载配置
source ~/.zshrc
# 或者
source ~/.bash_profile
```

## 方式二：使用 Docker（无需安装 Go）

如果您不想安装 Go 环境，可以直接使用 Docker：

### 构建和运行
```bash
# 构建 Docker 镜像
docker build -t nginx-proxy .

# 创建必要的目录
mkdir -p data nginx-conf nginx-certs logs

# 运行容器
docker run -d \
  --name nginx-proxy \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/nginx-conf:/etc/nginx/conf.d \
  -v $(pwd)/nginx-certs:/etc/nginx/certs \
  nginx-proxy
```

### 使用 Docker Compose
```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f nginx-proxy

# 停止服务
docker-compose down
```

## 方式三：使用预编译二进制文件

如果您有其他平台的 Go 环境，可以交叉编译：

```bash
# 在有 Go 环境的机器上编译 macOS 版本
GOOS=darwin GOARCH=amd64 go build -o nginx-proxy-darwin-amd64 ./cmd/server
# 或者 ARM64 版本
GOOS=darwin GOARCH=arm64 go build -o nginx-proxy-darwin-arm64 ./cmd/server
```

## 验证安装

安装完成后，验证项目是否可以正常运行：

```bash
# 检查 Go 版本
go version

# 安装依赖
make deps

# 构建项目
make build

# 运行项目
make run
```

## 常见问题

### Q: 提示 "go: command not found"
A: Go 没有正确安装或不在 PATH 中，请按照上述步骤重新安装。

### Q: 提示权限错误
A: 确保有足够的权限创建目录和文件：
```bash
sudo chown -R $(whoami) /etc/nginx/conf.d /etc/nginx/certs
```

### Q: Docker 构建失败
A: 确保 Docker 已安装并正在运行：
```bash
docker --version
docker info
```

## 快速测试

安装完成后，可以快速测试 API：

```bash
# 启动服务
make run

# 在另一个终端测试 API
curl http://localhost:8080/api/rules
```

应该返回空的规则列表：`{"rules":[]}`