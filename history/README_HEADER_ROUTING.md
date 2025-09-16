# HTTP头部路由功能完整指南

## 功能概述

nginx-proxy现在支持基于HTTP请求头部的智能路由功能。您可以根据请求头中的key=value对将请求路由到不同的后端服务，实现更灵活的流量分发策略。

## 主要特性

- ✅ **头部条件路由**: 根据HTTP头部进行精确匹配路由
- ✅ **IP条件路由**: 基于客户端IP地址的路由（原有功能）
- ✅ **混合条件路由**: 同时满足IP和头部条件的复合路由
- ✅ **多头部支持**: 支持多个头部条件的AND逻辑组合
- ✅ **可视化管理**: 通过Web界面轻松配置和管理
- ✅ **实时生效**: 配置更改后自动重载nginx配置

## 快速开始

### 1. 启动服务

```bash
# 编译并启动服务
go build ./cmd/server
./server -config config.json
```

### 2. 访问管理界面

打开浏览器访问: `http://localhost:8080`

### 3. 创建头部路由规则

1. 点击"代理配置"页面
2. 点击"添加代理"按钮
3. 填写基本信息（域名、路径等）
4. 在"分流配置"中添加头部路由条件
5. 保存配置

## 配置示例

### 示例1: API版本路由

根据`X-API-Version`头部将请求路由到不同版本的API服务：

```json
{
  "server_name": "api.example.com",
  "listen_ports": [80],
  "locations": [{
    "path": "/api/v1",
    "upstreams": [
      {
        "target": "http://api-v1.backend:8080",
        "headers": {
          "X-API-Version": "v1"
        }
      },
      {
        "target": "http://api-v2.backend:8080", 
        "headers": {
          "X-API-Version": "v2"
        }
      },
      {
        "target": "http://api-default.backend:8080"
      }
    ]
  }]
}
```

**测试命令:**

```bash
# 路由到v1服务
curl -H "X-API-Version: v1" http://api.example.com/api/v1/users

# 路由到v2服务  
curl -H "X-API-Version: v2" http://api.example.com/api/v1/users

# 路由到默认服务
curl http://api.example.com/api/v1/users
```

### 示例2: 客户端类型路由

根据`User-Agent`头部区分移动端和桌面端：

```json
{
  "server_name": "app.example.com",
  "listen_ports": [80],
  "locations": [{
    "path": "/",
    "upstreams": [
      {
        "target": "http://mobile.backend:8080",
        "headers": {
          "User-Agent": "Mobile"
        }
      },
      {
        "target": "http://web.backend:8080"
      }
    ]
  }]
}
```

### 示例3: 混合条件路由

同时基于IP和头部进行路由：

```json
{
  "server_name": "admin.example.com",
  "listen_ports": [80],
  "locations": [{
    "path": "/admin",
    "upstreams": [
      {
        "condition_ip": "192.168.1.0/24",
        "target": "http://internal-admin.backend:8080",
        "headers": {
          "X-Admin-Token": "secret123"
        }
      },
      {
        "target": "http://public-admin.backend:8080"
      }
    ]
  }]
}
```

### 示例4: 多头部条件

多个头部条件必须同时满足：

```json
{
  "target": "http://special.backend:8080",
  "headers": {
    "X-API-Version": "v2",
    "X-Client-Type": "premium",
    "X-Feature-Flag": "enabled"
  }
}
```

## Web界面使用

### 添加头部路由

1. **基本配置**
    - 域名: 输入要代理的域名
    - 路径: 设置location路径（默认为/）
    - SSL: 可选择启用HTTPS

2. **分流配置**
    - 来源IP: 设置IP条件（可选，默认0.0.0.0/0）
    - 目标地址: 后端服务地址
    - HTTP头部路由: 点击"+"添加头部条件

3. **头部条件**
    - Header名称: 如`X-API-Version`、`User-Agent`等
    - Header值: 对应的匹配值
    - 支持添加多个头部条件

### 管理现有规则

- **查看**: 代理列表显示所有路由条件
- **编辑**: 点击"编辑"修改现有规则
- **删除**: 点击"删除"移除规则

## 生成的Nginx配置

头部路由会生成如下nginx配置：

```nginx
location /api/v1 {
    set $backend "";
    
    # 头部条件: X-API-Version=v1
    if ($http_x_api_version = "v1") {
        set $backend "http://api-v1.backend:8080";
    }
    
    # 头部条件: X-API-Version=v2  
    if ($http_x_api_version = "v2") {
        set $backend "http://api-v2.backend:8080";
    }
    
    # 默认后端
    if ($backend = "") {
        set $backend "http://api-default.backend:8080";
    }
    
    set $upstream $backend;
    proxy_pass $upstream;
    
    # 标准代理头部
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

## API接口

### 创建规则

```bash
POST /api/rules
Content-Type: application/json

{
  "server_name": "api.example.com",
  "listen_ports": [80],
  "locations": [{
    "path": "/api",
    "upstreams": [{
      "condition_ip": "",
      "target": "http://backend:8080",
      "headers": {
        "X-API-Version": "v1"
      }
    }]
  }]
}
```

### 更新规则

```bash
PUT /api/rules/{id}
Content-Type: application/json

{
  "server_name": "api.example.com",
  "listen_ports": [80],
  "locations": [{
    "path": "/api",
    "upstreams": [{
      "condition_ip": "",
      "target": "http://new-backend:8080",
      "headers": {
        "X-API-Version": "v2"
      }
    }]
  }]
}
```

## 测试工具

### 1. 使用测试页面

访问 `examples/test-frontend.html` 进行交互式测试。

### 2. 使用测试脚本

```bash
chmod +x examples/test-header-routing.sh
./examples/test-header-routing.sh
```

### 3. 手动测试

```bash
# 测试不同的头部值
curl -H "X-API-Version: v1" http://your-domain/api/test
curl -H "X-API-Version: v2" http://your-domain/api/test
curl -H "User-Agent: Mobile" http://your-domain/mobile/test
```

## 最佳实践

### 1. 路由优先级

- 具体条件优先于默认条件
- IP + 头部条件优先于单一条件
- 建议总是设置一个默认后端

### 2. 头部命名

- 使用标准HTTP头部名称
- 自定义头部建议使用`X-`前缀
- 头部名称会自动转换为nginx变量格式

### 3. 性能考虑

- 避免过多的条件判断
- 合理设置upstream数量
- 考虑使用nginx的geo模块处理复杂IP条件

### 4. 安全建议

- 不要在头部中传递敏感信息
- 验证头部值的合法性
- 考虑头部伪造的安全风险

## 故障排除

### 1. 配置不生效

- 检查nginx配置语法: `nginx -t`
- 查看nginx错误日志
- 确认规则保存成功

### 2. 路由不正确

- 验证头部名称和值的准确性
- 检查条件优先级
- 使用curl测试具体场景

### 3. 性能问题

- 监控nginx访问日志
- 检查后端服务响应时间
- 考虑启用nginx缓存

## 更新日志

- **v1.0.0**: 初始版本，支持基本头部路由
- **v1.1.0**: 添加多头部条件支持
- **v1.2.0**: 完善Web界面，添加可视化管理

## 技术支持

如有问题，请查看：

- 项目文档: `examples/HEADER_ROUTING.md`
- 配置示例: `examples/header-routing-example.json`
- 测试工具: `examples/test-header-routing.sh`