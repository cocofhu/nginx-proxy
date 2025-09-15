# 构建选项说明

由于 SQLite CGO 在 Alpine Linux 中的编译问题，我们提供了多种构建方案：

## 🚀 推荐方案

### 方案一：纯 Go 构建（默认，推荐）
```bash
# 使用默认 Dockerfile（纯 Go，无 CGO）
docker build -t nginx-proxy .

# 或使用 make
make docker-build
```

**优势**：
- ✅ 无编译问题
- ✅ 镜像更小
- ✅ 启动更快
- ✅ 跨平台兼容

**注意**：使用 GORM 的 SQLite 驱动，功能完整但性能略低于 CGO 版本。

### 方案二：Debian 基础镜像（CGO 版本）
```bash
# 使用 Debian 基础镜像，完全兼容 CGO
docker build -f Dockerfile.debian -t nginx-proxy .

# 或使用 make
make docker-build-debian
```

**优势**：
- ✅ 完全兼容 CGO
- ✅ SQLite 性能最佳
- ✅ 功能完整

**劣势**：
- ❌ 镜像较大
- ❌ 构建时间较长

### 方案三：最简构建
```bash
# 最简单的纯 Go 构建
docker build -f Dockerfile.simple -t nginx-proxy .

# 或使用 make
make docker-build-simple
```

**优势**：
- ✅ 构建最快
- ✅ 镜像最小
- ✅ 静态链接

## 📊 方案对比

| 方案 | 基础镜像 | CGO | 镜像大小 | 构建时间 | SQLite 性能 | 推荐度 |
|------|----------|-----|----------|----------|-------------|--------|
| 纯 Go | Alpine | 否 | 小 | 快 | 良好 | ⭐⭐⭐⭐⭐ |
| Debian | Debian | 是 | 大 | 慢 | 最佳 | ⭐⭐⭐⭐ |
| 最简 | Alpine | 否 | 最小 | 最快 | 良好 | ⭐⭐⭐ |

## 🔧 本地构建选项

### 本地开发（推荐）
```bash
make build          # 使用 CGO（如果环境支持）
./bin/nginx-proxy
```

### 纯 Go 本地构建
```bash
make build-no-cgo   # 纯 Go 构建
./bin/nginx-proxy
```

### 测试不同构建
```bash
# 测试纯 Go 版本
make build-no-cgo && ./bin/nginx-proxy &
curl http://localhost:8080/api/rules

# 测试 CGO 版本（如果环境支持）
make build && ./bin/nginx-proxy &
curl http://localhost:8080/api/rules
```

## 🐳 Docker 构建命令

```bash
# 默认构建（推荐）
make docker-build

# Debian 版本（如果需要最佳性能）
make docker-build-debian

# 最简版本（如果需要最小镜像）
make docker-build-simple
```

## 🎯 选择建议

1. **开发环境**：使用 `make build`（本地 CGO）
2. **生产环境**：使用默认 Dockerfile（纯 Go）
3. **高性能需求**：使用 Dockerfile.debian（CGO）
4. **资源受限**：使用 Dockerfile.simple（最小）

## 🔍 故障排除

### 如果默认构建失败
```bash
# 尝试 Debian 版本
make docker-build-debian
```

### 如果需要最小镜像
```bash
# 使用最简构建
make docker-build-simple
```

### 如果需要调试
```bash
# 本地构建测试
make build-no-cgo
./bin/nginx-proxy --help