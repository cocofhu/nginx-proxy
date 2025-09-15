# 分流配置示例

本文档展示了如何使用 Nginx Proxy 管理器配置基于来源 IP 的流量分流。

## 基本分流配置

### 示例 1：内外网分流

```json
{
  "server_name": "api.example.com",
  "listen_ports": [80, 443],
  "ssl_cert": "/etc/nginx/certs/example.com.crt",
  "ssl_key": "/etc/nginx/certs/example.com.key",
  "locations": [
    {
      "path": "/",
      "upstreams": [
        {
          "condition_ip": "192.168.1.0/24",
          "target": "http://internal-api:8080"
        },
        {
          "condition_ip": "10.0.0.0/8",
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

### 示例 2：多地域分流

```json
{
  "server_name": "cdn.example.com",
  "listen_ports": [80, 443],
  "ssl_cert": "/etc/nginx/certs/example.com.crt",
  "ssl_key": "/etc/nginx/certs/example.com.key",
  "locations": [
    {
      "path": "/",
      "upstreams": [
        {
          "condition_ip": "192.168.1.0/24",
          "target": "http://asia-server:8080"
        },
        {
          "condition_ip": "252.5.94.0/24",
          "target": "http://europe-server:8080"
        },
        {
          "condition_ip": "0.0.0.0/0",
          "target": "http://global-server:8080"
        }
      ]
    }
  ]
}
```

### 示例 3：不同路径的分流

```json
{
  "server_name": "app.example.com",
  "listen_ports": [80, 443],
  "ssl_cert": "/etc/nginx/certs/example.com.crt",
  "ssl_key": "/etc/nginx/certs/example.com.key",
  "locations": [
    {
      "path": "/api/",
      "upstreams": [
        {
          "condition_ip": "192.168.0.0/16",
          "target": "http://internal-api:3000"
        },
        {
          "condition_ip": "0.0.0.0/0",
          "target": "http://public-api:3000"
        }
      ]
    },
    {
      "path": "/static/",
      "upstreams": [
        {
          "condition_ip": "0.0.0.0/0",
          "target": "http://cdn-server:8080"
        }
      ]
    },
    {
      "path": "/",
      "upstreams": [
        {
          "condition_ip": "0.0.0.0/0",
          "target": "http://web-server:8080"
        }
      ]
    }
  ]
}
```

## 分流规则说明

### IP 条件格式

- **单个 IP**: `192.168.1.100/32`
- **子网**: `192.168.1.0/24`
- **大网段**: `10.0.0.0/8`
- **所有 IP**: `0.0.0.0/0` (通常作为默认规则)

### 匹配优先级

分流规则按照在 `upstreams` 数组中的顺序进行匹配，第一个匹配的规则将被使用。因此：

1. 将更具体的 IP 范围放在前面
2. 将默认规则 (`0.0.0.0/0`) 放在最后

### 最佳实践

1. **总是包含默认规则**: 确保有一个 `0.0.0.0/0` 的规则作为兜底
2. **从小到大排序**: 将更小的 IP 范围放在更大的范围之前
3. **测试配置**: 使用 `POST /api/reload` 测试配置是否正确
4. **监控日志**: 通过 Nginx 日志监控分流效果

## 使用 Web 界面配置

1. 访问管理界面：`http://your-server:8080`
2. 点击"添加代理"
3. 填写域名和路径
4. 在"分流配置"部分添加多个规则：
   - 来源IP：输入 CIDR 格式的 IP 范围
   - 目标地址：输入后端服务器地址
5. 点击"+ 添加分流规则"可以添加更多规则
6. 点击"添加"保存配置

## API 调用示例

```bash
# 创建分流规则
curl -X POST http://localhost:8080/api/rules \
  -H "Content-Type: application/json" \
  -d @traffic_split_config.json

# 查看所有规则
curl -X GET http://localhost:8080/api/rules

# 重载 Nginx 配置
curl -X POST http://localhost:8080/api/reload
```

## 故障排除

### 常见问题

1. **分流不生效**
   - 检查 IP 格式是否正确（必须是 CIDR 格式）
   - 确认规则顺序是否正确
   - 检查 Nginx 配置是否重载成功

2. **配置测试失败**
   - 验证后端服务器地址是否可达
   - 检查 SSL 证书路径是否正确
   - 确认端口是否被占用

3. **性能问题**
   - 避免过多的分流规则
   - 使用适当的 IP 范围大小
   - 监控后端服务器负载