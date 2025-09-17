# SSL配置HTTP端口选项功能

## 功能说明

在启用SSL时，现在用户可以选择是否同时启用HTTP 80端口。

## 配置选项

### 1. **仅HTTPS（推荐用于安全要求高的场景）**

- ✅ 启用SSL
- ❌ 取消勾选"同时启用HTTP 80端口"
- **结果**：只监听443端口，不接受HTTP请求

### 2. **HTTPS + HTTP重定向（默认推荐）**

- ✅ 启用SSL
- ✅ 勾选"同时启用HTTP 80端口（自动重定向到HTTPS）"
- **结果**：监听80和443端口，HTTP自动重定向到HTTPS

### 3. **仅HTTP（不推荐）**

- ❌ 不启用SSL
- **结果**：只监听80端口

## 使用场景

### 仅HTTPS模式适用于：

- 内部API服务
- 高安全要求的应用
- 不希望暴露HTTP端口的服务

### HTTPS + HTTP重定向模式适用于：

- 面向用户的网站
- 需要SEO友好的应用
- 兼容旧链接的服务

## 技术实现

### 端口配置逻辑：

```javascript
let listenPorts = [80]; // 默认HTTP

if (sslEnabled) {
    listenPorts.push(443); // 添加HTTPS
    if (!httpRedirect) {
        listenPorts = [443]; // 仅HTTPS
    }
}
```

### 生成的配置示例：

**仅HTTPS模式：**

```json
{
  "listen_ports": [443],
  "ssl_cert": "/path/to/cert.pem",
  "ssl_key": "/path/to/key.pem"
}
```

**HTTPS + HTTP重定向模式：**

```json
{
  "listen_ports": [80, 443],
  "ssl_cert": "/path/to/cert.pem", 
  "ssl_key": "/path/to/key.pem"
}
```

现在用户在配置SSL时有了更灵活的选择！