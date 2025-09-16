# HTTP头部路由功能

## 功能说明

nginx-proxy现在支持基于HTTP请求头部的路由功能。您可以根据请求头中的key=value对来将请求路由到不同的后端服务。

## 配置方式

在upstream配置中添加`headers`字段：

```json
{
  "condition_ip": "",
  "target": "http://backend-v1:8080",
  "headers": {
    "X-API-Version": "v1",
    "X-Client-Type": "mobile"
  }
}
```

## 路由规则

1. **仅头部条件**: 当`condition_ip`为空且`headers`不为空时，仅根据头部进行路由
2. **仅IP条件**: 当`headers`为空且`condition_ip`不为空时，仅根据IP进行路由  
3. **IP+头部条件**: 当两者都不为空时，需要同时满足IP和头部条件
4. **默认路由**: 当两者都为空时，作为默认后端

## 使用示例

### 示例1: API版本路由

```json
{
  "path": "/api/v1",
  "upstreams": [
    {
      "target": "http://backend-v1:8080",
      "headers": {
        "X-API-Version": "v1"
      }
    },
    {
      "target": "http://backend-v2:8080", 
      "headers": {
        "X-API-Version": "v2"
      }
    },
    {
      "target": "http://backend-default:8080"
    }
  ]
}
```

### 示例2: 客户端类型路由

```json
{
  "path": "/app",
  "upstreams": [
    {
      "target": "http://mobile-backend:8080",
      "headers": {
        "User-Agent": "Mobile"
      }
    },
    {
      "target": "http://web-backend:8080"
    }
  ]
}
```

### 示例3: 混合条件路由

```json
{
  "path": "/admin",
  "upstreams": [
    {
      "condition_ip": "192.168.1.0/24",
      "target": "http://internal-admin:8080",
      "headers": {
        "X-Admin-Token": "secret123"
      }
    },
    {
      "target": "http://public-admin:8080"
    }
  ]
}
```

## 注意事项

1. 头部名称会自动转换为nginx变量格式（小写，横线转下划线）
2. 多个头部条件使用AND逻辑连接，必须全部匹配
3. 头部值支持精确匹配，特殊字符会自动转义
4. 建议总是设置一个默认后端（无条件的upstream）

## 生成的nginx配置示例

```nginx
location /api/v1 {
    set $backend "";
    
    # 头部条件: X-API-Version=v1
    if ($http_x_api_version = "v1") {
        set $backend "http://backend-v1:8080";
    }
    
    # 头部条件: X-API-Version=v2  
    if ($http_x_api_version = "v2") {
        set $backend "http://backend-v2:8080";
    }
    
    # 默认后端
    if ($backend = "") {
        set $backend "http://backend-default:8080";
    }
    
    set $upstream $backend;
    proxy_pass $upstream;
}