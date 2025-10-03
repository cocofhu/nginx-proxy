# 智能反向代理管理器

基于 **OpenResty + Go 接口** 架构的智能反向代理管理系统，支持复杂路由规则、头部条件匹配和动态配置管理。

## 🎯 核心特性

### 🚀 **智能路由系统**

- **头部条件"且"关系**：支持多个 HTTP 头部条件同时匹配
- **IP 段匹配**：支持 CIDR 格式的 IP 段路由（如 `192.168.1.0/24`）
- **动态路由判断**：通过 Go 接口实现复杂路由逻辑
- **实时路由切换**：无需重启即可更新路由规则

### 🔧 **管理功能**

- **REST API 管理**：完整的 CRUD 操作管理反向代理规则
- **Web 管理界面**：现代化的响应式管理界面
- **证书管理**：SSL 证书上传、管理和自动配置
- **腾讯云证书集成**：自动申请、续期和管理腾讯云免费SSL证书
- **智能证书清理**：删除证书时同步删除腾讯云端证书
- **配置验证**：自动验证 OpenResty 配置正确性
- **热重载**：配置变更后自动重载 OpenResty

### 💾 **数据存储**

- **SQLite 数据库**：使用纯 Go 驱动，无 CGO 依赖
- **持久化配置**：所有路由规则持久化存储
- **证书状态跟踪**：实时跟踪腾讯云证书状态和过期时间
- **配置备份**：自动生成配置文件备份

### 🔒 **SSL证书管理**

- **多证书来源**：支持本地上传和腾讯云证书
- **自动申请**：一键申请腾讯云免费SSL证书
- **智能续期**：证书到期前自动续期，无需人工干预
- **同步删除**：删除证书时自动清理腾讯云端对应证书
- **状态监控**：实时监控证书状态（申请中/正常/即将过期/已过期）
- **灵活配置**：支持仅HTTPS或HTTPS+HTTP重定向模式

## 🏗️ 技术架构

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Client        │───▶│   OpenResty      │───▶│   Go Service    │
│   Request       │    │   (Nginx + Lua)  │    │   (Port 8080)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │                         │
                              │   HTTP POST             │
                              │   /api/route            │
                              │                         │
                              ▼                         ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │  Route Decision  │    │  Route Logic    │
                       │  (Lua Script)    │    │  (Go Handler)   │
                       └──────────────────┘    └─────────────────┘
```

### 技术栈

- **OpenResty**：高性能 Web 平台（Nginx + LuaJIT）
- **Go 1.24+**：路由逻辑处理（纯 Go，无 CGO 依赖）
- **Gin**：HTTP 框架
- **SQLite + GORM**：数据库和 ORM
- **Lua**：动态路由脚本
- **Docker**：容器化部署

## 🚀 快速开始

### Docker 部署（推荐）

使用 Docker 一键部署 OpenResty + Go 服务：

```bash

# quick start
docker run --restart always -d --name nginx-proxy \
  -p 80:80 \
  -p 443:443 \
  -e TENCENT_SECRET_ID=ak \
  -e TENCENT_SECRET_KEY=sk \
  -e TENCENT_REGION=ap-beijing \
  -e CLOUDFLARE_TOKEN=token \
  -e CLOUDFLARE_DOMAINS=cocofhu.cc \
  -e SERVICE_NAME=bridge \
  -e SERVICE_PORT=8080 \
  -e SERVICE_CHECK_TCP=true \
  -e SERVICE_CHECK_DEREGISTER_AFTER=30s \
  --volume /volume1/docker/nginx-proxy/data:/app/data \
  --volume /volume1/docker/nginx-proxy/logs:/app/logs \
  --volume /volume1/docker/nginx-proxy/certs:/etc/nginx/certs \
  --volume /volume1/docker/nginx-proxy/config:/etc/nginx/conf.d \
  --volume /volume1/docker/nginx-proxy/nginx-logs:/var/log/nginx \
  --volume /volume1/docker/nginx-proxy/nginx-logs:/var/log/nginx \
  ccr.ccs.tencentyun.com/cocofhu/nginx-proxy

# macvlan
docker run --restart always -d --name nginx-proxy --net=macvlan_net \
  -e TENCENT_SECRET_ID=ak \
  -e TENCENT_SECRET_KEY=sk \
  -e TENCENT_REGION=ap-beijing \
  -e CLOUDFLARE_TOKEN=token \
  -e CLOUDFLARE_DOMAINS=cocofhu.cc \
  -e SERVICE_NAME=bridge \
  -e SERVICE_PORT=8080 \
  -e SERVICE_CHECK_TCP=true \
  -e SERVICE_CHECK_DEREGISTER_AFTER=30s \
  --volume /volume1/docker/nginx-proxy/data:/app/data \
  --volume /volume1/docker/nginx-proxy/logs:/app/logs \
  --volume /volume1/docker/nginx-proxy/certs:/etc/nginx/certs \
  --volume /volume1/docker/nginx-proxy/config:/etc/nginx/conf.d \
  --volume /volume1/docker/nginx-proxy/nginx-logs:/var/log/nginx \
  --volume /volume1/docker/nginx-proxy/nginx-logs:/var/log/nginx \
  ccr.ccs.tencentyun.com/cocofhu/nginx-proxy
```

### Web 管理界面

访问 `http://localhost:8080` 使用 Web 界面管理：

- 📊 **仪表板**：系统概览和路由统计
- ⚙️ **路由配置**：管理复杂路由规则和头部条件
- 🔒 **证书管理**：SSL 证书上传和管理
- 📋 **日志查看**：实时查看路由匹配日志

## 📋 更新日志

### v1.1.0 (最新版本)

- ✨ **新增腾讯云SSL证书集成**
    - 支持一键申请腾讯云免费SSL证书
    - 自动DNS验证和证书下载
    - 证书状态实时监控和管理
- ✨ **智能证书生命周期管理**
    - 证书到期前自动续期
    - 删除证书时同步清理腾讯云端证书
    - 批量证书续期功能
- ✨ **增强的SSL配置选项**
    - 支持仅HTTPS模式（高安全）
    - 支持HTTPS+HTTP重定向模式
    - 灵活的SSL端口配置
- 🐛 **前端功能修复**
    - 修复添加代理功能无响应问题
    - 修复证书选择下拉框数据加载
    - 修复JavaScript运行时错误
- 🔧 **用户体验优化**
    - 改进证书管理界面
    - 添加证书调试工具页面
    - 优化异步数据加载和错误处理

### v1.0.0

- 🎉 初始版本发布
- ✨ 基础反向代理功能
- ✨ Web 管理界面
- ✨ SQLite 数据存储
- ✨ 智能路由系统重构
- ✨ OpenResty + Lua 架构升级
- ✨ 复杂头部条件匹配支持
- 🔧 Docker 容器化部署

## 🌟 致谢

感谢所有贡献者对项目的支持！特别感谢：

- 腾讯云团队提供的免费SSL证书服务
- OpenResty 社区的技术支持
- 所有提交Bug报告和功能建议的用户

---

**⭐ 如果这个项目对您有帮助，请给我们一个 Star！**

**🚀 立即体验腾讯云证书自动管理功能，让SSL证书管理变得简单高效！**