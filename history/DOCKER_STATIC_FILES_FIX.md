# Docker 静态文件修复说明

## 问题描述

之前的 Dockerfile 配置中缺少了 Web 静态文件的复制，导致 Docker 容器中无法访问管理界面。

## 修复内容

### 1. Dockerfile 修复

**添加静态文件复制**：
```dockerfile
# 创建必要的目录
RUN mkdir -p /app/data /app/config /app/template /app/web/static /etc/nginx/certs

# 复制静态文件
COPY web/static/ /app/web/static/
```

**更新卷定义**：
```dockerfile
# 定义卷
VOLUME ["/app/data", "/etc/nginx/conf.d", "/etc/nginx/certs", "/var/log/nginx", "/app/template", "/app/config", "/app/web/static"]
```

### 2. 路由配置优化

**移除不存在的 favicon.ico 路由**：
```go
// 修复前
r.StaticFile("/favicon.ico", "./web/static/favicon.ico")

// 修复后 - 已移除，因为文件不存在
```

### 3. 测试脚本

创建了 `scripts/test-docker.sh` 用于验证 Docker 构建：
- 检查静态文件是否正确复制
- 测试 Web 界面访问
- 验证 API 接口功能

## 验证方法

### 1. 构建测试
```bash
# 运行完整测试
./scripts/test-docker.sh
```

### 2. 手动验证
```bash
# 构建镜像
docker build -t nginx-proxy:test .

# 检查静态文件
docker run --rm nginx-proxy:test ls -la /app/web/static/

# 启动容器
docker run -d -p 8080:8080 nginx-proxy:test

# 访问管理界面
curl http://localhost:8080/
```

## 文件结构

修复后的容器内文件结构：
```
/app/
├── web/
│   └── static/
│       ├── index.html      # 主页面
│       └── js/
│           └── app.js      # JavaScript 应用
├── template/               # Nginx 模板
├── config/                 # 配置文件
└── data/                   # 数据库文件
```

## 影响范围

- ✅ Docker 容器现在包含完整的 Web 管理界面
- ✅ 静态文件路由正常工作
- ✅ 移除了不存在文件的路由，避免 404 错误
- ✅ 支持卷挂载自定义静态文件（可选）

## 注意事项

1. **静态文件路径**：应用使用相对路径 `./web/static`，在容器中对应 `/app/web/static`
2. **卷挂载**：如需自定义静态文件，可挂载 `/app/web/static` 卷
3. **权限**：容器内静态文件具有适当的读取权限

## 测试结果

修复后的 Docker 镜像应该能够：
- ✅ 正常访问 Web 管理界面 (http://localhost:8080/)
- ✅ 加载 JavaScript 和 CSS 资源
- ✅ 提供完整的代理配置和证书管理功能