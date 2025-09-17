# 智能反向代理管理器

基于 **OpenResty + Go 接口** 架构的智能反向代理管理系统，支持复杂路由规则、头部条件匹配和动态配置管理。

## 🎯 核心特性

### 🚀 **智能路由系统**

- **头部条件"且"关系**：支持多个 HTTP 头部条件同时匹配
- **IP 段匹配**：支持 CIDR 格式的 IP 段路由（如 `192.168.1.0/24`）
- **动态路由判断**：通过 Go 接口实现复杂路由逻辑
- **实时路由切换**：无需重启即可更新路由规则

### 🔧 **管理功能**

- **REST API 管理**：完整的 CRUD 操作管理反向代理规则
- **Web 管理界面**：现代化的响应式管理界面
- **证书管理**：SSL 证书上传、管理和自动配置
- **腾讯云证书集成**：自动申请、续期和管理腾讯云免费SSL证书
- **智能证书清理**：删除证书时同步删除腾讯云端证书
- **配置验证**：自动验证 OpenResty 配置正确性
- **热重载**：配置变更后自动重载 OpenResty

### 💾 **数据存储**

- **SQLite 数据库**：使用纯 Go 驱动，无 CGO 依赖
- **持久化配置**：所有路由规则持久化存储
- **证书状态跟踪**：实时跟踪腾讯云证书状态和过期时间
- **配置备份**：自动生成配置文件备份

### 🔒 **SSL证书管理**

- **多证书来源**：支持本地上传和腾讯云证书
- **自动申请**：一键申请腾讯云免费SSL证书
- **智能续期**：证书到期前自动续期，无需人工干预
- **同步删除**：删除证书时自动清理腾讯云端对应证书
- **状态监控**：实时监控证书状态（申请中/正常/即将过期/已过期）
- **灵活配置**：支持仅HTTPS或HTTPS+HTTP重定向模式

## 🏗️ 技术架构

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Client        │───▶│   OpenResty      │───▶│   Go Service    │
│   Request       │    │   (Nginx + Lua)  │    │   (Port 8080)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │                         │
                              │   HTTP POST             │
                              │   /api/route            │
                              │                         │
                              ▼                         ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │  Route Decision  │    │  Route Logic    │
                       │  (Lua Script)    │    │  (Go Handler)   │
                       └──────────────────┘    └─────────────────┘
```

### 技术栈

- **OpenResty**：高性能 Web 平台（Nginx + LuaJIT）
- **Go 1.21+**：路由逻辑处理（纯 Go，无 CGO 依赖）
- **Gin**：HTTP 框架
- **SQLite + GORM**：数据库和 ORM
- **Lua**：动态路由脚本
- **Docker**：容器化部署

## 🚀 快速开始

### Docker 部署（推荐）

使用 Docker 一键部署 OpenResty + Go 服务：

```bash
# 1. 克隆项目
git clone <repository-url>
cd nginx-proxy

# 2. 构建 OpenResty 镜像
docker build -t nginx-proxy-openresty .

# 3. 启动服务
docker run -d \
  --name nginx-proxy \
  -p 80:80 \
  -p 443:443 \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/nginx-conf:/etc/nginx/conf.d \
  -v $(pwd)/nginx-certs:/etc/nginx/certs \
  nginx-proxy-openresty

# 4. 访问管理界面
open http://localhost:8080
```

### 本地开发

```bash
# 1. 安装依赖
make deps

# 2. 构建 Go 服务
make build

# 3. 启动 Go 服务（OpenResty 模式）
./nginx-proxy -config=config.json

# 4. 测试路由 API
chmod +x test_route_api.sh
./test_route_api.sh
```

### 验证部署

```bash
# 检查服务状态
curl http://localhost:8080/api/health

# 测试路由功能
curl -X POST http://localhost:8080/api/route \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/api",
    "remote_addr": "192.168.1.100",
    "headers": {"tt": "t", "x-env": "test", "x-token": "123"},
    "upstreams": [
      {
        "target": "http://21.91.124.161:8080",
        "condition_ip": "",
        "headers": {"tt": "t", "x-env": "test", "x-token": "123"}
      }
    ]
  }'
```

### Web 管理界面

访问 `http://localhost:8080` 使用 Web 界面管理：

- 📊 **仪表板**：系统概览和路由统计
- ⚙️ **路由配置**：管理复杂路由规则和头部条件
- 🔒 **证书管理**：SSL 证书上传和管理
- 📋 **日志查看**：实时查看路由匹配日志

## 📡 API 接口

### 🎯 核心路由接口

#### `POST /api/route` - 智能路由判断

OpenResty 调用此接口进行动态路由决策：

```bash
curl -X POST http://localhost:8080/api/route \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/api",
    "remote_addr": "192.168.1.100",
    "headers": {
      "tt": "t",
      "x-env": "test",
      "x-token": "123"
    },
    "upstreams": [
      {
        "target": "http://21.91.124.161:8080",
        "condition_ip": "192.168.1.0/24",
        "headers": {
          "tt": "t",
          "x-env": "test",
          "x-token": "123"
        }
      },
      {
        "target": "http://default-backend:8080",
        "condition_ip": "",
        "headers": {}
      }
    ]
  }'
```

**响应示例**：

```json
{
  "target": "http://21.91.124.161:8080"
}
```

### 🛠️ 管理接口

#### 规则管理

- `GET /api/rules` - 获取所有代理规则
- `GET /api/rules/{id}` - 获取指定规则详情
- `POST /api/rules` - 创建新的代理规则
- `PUT /api/rules/{id}` - 更新现有规则
- `DELETE /api/rules/{id}` - 删除规则

#### 系统管理

- `POST /api/reload` - 重载 OpenResty 配置
- `GET /api/health` - 系统健康检查

#### 证书管理

- `GET /api/certificates` - 获取所有 SSL 证书
- `POST /api/certificates` - 上传新证书（自动解析证书信息）
- `DELETE /api/certificates/{id}` - 删除证书

#### 腾讯云证书管理

- `GET /api/certificates/tencent/list` - 获取腾讯云证书列表
- `POST /api/certificates/tencent/apply` - 申请腾讯云证书
- `GET /api/certificates/tencent/{id}/status` - 检查证书状态
- `POST /api/certificates/tencent/{id}/renew` - 续期证书
- `POST /api/certificates/tencent/batch-renew` - 批量续期证书
- `DELETE /api/certificates/tencent/{id}` - 删除腾讯云证书（同步删除云端）

### 📋 配置示例

#### 复杂路由规则配置

```json
{
  "server_name": "api.example.com",
  "listen_ports": [80, 443],
  "ssl_cert": "/etc/nginx/certs/example.com.crt",
  "ssl_key": "/etc/nginx/certs/example.com.key",
  "locations": [
    {
      "path": "/api/v1",
      "upstreams": [
        {
          "condition_ip": "192.168.1.0/24",
          "target": "http://internal-api:8080",
          "headers": {
            "x-env": "internal",
            "x-version": "v1"
          }
        },
        {
          "condition_ip": "",
          "target": "http://public-api:8080",
          "headers": {
            "x-env": "production",
            "x-version": "v1"
          }
        },
        {
          "condition_ip": "",
          "target": "http://default-api:8080",
          "headers": {}
        }
      ]
    }
  ]
}
```

#### 头部条件匹配示例

支持多个头部条件的"且"关系匹配：

```json
{
  "headers": {
    "tt": "t",           // 必须同时满足
    "x-env": "test",     // 所有这些条件
    "x-token": "123"     // 才会路由到目标服务器
  }
}
```

## ⚙️ 配置文件

### OpenResty 模式配置 (config.json)

```json
{
  "port": "8080",
  "nginx_path": "/usr/local/openresty/bin/openresty",
  "config_dir": "/etc/nginx/conf.d",
  "cert_dir": "/etc/nginx/certs",
  "database_path": "./nginx-proxy.db",
  "template_dir": "./template",
  "tencent_cloud": {
    "secret_id": "your-tencent-secret-id",
    "secret_key": "your-tencent-secret-key",
    "region": "ap-beijing"
  }
}
```

### 标准模式配置 (config.json)

```json
{
  "port": "8080",
  "nginx_path": "/usr/sbin/nginx",
  "config_dir": "/etc/nginx/conf.d",
  "cert_dir": "/etc/nginx/certs",
  "database_path": "./nginx-proxy.db",
  "template_dir": "./template",
  "tencent_cloud": {
    "secret_id": "your-tencent-secret-id",
    "secret_key": "your-tencent-secret-key",
    "region": "ap-beijing"
  }
}
```

## 🔒 腾讯云证书管理详细指南

### 配置腾讯云API密钥

1. **获取API密钥**
    - 登录腾讯云控制台
    - 访问 [API密钥管理](https://console.cloud.tencent.com/cam/capi)
    - 创建或查看现有的SecretId和SecretKey

2. **配置文件设置**

```json
{
  "tencent_cloud": {
    "secret_id": "AKIDxxxxxxxxxxxxxxxxxxxxx",
    "secret_key": "xxxxxxxxxxxxxxxxxxxxxxxx",
    "region": "ap-beijing"
  }
}
```

### 证书申请流程

#### 1. 通过Web界面申请

1. 访问管理界面 `http://localhost:8080`
2. 点击"证书管理" → "申请腾讯云证书"
3. 填写域名和证书别名
4. 选择验证方式：
    - **DNS自动验证**（推荐）：系统自动完成DNS验证
    - **DNS手动验证**：需要手动添加DNS记录
    - **文件验证**：需要在网站根目录放置验证文件

#### 2. 通过API申请

```bash
curl -X POST http://localhost:8080/api/certificates/tencent/apply \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "example.com",
    "cert_alias": "example-cert",
    "validate_type": "DNS_AUTO"
  }'
```

### 证书状态说明

| 状态       | 说明         | 操作建议   |
|----------|------------|--------|
| **申请中**  | 证书正在申请或验证中 | 等待验证完成 |
| **正常**   | 证书有效且未过期   | 可正常使用  |
| **即将过期** | 证书将在30天内过期 | 建议续期   |
| **已过期**  | 证书已过期      | 需要立即续期 |

### 自动续期配置

系统会自动检查证书状态并在到期前续期：

```bash
# 手动触发批量续期
curl -X POST http://localhost:8080/api/certificates/tencent/batch-renew

# 检查特定证书状态
curl http://localhost:8080/api/certificates/tencent/123/status
```

### 证书使用示例

#### 在代理配置中使用腾讯云证书

```json
{
  "server_name": "api.example.com",
  "listen_ports": [80, 443],
  "ssl_cert": "/etc/nginx/certs/tencent_123_cert.pem",
  "ssl_key": "/etc/nginx/certs/tencent_123_key.pem",
  "locations": [...]
}
```

#### SSL配置选项

- **仅HTTPS模式**：只监听443端口，拒绝HTTP请求
- **HTTPS+重定向模式**：监听80和443端口，HTTP自动重定向到HTTPS

### 故障排除

#### 常见问题

1. **证书申请失败**

```bash
# 检查API密钥配置
curl http://localhost:8080/api/certificates/tencent/list

# 查看详细错误日志
docker logs nginx-proxy | grep "Tencent"
```

2. **DNS验证失败**

- 确保域名DNS解析正确
- 检查域名是否已备案（中国大陆）
- 验证DNS记录是否生效

3. **证书续期失败**

```bash
# 手动触发续期
curl -X POST http://localhost:8080/api/certificates/tencent/123/renew

# 检查证书状态
curl http://localhost:8080/api/certificates/tencent/123/status
```

#### 调试工具

访问调试页面：`http://localhost:8080/debug_certificates.html`

- 测试腾讯云API连接
- 查看证书数据格式
- 验证API响应状态

## 📁 项目结构

```
nginx-proxy/
├── cmd/server/                    # 应用入口
│   └── main.go                   # 主程序入口
├── internal/
│   ├── api/                      # API 处理器
│   │   ├── handlers.go           # 基础路由接口实现
│   │   ├── certificate.go        # 证书管理API
│   │   └── tencent_ssl.go        # 腾讯云证书API
│   ├── core/                     # 核心逻辑
│   │   ├── generator.go          # 配置生成器
│   │   ├── nginx.go              # OpenResty 管理
│   │   ├── certificate.go        # 证书核心逻辑
│   │   └── tencent_ssl.go        # 腾讯云SSL服务
│   └── db/                       # 数据库模型
│       └── models.go             # 数据模型定义
├── web/static/                   # 前端静态文件
│   ├── index.html                # 主管理界面
│   ├── debug_certificates.html   # 证书调试页面
│   └── js/app.js                 # 前端JavaScript逻辑
├── template/
│   └── nginx.conf.tpl            # OpenResty 配置模板
├── examples/                     # 配置示例
│   └── openresty-routing-example.json
├── certs/                        # SSL证书存储目录
├── config.json                   # 配置文件
├── go.mod                        # Go模块依赖
├── Dockerfile                    # OpenResty Docker 构建
├── test_route_api.sh            # API 测试脚本
└── README.md                     # 项目文档
```

## 🚀 生产环境部署

### 1. 环境准备

```bash
# 创建必要目录
mkdir -p data nginx-conf nginx-certs logs config template

# 设置权限
chmod 755 data nginx-conf nginx-certs logs
```

### 2. Docker 部署

```bash
# 构建镜像
docker build -t nginx-proxy-openresty .

# 启动服务
docker run -d \
  --name nginx-proxy \
  --restart unless-stopped \
  -p 80:80 \
  -p 443:443 \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/nginx-conf:/etc/nginx/conf.d \
  -v $(pwd)/nginx-certs:/etc/nginx/certs \
  -v $(pwd)/logs:/var/log/nginx \
  -v $(pwd)/config.json:/app/config/config.json \
  nginx-proxy-openresty
```

### 3. 健康检查

```bash
# 检查服务状态
curl http://localhost:8080/api/health

# 检查容器状态
docker ps | grep nginx-proxy

# 查看应用日志
docker logs nginx-proxy

# 查看 OpenResty 日志
docker exec nginx-proxy tail -f /var/log/nginx/access.log
docker exec nginx-proxy tail -f /var/log/nginx/error.log
```

### 4. 监控和维护

```bash
# 查看路由匹配日志
docker logs nginx-proxy | grep "Route"

# 重载配置
curl -X POST http://localhost:8080/api/reload

# 备份数据库
docker exec nginx-proxy cp /app/data/nginx-proxy.db /app/data/nginx-proxy.db.backup
```

## 🛠️ 开发指南

### 本地开发环境

```bash
# 1. 安装 Go 1.21+
go version

# 2. 克隆项目
git clone <repository-url>
cd nginx-proxy

# 3. 安装依赖
go mod download

# 4. 构建项目
make build

# 5. 启动开发服务
./nginx-proxy -config=config.json
```

### 测试和验证

```bash
# 运行 API 测试
chmod +x test_route_api.sh
./test_route_api.sh

# 代码格式化
go fmt ./...

# 代码检查
go vet ./...

# 运行单元测试
go test ./...
```

### 调试路由逻辑

```bash
# 启用详细日志
export GIN_MODE=debug

# 测试特定路由条件
curl -X POST http://localhost:8080/api/route \
  -H "Content-Type: application/json" \
  -d @examples/openresty-routing-example.json
```

## 🔧 故障排除

### 常见问题

#### 1. **OpenResty 相关问题**

```bash
# 检查 OpenResty 是否正确安装
docker exec nginx-proxy /usr/local/openresty/bin/openresty -v

# 检查 Lua 模块
docker exec nginx-proxy /usr/local/openresty/luajit/bin/luarocks list

# 测试 OpenResty 配置
docker exec nginx-proxy /usr/local/openresty/bin/openresty -t
```

#### 2. **路由匹配问题**

```bash
# 查看路由匹配日志
docker logs nginx-proxy | grep "Route"

# 检查头部条件匹配
docker logs nginx-proxy | grep "Header"

# 验证 IP 匹配逻辑
docker logs nginx-proxy | grep "IP"
```

#### 3. **API 连接问题**

```bash
# 检查 Go 服务是否运行在 8080 端口
netstat -tlnp | grep 8080

# 测试 API 连通性
curl -v http://localhost:8080/api/health

# 检查防火墙设置
iptables -L | grep 8080
```

#### 4. **配置生成问题**

```bash
# 检查生成的配置文件
ls -la nginx-conf/

# 查看配置文件内容
cat nginx-conf/*.conf

# 验证模板文件
cat template/nginx.conf.tpl
```

### 日志分析

```bash
# 实时查看所有日志
docker logs -f nginx-proxy

# 过滤路由相关日志
docker logs nginx-proxy 2>&1 | grep -E "(Route|Header|IP)"

# 查看 OpenResty 访问日志
docker exec nginx-proxy tail -f /var/log/nginx/access.log

# 查看 OpenResty 错误日志
docker exec nginx-proxy tail -f /var/log/nginx/error.log

# 查看 Lua 脚本错误
docker logs nginx-proxy 2>&1 | grep -i lua
```

### 性能优化

```bash
# 监控 API 响应时间
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8080/api/route

# 查看内存使用情况
docker stats nginx-proxy

# 分析路由匹配性能
docker logs nginx-proxy | grep "Route matched" | wc -l
```

### 配置验证

```bash
# 验证 JSON 配置格式
cat config.json | jq .

# 检查配置文件权限
ls -la config.json

# 验证证书文件
openssl x509 -in nginx-certs/cert.pem -text -noout
```

## 🤝 贡献指南

我们欢迎所有形式的贡献！

### 开发流程

1. **Fork 项目** 并克隆到本地
2. **创建功能分支**: `git checkout -b feature/amazing-feature`
3. **提交更改**: `git commit -m 'Add amazing feature'`
4. **推送分支**: `git push origin feature/amazing-feature`
5. **创建 Pull Request**

### 代码规范

```bash
# 代码格式化
go fmt ./...

# 代码检查
go vet ./...

# 运行测试
go test ./...

# 检查 API 功能
./test_route_api.sh
```

### 提交规范

- `feat:` 新功能
- `fix:` 修复 bug
- `docs:` 文档更新
- `style:` 代码格式调整
- `refactor:` 代码重构
- `test:` 测试相关
- `chore:` 构建过程或辅助工具的变动

## 📚 相关文档

- [OpenResty 路由解决方案](README_OPENRESTY_SOLUTION.md) - 详细的架构说明
- [代码 Review 修复记录](CODE_REVIEW_FIXES.md) - 代码质量改进记录
- [API 测试脚本](test_route_api.sh) - 完整的 API 测试用例

## 🔗 相关链接

- [OpenResty 官方文档](https://openresty.org/)
- [Lua Resty HTTP](https://github.com/ledgetech/lua-resty-http)
- [Gin Web Framework](https://gin-gonic.com/)
- [GORM 文档](https://gorm.io/)

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📋 更新日志

### v3.0.0 (最新版本)

- ✨ **新增腾讯云SSL证书集成**
    - 支持一键申请腾讯云免费SSL证书
    - 自动DNS验证和证书下载
    - 证书状态实时监控和管理
- ✨ **智能证书生命周期管理**
    - 证书到期前自动续期
    - 删除证书时同步清理腾讯云端证书
    - 批量证书续期功能
- ✨ **增强的SSL配置选项**
    - 支持仅HTTPS模式（高安全）
    - 支持HTTPS+HTTP重定向模式
    - 灵活的SSL端口配置
- 🐛 **前端功能修复**
    - 修复添加代理功能无响应问题
    - 修复证书选择下拉框数据加载
    - 修复JavaScript运行时错误
- 🔧 **用户体验优化**
    - 改进证书管理界面
    - 添加证书调试工具页面
    - 优化异步数据加载和错误处理

### v2.0.0

- ✨ 智能路由系统重构
- ✨ OpenResty + Lua 架构升级
- ✨ 复杂头部条件匹配支持
- 🔧 Docker 容器化部署

### v1.0.0

- 🎉 初始版本发布
- ✨ 基础反向代理功能
- ✨ Web 管理界面
- ✨ SQLite 数据存储

## 💬 支持与反馈

- 🐛 **Bug 报告**: [提交 Issue](../../issues/new?template=bug_report.md)
- 💡 **功能建议**: [提交 Feature Request](../../issues/new?template=feature_request.md)
- 📖 **文档问题**: [提交文档 Issue](../../issues/new?template=documentation.md)
- 💬 **讨论交流**: [GitHub Discussions](../../discussions)

## 🌟 致谢

感谢所有贡献者对项目的支持！特别感谢：

- 腾讯云团队提供的免费SSL证书服务
- OpenResty 社区的技术支持
- 所有提交Bug报告和功能建议的用户

---

**⭐ 如果这个项目对您有帮助，请给我们一个 Star！**

**🚀 立即体验腾讯云证书自动管理功能，让SSL证书管理变得简单高效！**