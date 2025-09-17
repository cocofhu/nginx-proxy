# 腾讯云SSL证书删除和续期功能测试

## 功能说明

本次更新为腾讯云SSL证书管理添加了以下功能：

### 1. 删除证书时同时删除腾讯云端证书

当删除本地证书记录时，系统会自动调用腾讯云API删除云端的证书，确保云端和本地数据的一致性。

**API接口：** `DELETE /api/certificates/tencent/{id}`

**功能流程：**

1. 查找本地证书记录
2. 调用腾讯云API删除云端证书
3. 删除本地证书文件
4. 删除数据库记录

### 2. 续期时自动删除老证书

当证书续期完成后，系统会自动删除腾讯云端的老证书，避免证书堆积。

**功能流程：**

1. 申请新证书
2. 等待新证书签发完成
3. 下载新证书文件
4. 更新本地记录使用新证书
5. **删除腾讯云端的老证书**
6. 删除本地老证书文件
7. 更新相关nginx配置

## 新增函数

### `deleteTencentCloudCertificate(certificateID string) error`

专门用于删除腾讯云端证书的内部函数。

**特性：**

- 调用腾讯云DeleteCertificate API
- 处理证书不存在的情况（不视为错误）
- 详细的错误日志记录

### 更新的函数

#### `DeleteTencentCertificate(certificateID string) error`

增强了删除功能，现在会同时删除腾讯云端的证书。

#### `handleRenewalCompletion(originalCertID, newCertID string) error`

在续期完成处理中增加了删除老证书的逻辑。

## 测试步骤

### 1. 测试删除功能

```bash
# 1. 申请一个测试证书
curl -X POST http://localhost:8080/api/certificates/tencent/apply \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "test.example.com",
    "validate_type": "DNS_AUTO",
    "cert_alias": "test-cert-for-delete"
  }'

# 2. 等待证书签发完成（可通过状态查询接口检查）
curl http://localhost:8080/api/certificates/tencent/{certificate_id}/status

# 3. 删除证书（应该同时删除腾讯云端的证书）
curl -X DELETE http://localhost:8080/api/certificates/tencent/{certificate_id}
```

### 2. 测试续期功能

```bash
# 1. 申请一个测试证书
curl -X POST http://localhost:8080/api/certificates/tencent/apply \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "renew-test.example.com",
    "validate_type": "DNS_AUTO",
    "cert_alias": "test-cert-for-renew"
  }'

# 2. 等待证书签发完成
curl http://localhost:8080/api/certificates/tencent/{certificate_id}/status

# 3. 执行续期（会申请新证书并在完成后删除老证书）
curl -X POST http://localhost:8080/api/certificates/tencent/{certificate_id}/renew
```

## 注意事项

1. **权限要求：** 确保腾讯云API密钥有删除证书的权限
2. **错误处理：** 如果腾讯云端删除失败，本地删除仍会继续执行
3. **日志记录：** 所有删除操作都有详细的日志记录
4. **数据一致性：** 通过数据库迁移确保所有必要字段都存在

## 配置要求

确保 `config.json` 中的腾讯云配置正确：

```json
{
  "tencent_cloud": {
    "secret_id": "your_secret_id",
    "secret_key": "your_secret_key",
    "region": "ap-beijing"
  }
}
```

## 错误处理

系统会优雅处理以下情况：

- 证书在腾讯云端不存在
- 证书已被删除
- 网络连接问题
- API权限不足

在这些情况下，系统会记录警告日志但不会阻止本地操作的完成。