# 🚀 快速修复指南

## 问题总结

1. ❌ DNS 解析错误：`no resolver defined to resolve git.service.arpa`
2. ❌ 日志读取卡顿：`tail -f /var/log/nginx/access.log` 卡住

## ⚡ 立即修复（3 分钟解决）

### 方法 1：重新构建 Docker 镜像（推荐）

```bash
# 1. 停止当前容器
docker stop f85b35c77dce

# 2. 重新构建镜像（已包含 DNS 和日志优化）
docker-compose build

# 3. 启动新容器
docker-compose up -d

# 4. 创建使用 IP 地址的分流规则
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

# 5. 重载配置
curl -X POST http://localhost:8080/api/reload
```

### 方法 2：手动修复现有容器

```bash
# 1. 复制优化配置到容器
docker cp nginx-docker.conf f85b35c77dce:/etc/nginx/nginx.conf

# 2. 重启 Nginx
docker exec f85b35c77dce nginx -s reload

# 3. 测试日志（应该不再卡顿）
docker exec f85b35c77dce tail -f /var/log/nginx/access.log
```

## ✅ 验证修复效果

### 1. 测试 DNS 解析

```bash
# 检查 DNS 配置
docker exec nginx-proxy grep resolver /etc/nginx/nginx.conf

# 应该看到：
# resolver 127.0.0.11 8.8.8.8 8.8.4.4 valid=300s ipv6=off;
```

### 2. 测试日志功能

```bash
# 测试日志读取（应该不再卡顿）
docker exec nginx-proxy tail -f /var/log/nginx/access.log

# 检查日志配置
docker exec nginx-proxy grep access_log /etc/nginx/nginx.conf

# 应该看到：
# access_log /var/log/nginx/access.log main buffer=64k flush=1s;
```

### 3. 测试分流效果

```bash
# 查看生成的配置
docker exec nginx-proxy cat /etc/nginx/conf.d/*.conf

# 应该看到正确的 IP 匹配：
# if ($remote_addr = "192.168.2.45") {
#     set $backend "http://192.168.2.100:3000";
# }
```

## 🔧 关键修改说明

### DNS 解析修复

- ✅ 添加了 Docker 内置 DNS (127.0.0.11)
- ✅ 添加了公共 DNS 备用
- ✅ 使用变量方式进行动态域名解析
- ✅ 建议使用 IP 地址避免 DNS 问题

### 日志卡顿修复

- ✅ 启用日志缓冲：`buffer=64k flush=1s`
- ✅ 异步日志写入
- ✅ 优化日志格式，添加性能监控

### IP 分流修复

- ✅ 单个 IP 使用精确匹配：`if ($remote_addr = "192.168.2.45")`
- ✅ IP 段使用优化正则表达式
- ✅ 自动转义特殊字符

## 📞 如果还有问题

### 检查容器状态

```bash
docker-compose ps
docker-compose logs nginx-proxy
```

### 进入容器调试

```bash
docker-compose exec nginx-proxy sh
nginx -t
ps aux | grep nginx
```

### 重置所有配置

```bash
docker-compose down
docker-compose up -d --build
```

## 🎯 推荐配置

**使用 IP 地址替代域名**（最稳定）：

```json
{
  "condition_ip": "192.168.2.45/32",
  "target": "http://192.168.2.100:3000"
}
```

这样可以完全避免 DNS 解析问题，确保分流功能稳定工作。