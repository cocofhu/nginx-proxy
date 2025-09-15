# 轻量级 Nginx 配置管理器

一个用 Go 编写的轻量级工具，通过 REST API 管理 Nginx 反向代理规则，支持基于来源 IP 的分流、泛域名证书管理和自动配置生成。

## 功能特性

- 🚀 **REST API 管理**：完整的 CRUD 操作管理反向代理规则
- 🌐 **智能分流**：基于来源 IP 段的流量分发
- 🔒 **证书管理**：支持上传和管理 SSL 证书
- 📝 **自动配置**：自动生成和验证 Nginx 配置文件
- 🔄 **热重载**：配置变更后自动重载 Nginx
- 💾 **持久化存储**：使用 SQLite 数据库存储配置
- 🐳 **容器化**：提供 Docker 支持

## 快速开始

### 本地运行

1. **克隆项目**
```bash
git clone <repository-url>
cd nginx-proxy
```

2. **安装依赖**
```bash
go mod tidy
```

3. **配置文件**
复制并修改配置文件：
```bash
cp config.json.example config.json
```

4. **启动服务**
```bash
go run cmd/server/main.go
```

### Docker 运行

1. **构建镜像**
```bash
docker build -t nginx-proxy .
```

2. **运行容器**
```bash
docker run -d \
  --name nginx-proxy \
  -p 8080:8080 \
  -v /etc/nginx/conf.d:/etc/nginx/conf.d \
  -v /etc/nginx/certs:/etc/nginx/certs \
  -v $(pwd)/data:/app/data \
  nginx-proxy
```

## API 文档

### 规则管理

#### 获取所有规则
```http
GET /api/rules
```

#### 获取单个规则
```http
GET /api/rules/{id}
```

#### 创建规则
```http
POST /api/rules
Content-Type: application/json

{
  "server_name": "example.com",
  "listen_ports": [443, 8443],
  "ssl_cert": "/etc/nginx/certs/example.crt",
  "ssl_key": "/etc/nginx/certs/example.key",
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
          "target": "http://external-server:8080"
        }
      ]
    }
  ]
}
```

#### 更新规则
```http
PUT /api/rules/{id}
Content-Type: application/json

{
  "server_name": "example.com",
  "listen_ports": [443],
  "ssl_cert": "/etc/nginx/certs/example.crt",
  "ssl_key": "/etc/nginx/certs/example.key",
  "locations": [...]
}
```

#### 删除规则
```http
DELETE /api/rules/{id}
```

### 系统管理

#### 手动重载 Nginx
```http
POST /api/reload
```

#### 上传证书
```http
POST /api/certificates
Content-Type: multipart/form-data

cert: <certificate-file>
key: <private-key-file>
```

## 配置说明

### config.json 配置文件

```json
{
  "port": "8080",                    // API 服务端口
  "nginx_path": "/usr/sbin/nginx",   // Nginx 可执行文件路径
  "config_dir": "/etc/nginx/conf.d", // Nginx 配置文件目录
  "cert_dir": "/etc/nginx/certs",    // SSL 证书存储目录
  "database_path": "./nginx-proxy.db", // SQLite 数据库文件路径
  "template_dir": "./template"       // 模板文件目录
}
```

### 规则字段说明

- **server_name**: 域名（支持泛域名）
- **listen_ports**: 监听端口列表
- **ssl_cert**: SSL 证书文件路径
- **ssl_key**: SSL 私钥文件路径
- **locations**: 路径配置列表
  - **path**: 匹配路径
  - **upstreams**: 上游服务器列表
    - **condition_ip**: IP 条件（CIDR 格式，0.0.0.0/0 表示所有）
    - **target**: 目标服务器地址

## IP 分流示例

系统支持基于来源 IP 的智能分流：

```json
{
  "server_name": "api.example.com",
  "listen_ports": [443],
  "ssl_cert": "/etc/nginx/certs/wildcard.crt",
  "ssl_key": "/etc/nginx/certs/wildcard.key",
  "locations": [
    {
      "path": "/api/",
      "upstreams": [
        {
          "condition_ip": "10.0.0.0/8",
          "target": "http://internal-api:8080"
        },
        {
          "condition_ip": "192.168.0.0/16",
          "target": "http://internal-api:8080"
        },
        {
          "condition_ip": "0.0.0.0/0",
          "target": "http://public-api:8080"
        }
      ]
    }
  ]
}
```

这个配置会：
- 内网 IP（10.x.x.x 和 192.168.x.x）访问内部 API 服务器
- 其他所有 IP 访问公共 API 服务器

## 生成的 Nginx 配置示例

```nginx
server {
    listen 443 ssl;
    
    server_name api.example.com;
    
    ssl_certificate     /etc/nginx/certs/wildcard.crt;
    ssl_certificate_key /etc/nginx/certs/wildcard.key;
    
    # SSL 优化配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;
    
    location /api/ {
        # IP 分流配置
        geo $remote_addr $is_internal {
            default 0;
            10.0.0.0/8 1;
            192.168.0.0/16 1;
        }
        
        map $is_internal $backend {
            1 "http://internal-api:8080";
            0 "http://public-api:8080";
        }
        
        proxy_pass $backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 安全特性

- ✅ 配置文件语法验证（nginx -t）
- ✅ 配置失败自动回滚
- ✅ 操作日志记录
- ✅ SSL/TLS 安全配置
- ✅ 安全响应头设置

## 故障排除

### 常见问题

1. **Nginx 配置测试失败**
   - 检查 nginx 可执行文件路径
   - 验证证书文件是否存在且可读
   - 检查配置目录权限

2. **证书上传失败**
   - 确保证书目录存在且可写
   - 检查证书文件格式是否正确

3. **数据库连接失败**
   - 检查数据库文件路径和权限
   - 确保 SQLite 支持已启用

### 日志查看

```bash
# 查看应用日志
docker logs nginx-proxy

# 查看 Nginx 错误日志
tail -f /var/log/nginx/error.log
```

## 开发

### 项目结构

```
├── cmd/server/main.go          # 程序入口
├── internal/
│   ├── api/handlers.go         # API 处理器
│   ├── core/
│   │   ├── generator.go        # 配置生成器
│   │   └── nginx.go           # Nginx 管理器
│   └── db/
│       ├── db.go              # 数据库初始化
│       └── models.go          # 数据模型
├── template/nginx.conf.tpl     # Nginx 配置模板
├── config.json                 # 配置文件
├── Dockerfile                  # Docker 构建文件
└── README.md                   # 项目文档
```

### 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 支持

如果您遇到问题或有功能建议，请：

1. 查看 [Issues](../../issues) 页面
2. 创建新的 Issue 描述问题
3. 提供详细的错误信息和环境信息

---

**注意**: 在生产环境中使用前，请确保：
- 正确配置防火墙规则
- 定期备份数据库文件
- 监控 Nginx 和应用程序日志
- 使用有效的 SSL 证书