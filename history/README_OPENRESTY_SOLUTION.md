# 完全基于接口的路由解决方案

## 问题解决

原始问题：nginx 头部条件使用多个独立 `if` 语句导致"或"关系，而不是期望的"且"关系。

## 解决方案

采用 **完全基于 Go 接口** 的架构，将所有路由判断逻辑从 nginx 配置中移到 Go 服务中。

## 核心组件

### 1. 极简 Nginx 配置模板 (`template/nginx.conf.tpl`)

所有路由逻辑统一通过 Go 接口处理：

```lua
location /api {
    # 统一使用 Go 接口进行路由判断
    access_by_lua_block {
        local http = require "resty.http"
        local cjson = require "cjson"
        
        -- 收集所有请求信息
        local request_data = {
            path = "/api",
            remote_addr = ngx.var.remote_addr,
            headers = ngx.req.get_headers(),
            upstreams = {
                {
                    target = "http://21.91.124.161:8080",
                    condition_ip = "",
                    headers = {
                        ["tt"] = "t",
                        ["x-env"] = "test",
                        ["x-token"] = "123"
                    }
                },
                {
                    target = "http://default-backend:8080",
                    condition_ip = "",
                    headers = {}
                }
            }
        }
        
        -- 调用统一路由接口
        local httpc = http.new()
        local res, err = httpc:request_uri("http://127.0.0.1:8080/api/route", {
            method = "POST",
            body = cjson.encode(request_data),
            headers = { ["Content-Type"] = "application/json" }
        })
        
        if res and res.status == 200 then
            local result = cjson.decode(res.body)
            if result.target then
                ngx.var.backend = result.target
            else
                ngx.var.backend = "http://default-backend:8080"
            end
        else
            ngx.var.backend = "http://default-backend:8080"
        end
    }

    set $upstream $backend;
    proxy_pass $upstream;
}
```

### 2. 统一 Go 路由接口 (`internal/api/handlers.go`)

```go
// Route 统一路由接口（供 OpenResty 调用）
func (h *Handler) Route(c *gin.Context) {
    var req RouteRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 遍历所有上游服务器，找到匹配的
    for _, upstream := range req.Upstreams {
        if h.matchUpstream(req, upstream) {
            c.JSON(http.StatusOK, RouteResponse{Target: upstream.Target})
            return
        }
    }

    // 如果没有匹配，返回空（使用默认）
    c.JSON(http.StatusOK, RouteResponse{Target: ""})
}

// matchUpstream 检查上游服务器是否匹配
func (h *Handler) matchUpstream(req RouteRequest, upstream Upstream) bool {
    // 检查 IP 条件
    if !h.matchIP(req.RemoteAddr, upstream.ConditionIP) {
        return false
    }

    // 检查头部条件（且关系）
    if !h.matchHeaders(req.Headers, upstream.Headers) {
        return false
    }

    return true
}

// matchHeaders 检查头部是否匹配（且关系）
func (h *Handler) matchHeaders(requestHeaders, expectedHeaders map[string]string) bool {
    // 如果没有期望的头部条件，直接匹配
    if len(expectedHeaders) == 0 {
        return true
    }

    // 检查所有期望的头部条件是否都匹配
    for expectedKey, expectedValue := range expectedHeaders {
        actualValue, exists := requestHeaders[expectedKey]
        if !exists || actualValue != expectedValue {
            return false
        }
    }

    return true
}
```

## 部署步骤

1. **启动 Go 服务**：
   ```bash
   ./nginx-proxy -config=config.json
   ```

2. **确保 OpenResty 环境**：
   - 安装 OpenResty
   - 确保包含 `resty.http` 和 `cjson` 模块

3. **测试配置**：
   使用 `examples/openresty-routing-example.json`

## 优势

- ✅ **逻辑简化**：避免复杂的 nginx 变量操作
- ✅ **易于扩展**：可以轻松添加更复杂的路由规则  
- ✅ **便于调试**：路由逻辑集中在 Go 代码中
- ✅ **性能优化**：减少 nginx 配置复杂度

## 测试示例

```bash
# 测试统一路由接口
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
        "condition_ip": "",
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

期望响应（匹配成功）：
```json
{
  "target": "http://21.91.124.161:8080"
}
```

期望响应（匹配失败，使用默认）：
```json
{
  "target": ""
}
```

## 优势总结

- ✅ **极简配置**：nginx 配置文件非常简洁
- ✅ **统一逻辑**：所有路由逻辑集中在 Go 服务中
- ✅ **易于扩展**：可以轻松添加任何复杂的路由规则
- ✅ **便于调试**：路由逻辑完全在 Go 代码中，便于调试和日志记录
- ✅ **完全消除**：彻底避免了 nginx 配置中的复杂变量操作

现在您的三个头部条件（`tt=t`, `x-env=test`, `x-token=123`）将通过 Go 接口进行正确的"且"关系判断，配置极其简洁！