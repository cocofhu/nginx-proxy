# DNS 解析和日志问题故障排除指南

## 🚨 问题描述

您遇到的两个主要问题：

1. **DNS 解析错误**：`no resolver defined to resolve git.service.arpa`
2. **日志读取卡顿**：`tail -f /var/log/nginx/access.log` 命令卡住

## 🔍 问题分析

### DNS 解析问题

- Nginx 默认不包含 DNS 解析器配置
- 内网域名 `git.service.arpa` 需要特定的 DNS 服务器才能解析
- Docker 环境中需要使用 Docker 内置 DNS (127.0.0.11)

### 日志卡顿问题

- 默认日志配置没有缓冲，每次写入都直接刷盘
- 高频访问时会导致 I/O 阻塞
- 日志文件可能存在权限问题

## 🛠️ 解决方案

### 方案 1：立即修复（推荐）

1. **使用修复脚本**：

```bash
# 运行自动修复脚本
chmod +x scripts/fix-dns-and-logs.sh
./scripts/fix-dns-and-logs.sh
```

2. **手动应用 Docker 配置**：

```bash
# 备份当前配置
cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.backup

# 应用 Docker 优化配置
cp nginx-docker.conf /etc/nginx/nginx.conf

# 测试配置
nginx -t

# 重载配置
nginx -s reload
```

### 方案 2：DNS 解析的三种解决方式

#### 选项 A：使用 IP 地址（最简单）

```json
{
  "condition_ip": "192.168.2.45/32",
  "target": "http://192.168.2.100:3000"
}
```

#### 选项 B：配置 DNS 解析器

在 `nginx.conf` 的 `http` 块中添加：

```nginx
# Docker 环境
resolver 127.0.0.11 8.8.8.8 valid=300s;

# 宿主机环境
resolver 192.168.2.1 8.8.8.8 valid=300s;
```

#### 选项 C：使用 hosts 文件

```bash
# 在容器中添加 hosts 映射
echo "192.168.2.100 git.service.arpa" >> /etc/hosts
```

### 方案 3：日志优化配置

在 `nginx.conf` 中优化日志设置：

```nginx
# 使用缓冲的访问日志
access_log /var/log/nginx/access.log main buffer=64k flush=1s;

# 异步错误日志
error_log /var/log/nginx/error.log warn;
```

## 🧪 测试验证

### 1. 测试 DNS 解析

```bash
# 在 Nginx 容器中测试
docker exec your-nginx-container nslookup git.service.arpa

# 或使用 dig
docker exec your-nginx-container dig git.service.arpa
```

### 2. 测试分流配置

```bash
# 从指定 IP 测试（如果可能）
curl -H "Host: fff.com" http://your-server/

# 查看 Nginx 错误日志
docker exec your-nginx-container tail -f /var/log/nginx/error.log
```

### 3. 测试日志功能

```bash
# 实时查看访问日志（应该不再卡顿）
docker exec your-nginx-container tail -f /var/log/nginx/access.log

# 检查日志文件权限
docker exec your-nginx-container ls -la /var/log/nginx/
```

## 📋 完整的配置示例

### 使用 IP 地址的配置（推荐）

```bash
curl -X POST http://localhost:8080/api/rules \
  -H "Content-Type: application/json" \
  -d '{
    "server_name": "fff.com",
    "listen_ports": [80],
    "locations": [{
      "path": "/",
      "upstreams": [
        {
          "condition_ip": "192.168.2.45/32",
          "target": "http://192.168.2.100:3000"
        },
        {
          "condition_ip": "0.0.0.0/0",
          "target": "http://192.168.2.1"
        }
      ]
    }]
  }'
```

### 使用域名的配置（需要 DNS 解析）

```bash
curl -X POST http://localhost:8080/api/rules \
  -H "Content-Type: application/json" \
  -d @examples/internal-dns-config.json
```

## 🔧 Docker 环境特殊配置

### Dockerfile 优化

```dockerfile
# 确保日志目录存在
RUN mkdir -p /var/log/nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx /var/cache/nginx

# 复制优化的配置文件
COPY nginx-docker.conf /etc/nginx/nginx.conf
```

### docker-compose.yml 配置

```yaml
services:
  nginx:
    image: nginx:alpine
    volumes:
      - ./nginx-docker.conf:/etc/nginx/nginx.conf
      - ./logs:/var/log/nginx
    networks:
      - internal
    dns:
      - 8.8.8.8
      - 8.8.4.4
```

## 🚨 常见错误和解决方案

### 错误 1：`no resolver defined`

**解决**：在 `nginx.conf` 中添加 `resolver` 指令

### 错误 2：日志文件权限错误

**解决**：

```bash
chown -R nginx:nginx /var/log/nginx
chmod 755 /var/log/nginx
```

### 错误 3：DNS 解析超时

**解决**：

```nginx
resolver 127.0.0.11 valid=300s;
resolver_timeout 10s;
```

### 错误 4：日志读取卡顿

**解决**：启用日志缓冲

```nginx
access_log /var/log/nginx/access.log main buffer=64k flush=1s;
```

## 📞 快速诊断命令

```bash
# 检查 DNS 配置
grep -n resolver /etc/nginx/nginx.conf

# 检查日志配置
grep -n access_log /etc/nginx/nginx.conf

# 测试 Nginx 配置
nginx -t

# 查看 Nginx 进程
ps aux | grep nginx

# 检查日志文件
ls -la /var/log/nginx/

# 测试域名解析
nslookup git.service.arpa
```

## 🎯 推荐的最终配置

基于您的需求，推荐使用以下配置：

1. **使用 IP 地址替代域名**（避免 DNS 问题）
2. **启用日志缓冲**（解决卡顿问题）
3. **添加健康检查端点**（便于监控）

这样可以确保系统稳定运行，同时避免 DNS 解析的复杂性。