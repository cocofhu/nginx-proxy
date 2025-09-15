# 证书管理功能使用指南

## 功能概述

证书管理功能允许您：
- 上传和管理SSL证书
- 自动解析证书信息（域名、过期时间）
- 在创建代理规则时选择已上传的证书
- 监控证书过期状态

## 使用方法

### 1. 上传证书

#### Web界面操作
1. 访问管理界面
2. 点击左侧导航的"证书管理"
3. 点击"上传证书"按钮
4. 填写证书名称
5. 选择证书文件（.crt/.pem）
6. 选择私钥文件（.key）
7. 点击"上传"

#### API调用
```bash
curl -X POST http://localhost:8080/api/certificates \
  -F "cert=@example.crt" \
  -F "key=@example.key"
```

### 2. 查看证书列表

#### Web界面
在证书管理页面可以看到：
- 证书名称
- 关联域名
- 过期时间
- 状态（有效/即将过期/已过期）

#### API调用
```bash
curl http://localhost:8080/api/certificates
```

### 3. 在代理配置中使用证书

#### Web界面操作
1. 在代理配置页面点击"添加代理"
2. 填写域名和其他配置
3. 勾选"启用SSL"
4. 从下拉列表中选择已上传的证书
5. 完成其他配置并提交

#### API调用示例
```bash
# 首先获取证书ID
curl http://localhost:8080/api/certificates

# 使用证书创建代理规则
curl -X POST http://localhost:8080/api/rules \
  -H "Content-Type: application/json" \
  -d '{
    "server_name": "example.com",
    "listen_ports": [443],
    "ssl_cert": "/etc/nginx/certs/cert-id.crt",
    "ssl_key": "/etc/nginx/certs/cert-id.key",
    "locations": [{
      "path": "/",
      "upstreams": [{
        "condition_ip": "0.0.0.0/0",
        "target": "http://backend:8080"
      }]
    }]
  }'
```

## 证书状态说明

- **有效**：证书在有效期内且距离过期超过30天
- **即将过期**：证书距离过期少于30天
- **已过期**：证书已经过期

## 最佳实践

### 1. 证书命名规范
建议使用有意义的名称，如：
- `example-com-2024`
- `wildcard-mydomain-com`
- `api-server-ssl`

### 2. 证书更新流程
1. 上传新证书
2. 更新相关的代理规则
3. 测试配置
4. 删除旧证书

### 3. 监控证书过期
- 定期检查证书管理页面
- 关注"即将过期"状态的证书
- 提前准备证书更新

## 故障排除

### 证书上传失败
- 检查证书文件格式（支持.crt/.pem）
- 确保私钥文件格式正确（.key）
- 验证证书和私钥是否匹配

### 证书无法删除
- 检查是否有代理规则正在使用该证书
- 先更新或删除相关的代理规则
- 然后再删除证书

### SSL配置不生效
- 确认证书文件路径正确
- 检查Nginx配置是否正确生成
- 验证证书是否有效

## 安全注意事项

1. **私钥保护**：私钥文件存储在服务器上，确保适当的文件权限
2. **证书验证**：系统会自动验证证书格式，但建议手动验证证书内容
3. **定期更新**：及时更新即将过期的证书
4. **备份**：定期备份证书文件和配置

## API参考

### 获取所有证书
```
GET /api/certificates
```

### 获取单个证书
```
GET /api/certificates/{id}
```

### 上传证书
```
POST /api/certificates
Content-Type: multipart/form-data

cert: 证书文件
key: 私钥文件
```

### 删除证书
```
DELETE /api/certificates/{id}
```

## 示例场景

### 场景1：为新域名配置HTTPS
1. 获取域名的SSL证书
2. 通过Web界面上传证书
3. 创建代理规则时选择该证书
4. 配置分流规则
5. 测试HTTPS访问

### 场景2：更新即将过期的证书
1. 获取新的SSL证书
2. 上传新证书到系统
3. 编辑使用旧证书的代理规则
4. 选择新证书并保存
5. 删除旧证书

### 场景3：泛域名证书管理
1. 上传泛域名证书（*.example.com）
2. 为多个子域名创建代理规则
3. 所有子域名使用同一个泛域名证书
4. 统一管理证书更新