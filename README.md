# 轻量级 Nginx 配置管理器

一个用 Go 编写的轻量级工具，通过 REST API 管理 Nginx 反向代理规则，支持基于来源 IP 的分流、泛域名证书管理和自动配置生成。

## 功能特性

- 🚀 **REST API 管理**：完整的 CRUD 操作管理反向代理规则
- 🌐 **智能分流**：基于来源 IP 段的流量分发，支持多条件分流配置
- 🔒 **证书管理**：支持上传 SSL 证书文件
- 📝 **自动配置**：自动生成和验证 Nginx 配置文件
- 🔄 **热重载**：配置变更后自动重载 Nginx
- 💾 **持久化存储**：使用 SQLite 数据库存储配置（纯 Go 驱动）
- 🐳 **容器化**：提供 Docker 支持，无 CGO 依赖
- 🎯 **简洁界面**：专注于代理配置管理的简洁 Web 界面
- 🖥️ **Web 管理界面**：现代化的响应式 Web 界面

- 📈 **可视化图表**：请求趋势和响应时间图表
- 📋 **日志管理**：实时日志查看和过滤功能

## 技术栈

- **Go 1.21+**：主要编程语言（纯 Go，无 CGO 依赖）
- **Gin**：HTTP 框架
- **SQLite + GORM**：数据库和 ORM（使用 modernc.org/sqlite 纯 Go 驱动）
- **Docker**：容器化部署
- **Nginx**：反向代理服务器

## 快速开始

### 本地运行

1. **克隆项目**
```bash
git clone <repository-url>
cd nginx-proxy
```

2. **安装依赖**
```bash
make deps
```

3. **构建应用（纯 Go）**
```bash
make build
```

4. **启动服务**
```bash
make run
# 或者使用启动脚本
./start.sh
```

5. **访问 Web 界面**
```
打开浏览器访问：http://localhost:8080
```

Web 界面功能：
- 📊 **仪表板**：系统概览和快速操作入口
- ⚙️ **代理配置**：管理反向代理规则，支持基于IP的分流配置
- 🔒 **证书管理**：上传、管理SSL证书，支持证书选择和过期提醒

### Docker 运行（推荐）

项目使用纯 Go 构建，无 CGO 依赖，确保最佳兼容性。Docker 镜像已包含完整的 Web 静态文件：

```bash
# 构建并启动（包含 nginx）
make docker-single
```

或者手动构建：
```bash
# 构建镜像
make docker-build

# 运行容器
make docker-run
```

**Docker 镜像特性**：
- ✅ 包含完整的 Web 管理界面静态文件
- ✅ 自动复制 `/web/static/` 目录到容器
- ✅ 支持静态文件卷挂载（可选）
- ✅ 纯 Go 构建，无 CGO 依赖

**测试 Docker 构建**：
```bash
# 运行 Docker 测试脚本
./scripts/test-docker.sh
```



## API 接口

### 规则管理
- `GET /api/rules` - 获取所有规则
- `GET /api/rules/{id}` - 获取指定规则
- `POST /api/rules` - 创建新规则
- `PUT /api/rules/{id}` - 更新规则
- `DELETE /api/rules/{id}` - 删除规则

### 系统管理
- `POST /api/reload` - 重载 Nginx 配置
- `GET /api/health` - 健康检查

### 证书管理
- `GET /api/certificates` - 获取所有证书
- `GET /api/certificates/{id}` - 获取指定证书
- `POST /api/certificates` - 上传 SSL 证书（支持自动解析证书信息）
- `DELETE /api/certificates/{id}` - 删除证书



### 规则配置示例

```json
{
  "server_name": "example.com",
  "listen_ports": [443],
  "ssl_cert": "/etc/nginx/certs/example.com.crt",
  "ssl_key": "/etc/nginx/certs/example.com.key",
  "locations": [
    {
      "path": "/",
      "upstreams": [
        {
          "condition_ip": "192.168.1.0/24",
          "target": "http://internal-server:8080"
        },
        {
          "condition_ip": "0.0.0.0/0",
          "target": "http://public-server:8080"
        }
      ]
    }
  ]
}
```

## 配置文件

### config.json
```json
{
  "port": "8080",
  "nginx_path": "/usr/sbin/nginx",
  "config_dir": "/etc/nginx/conf.d",
  "cert_dir": "/etc/nginx/certs",
  "database_path": "./nginx-proxy.db",
  "template_dir": "./template"
}
```

## 目录结构

```
nginx-proxy/
├── cmd/server/           # 应用入口
├── internal/
│   ├── api/             # API 处理器
│   ├── core/            # 核心逻辑
│   └── db/              # 数据库模型
├── template/            # Nginx 配置模板
├── config.json          # 配置文件
├── Dockerfile           # Docker 构建文件
└── Makefile            # 构建脚本
```

## 部署说明

### 生产环境部署

1. **创建必要目录**
```bash
mkdir -p data nginx-conf nginx-certs logs config template
```

2. **复制配置文件**
```bash
cp config.json config/
cp -r template/* template/
```

3. **启动服务**
```bash
docker run -d \
  --name nginx-proxy \
  -p 80:80 \
  -p 443:443 \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/nginx-conf:/etc/nginx/conf.d \
  -v $(pwd)/nginx-certs:/etc/nginx/certs \
  -v $(pwd)/logs:/var/log/nginx \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/template:/app/template \
  nginx-proxy:latest
```

### 健康检查

```bash
# 检查 API 状态
curl http://localhost:8080/api/rules

# 检查容器状态
docker ps | grep nginx-proxy

# 查看日志
docker logs nginx-proxy
```

## 开发指南

### 构建选项

```bash
# 本地构建（纯 Go）
make build

# Docker 构建
make docker-build
```

### 测试

```bash
# 运行测试
make test

# 代码格式化
make fmt

# 代码检查
make lint
```

### 开发环境

```bash
# 启动开发环境
make dev-setup
```

## 故障排除

### 常见问题

1. **SQLite 编译错误**
   - 项目已使用纯 Go SQLite 驱动，无需 CGO
   - 如遇问题，参考 `SQLITE_FIX.md`

2. **Docker 卷挂载冲突**
   - 使用目录挂载而非文件挂载
   - 参考 `DEPLOYMENT.md`

3. **Nginx 配置错误**
   - 检查生成的配置：`ls -la nginx-conf/`
   - 测试配置：`nginx -t`

### 日志查看

```bash
# 应用日志
docker logs nginx-proxy

# Nginx 日志
docker exec nginx-proxy tail -f /var/log/nginx/access.log
docker exec nginx-proxy tail -f /var/log/nginx/error.log
```

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License

## 支持

如有问题，请提交 GitHub Issues 进行问题反馈。