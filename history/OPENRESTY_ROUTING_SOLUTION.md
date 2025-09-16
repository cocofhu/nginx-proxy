# OpenResty + Go 接口路由解决方案

## 方案概述

为了简化复杂的头部条件判断逻辑，我们采用 OpenResty 调用 Go 接口的方式来实现路由判断。这种方案具有以下优势：

- **逻辑简化**：避免复杂的 nginx 变量操作
- **灵活性强**：可以轻松扩展复杂的路由逻辑
- **易于维护**：路由逻辑集中在 Go 代码中
- **性能优化**：减少 nginx 配置文件的复杂度

## 架构设计

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Client        │───▶│   OpenResty      │───▶│   Go Service    │
│   Request       │    │   (Nginx + Lua)  │    │   (Port 8080)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │                         │
                              │   HTTP POST             │
                              │   /api/route-match      │
                              │                         │
                              ▼                         ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │  Route Decision  │    │  Route Logic    │
                       │  (Lua Script)    │    │  (Go Handler)   │
                       └──────────────────┘    └─────────────────┘
```

## 实现细节

### 1. Nginx 配置模板更新

在 `template/nginx.conf.tpl` 中，头部条件判断现在使用 Lua 脚本：

```lua
access_by_lua_block {
    local http = require "resty.http"
    local cjson = require "cjson"
    
    -- 收集请求头
    local headers = {}
    headers["tt"] = ngx.var.http_tt
    headers["x-env"] = ngx.var.http_x_env
    headers["x-token"] = ngx.var.http_x_token
    
    -- 构建请求数据
    local request_data = {
        headers = headers,
        expected = {
            ["tt"] = "t",
            ["x-env"] = "test",
            ["x-token"] = "123"
        },
        target = "http://21.91.124.161:8080",
        rule_index = 0
    }
    
    -- 调用路由判断接口
    local httpc = http.new()
    local res, err = httpc:request_uri("http://127.0.0.1:8080/api/route-match", {
        method = "POST",
        body = cjson.encode(request_data),
        headers = {
            ["Content-Type"] = "application/json"
        }
    })
    
    if res and res.status == 200 then
        local result = cjson.decode(res.body)
        if result.match then
            ngx.var.backend = result.target
        end
    end
}
```

### 2. Go 路由匹配接口

在 `internal/api/handlers.go` 中新增了 `RouteMatch` 处理函数：

```go
// RouteMatchRequest 路由匹配请求结构
type RouteMatchRequest struct {
    Headers   map[string]string `json:"headers"`
    Expected  map[string]string `json:"expected"`
    Target    string            `json:"target"`
    RuleIndex int               `json:"rule_index"`
}

// RouteMatchResponse 路由匹配响应结构
type RouteMatchResponse struct {
    Match  bool   `json:"match"`
    Target string `json:"target"`
}

// RouteMatch 路由匹配接口（供 OpenResty 调用）
func (h *Handler) RouteMatch(c *gin.Context) {
    var req RouteMatchRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 检查所有期望的头部条件是否都匹配
    match := true
    for expectedKey, expectedValue := range req.Expected {
        actualValue, exists := req.Headers[expectedKey]
        if !exists || actualValue != expectedValue {
            match = false
            break
        }
    }

    response := RouteMatchResponse{
        Match:  match,
        Target: req.Target,
    }

    c.JSON(http.StatusOK, response)
}
```

### 3. 服务配置

创建了 `config.json` 配置文件，让 Go 服务监听在 8080 端口：

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

## 部署步骤

### 1. 启动 Go 服务

```bash
# 使用 OpenResty 配置启动服务
./nginx-proxy -config=config.json
```

### 2. 确保 OpenResty 环境

需要安装 OpenResty 并确保包含以下 Lua 模块：
- `resty.http` - HTTP 客户端
- `cjson` - JSON 编解码

### 3. 测试路由匹配

```bash
# 测试路由匹配接口
curl -X POST http://localhost:8080/api/route-match \
  -H "Content-Type: application/json" \
  -d '{
    "headers": {
      "tt": "t",
      "x-env": "test",
      "x-token": "123"
    },
    "expected": {
      "tt": "t",
      "x-env": "test",
      "x-token": "123"
    },
    "target": "http://21.91.124.161:8080",
    "rule_index": 0
  }'
```

期望响应：
```json
{
  "match": true,
  "target": "http://21.91.124.161:8080"
}
```

## 优势对比

### 原方案（复杂变量操作）
```nginx
set $match0_h0 0;
set $match0_h1 0;
set $match0_h2 0;
if ($http_tt = "t") { set $match0_h0 1; }
if ($http_x_env = "test") { set $match0_h1 1; }
if ($http_x_token = "123") { set $match0_h2 1; }
if ($match0_h0$match0_h1$match0_h2 = "111") {
    set $backend "http://21.91.124.161:8080";
}
```

### 新方案（OpenResty + Go）
- ✅ 逻辑清晰，易于理解
- ✅ 支持复杂的条件判断（如正则、范围等）
- ✅ 便于调试和日志记录
- ✅ 可以轻松扩展新的路由规则
- ✅ 集中化的路由逻辑管理

## 性能考虑

- **延迟**：每次路由判断需要一次内部 HTTP 调用（通常 < 1ms）
- **并发**：Go 服务可以处理高并发的路由判断请求
- **缓存**：可以在 Lua 层面添加缓存机制优化性能
- **连接池**：`resty.http` 支持连接池，减少连接开销

## 扩展能力

这种架构支持未来的扩展需求：
- 基于 IP 地理位置的路由
- 基于请求频率的限流路由
- 基于用户认证状态的路由
- 复杂的 A/B 测试路由规则