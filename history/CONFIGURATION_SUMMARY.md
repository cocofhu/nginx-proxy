# 配置文件统一说明

## ✅ 配置简化完成

项目现在使用**单一配置文件** `config.json`：

### 📋 **统一配置**

**config.json**

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

### 🚀 **启动命令**

```bash
# 启动 Go 服务
./nginx-proxy -config=config.json

# Docker 部署
docker run -d -p 80:80 -p 8080:8080 \
  -v $(pwd)/config.json:/app/config.json \
  nginx-proxy-openresty
```

### 🔧 **架构说明**

- **Go 服务端口**: 8080
- **OpenResty 调用**: `http://127.0.0.1:8080/api/route`
- **管理接口**: `http://localhost:8080/api/*`
- **配置文件**: 统一使用 `config.json`

### ✅ **简化效果**

- ❌ 删除了冗余的 `config-openresty.json`
- ✅ 统一使用 `config.json`
- ✅ 所有文档引用已更新
- ✅ 端口配置完全一致（8080）

---

**配置文件已简化，避免了重复和混淆！**