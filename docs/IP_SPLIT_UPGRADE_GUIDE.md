# IP 分流功能升级指南

## 🎯 功能概述

本次升级为 nginx-proxy 添加了智能 IP 分流功能，支持：

- ✅ 单个 IP 精确匹配（`192.168.2.45` 或 `192.168.2.45/32`）
- ✅ IP 段匹配（`192.168.1.0/24`、`10.0.0.0/16`、`172.16.0.0/8`）
- ✅ 自动优化匹配性能（单个 IP 使用精确匹配，IP 段使用正则匹配）
- ✅ 向后兼容现有配置

## 🔧 升级步骤

### 1. 备份现有配置
```bash
cp -r /etc/nginx/conf.d /etc/nginx/conf.d.backup
```

### 2. 重新编译项目
```bash
go build -o bin/nginx-proxy cmd/server/main.go
```

### 3. 重启服务
```bash
systemctl restart nginx-proxy
```

### 4. 测试新功能
```bash
# 使用提供的示例配置测试
curl -X POST http://localhost:8080/api/rules \
  -H "Content-Type: application/json" \
  -d @examples/single-ip-rule.json

# 重载 Nginx 配置
curl -X POST http://localhost:8080/api/reload

# 检查生成的配置
cat /etc/nginx/conf.d/*.conf
```

## 📋 配置示例

### 修复您当前的问题

**原问题配置**：
```json
{
  "condition_ip": "192.168.2.45/32",
  "target": "http://git.service.arpa"
}
```

**生成的 Nginx 配置**（修复后）：
```nginx
# 检查 IP 条件: 192.168.2.45/32
if ($remote_addr = "192.168.2.45") {
    set $backend "http://git.service.arpa";
}
```

### 支持的 IP 格式

| 输入格式 | 生成的 Nginx 条件 | 说明 |
|---------|------------------|------|
| `192.168.2.45/32` | `if ($remote_addr = "192.168.2.45")` | 单个 IP，精确匹配 |
| `192.168.2.45` | `if ($remote_addr = "192.168.2.45")` | 单个 IP，精确匹配 |
| `192.168.1.0/24` | `if ($remote_addr ~ "^192\.168\.1\.\d+$")` | C类子网 |
| `10.0.0.0/16` | `if ($remote_addr ~ "^10\.0\.\d+\.\d+$")` | B类子网 |
| `172.16.0.0/8` | `if ($remote_addr ~ "^172\.\d+\.\d+\.\d+$")` | A类子网 |
| `0.0.0.0/0` | `if ($backend = "")` | 默认路由 |

## 🧪 测试验证

### 1. 创建测试规则
```bash
# 测试单个 IP 分流
curl -X POST http://localhost:8080/api/rules \
  -H "Content-Type: application/json" \
  -d '{
    "server_name": "test.example.com",
    "listen_ports": [80],
    "locations": [{
      "path": "/",
      "upstreams": [
        {
          "condition_ip": "192.168.2.45/32",
          "target": "http://git.service.arpa"
        },
        {
          "condition_ip": "0.0.0.0/0",
          "target": "http://192.168.2.1"
        }
      ]
    }]
  }'
```

### 2. 验证生成的配置
```bash
# 查看生成的配置文件
ls /etc/nginx/conf.d/
cat /etc/nginx/conf.d/*.conf

# 测试 Nginx 配置语法
nginx -t
```

### 3. 重载并测试
```bash
# 重载 Nginx
curl -X POST http://localhost:8080/api/reload

# 测试分流效果
curl -H "Host: test.example.com" http://your-server/
```

## 🔍 故障排除

### 问题 1：配置不生效
**解决方案**：
1. 检查 Nginx 配置语法：`nginx -t`
2. 查看错误日志：`tail -f /var/log/nginx/error.log`
3. 确认配置已重载：`curl -X POST http://localhost:8080/api/reload`

### 问题 2：IP 匹配不正确
**解决方案**：
1. 检查生成的配置文件中的 IP 条件
2. 确认客户端 IP 是否正确（注意代理和负载均衡器的影响）
3. 查看访问日志确认 `$remote_addr` 的值

### 问题 3：性能问题
**解决方案**：
1. 单个 IP 会自动使用精确匹配（性能最佳）
2. 避免过多的分流规则
3. 将更具体的规则放在前面

## 📈 性能优化

新的智能匹配系统会根据 IP 格式自动选择最优的匹配方式：

- **精确匹配**：单个 IP 使用 `=` 操作符，性能最佳
- **正则匹配**：IP 段使用优化的正则表达式，平衡性能和功能
- **规则顺序**：更具体的规则自动优先匹配

## 🔄 回滚方案

如果遇到问题需要回滚：

1. 恢复备份的配置：
```bash
rm -rf /etc/nginx/conf.d
mv /etc/nginx/conf.d.backup /etc/nginx/conf.d
```

2. 重启 Nginx：
```bash
systemctl restart nginx
```

3. 使用旧版本的二进制文件

## 📞 技术支持

如果遇到问题，请提供：
1. 错误日志：`/var/log/nginx/error.log`
2. 生成的配置文件内容
3. 具体的 IP 分流需求
4. 测试步骤和预期结果