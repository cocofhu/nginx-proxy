# 证书验证记录清理服务

## 概述

证书验证记录清理服务是一个后台任务，用于自动清理SSL证书验证过程中产生的DNS记录。当证书验证完成后，这些临时的DNS记录应该被清理，以避免DNS记录越来越多。

## 功能特性

- **自动检测证书状态**：通过腾讯云SSL API检查证书验证是否完成
- **多DNS提供商支持**：支持Cloudflare和腾讯云DNS记录清理
- **定时清理**：默认每小时执行一次清理任务
- **安全清理**：只清理已完成验证的证书相关DNS记录

## 配置方法

### 1. 环境变量配置

创建 `.env` 文件或设置环境变量：

```bash
# Cloudflare配置（二选一）
CLOUDFLARE_API_TOKEN=your_api_token
# 或者
CLOUDFLARE_EMAIL=your_email@example.com
CLOUDFLARE_API_KEY=your_api_key

# 腾讯云配置
TENCENT_SECRET_ID=your_secret_id
TENCENT_SECRET_KEY=your_secret_key
TENCENT_REGION=ap-beijing
```

### 2. 配置文件

如果使用config.json配置文件，腾讯云配置会从配置文件中读取：

```json
{
  "tencent_cloud": {
    "secret_id": "your_secret_id",
    "secret_key": "your_secret_key",
    "region": "ap-beijing"
  }
}
```

## 工作流程

1. **启动服务**：在main server启动后自动启动清理服务
2. **定时扫描**：每小时扫描数据库中的AuthRecord记录
3. **状态检查**：对于每条记录，检查对应证书的验证状态
4. **清理操作**：
   - 如果证书验证已完成，删除对应的DNS记录
   - 从数据库中删除AuthRecord记录
5. **日志记录**：记录清理过程和结果

## 数据库表结构

清理服务操作的是 `AuthRecord` 表：

```go
type AuthRecord struct {
    ID            string    // 记录ID
    Domain        string    // 主域名
    Key           string    // DNS记录名
    Value         string    // DNS记录值
    Type          string    // 记录类型（通常是TXT）
    Action        string    // 操作类型（add/delete）
    Source        string    // DNS提供商（cloudflare/tencent_cloud）
    CertificateId string    // 关联的证书ID
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

## 证书状态说明

腾讯云证书状态码：
- 1: 审核中
- 2: 审核失败
- 3: 已通过
- 4: 已颁发
- 5: 已过期
- 6: 已撤销
- 7: 已删除

状态 >= 3 时认为验证已完成，可以清理DNS记录。

## 日志监控

清理服务会输出详细的日志信息：

```
启动证书验证记录清理服务，清理间隔: 1h0m0s
开始清理证书验证记录...
找到 5 条待清理记录
成功清理记录 [ID: xxx, Domain: example.com]
证书验证记录清理完成
```

## 安全注意事项

1. **API权限**：确保Cloudflare和腾讯云API具有适当的DNS管理权限
2. **证书状态**：只有在确认证书验证完成后才会清理DNS记录
3. **错误处理**：清理失败的记录会保留，等待下次清理
4. **日志审计**：所有清理操作都会记录日志，便于审计

## 配置验证

在启动服务前，建议先验证配置是否正确：

```bash
# 运行配置验证脚本
go run scripts/validate_config.go
```

该脚本会检查：
- 环境变量是否正确设置
- API客户端是否能正常初始化
- 提供配置建议

## 测试功能

### 基础测试

```bash
# 运行基础测试
go run scripts/test_cleanup.go
```

### 高级测试

```bash
# 运行高级测试（需要真实的API配置）
go run scripts/test_cleanup_advanced.go
```

高级测试会：
- 创建多种类型的测试记录
- 测试不同DNS提供商的清理逻辑
- 验证证书状态检查功能
- 显示详细的清理结果

## 故障排除

### 常见问题

1. **清理服务未启动**
   ```
   Warning: Failed to initialize cleanup service: xxx
   ```
   - 检查环境变量配置是否正确
   - 运行 `go run scripts/validate_config.go` 验证配置
   - 查看启动日志中的详细错误信息

2. **DNS记录清理失败**
   ```
   清理Cloudflare记录失败: xxx
   清理腾讯云记录失败: xxx
   ```
   - 检查API凭据是否有效且未过期
   - 确认API权限是否包含DNS管理
   - 检查域名是否在对应平台托管
   - 确认DNS记录是否真实存在

3. **证书状态检查失败**
   ```
   failed to query certificate from tencent cloud: xxx
   ```
   - 检查腾讯云API配置是否正确
   - 确认证书ID是否有效
   - 检查API权限是否包含SSL证书查询

4. **腾讯云DNS记录查询失败**
   ```
   查询DNS记录列表失败: xxx
   ```
   - 确认域名在腾讯云DNSPod托管
   - 检查DNSPod API权限
   - 验证域名格式是否正确

### 调试技巧

1. **启用详细日志**
   ```go
   // 在main.go中添加
   log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
   ```

2. **手动测试API连接**
   ```bash
   # 测试Cloudflare API
   curl -X GET "https://api.cloudflare.com/client/v4/user/tokens/verify" \
        -H "Authorization: Bearer YOUR_TOKEN"

   # 测试腾讯云API（需要签名，建议使用SDK测试）
   go run scripts/validate_config.go
   ```

3. **检查数据库记录**
   ```sql
   -- 查看所有待清理记录
   SELECT * FROM auth_records WHERE deleted_at IS NULL;

   -- 查看特定来源的记录
   SELECT * FROM auth_records WHERE source = 'cloudflare';
   SELECT * FROM auth_records WHERE source = 'tencent_cloud';

   -- 查看有证书ID的记录
   SELECT * FROM auth_records WHERE certificate_id != '';
   ```

### 手动清理

如果自动清理失败，可以手动清理：

```sql
-- 查看待清理记录
SELECT id, domain, key, value, source, certificate_id, created_at 
FROM auth_records 
WHERE action = 'add' 
ORDER BY created_at DESC;

-- 删除特定记录
DELETE FROM auth_records WHERE id = 'record_id';

-- 批量删除过期记录（超过7天的记录）
DELETE FROM auth_records 
WHERE created_at < datetime('now', '-7 days');
```

### 监控建议

1. **日志监控**
   - 监控清理成功/失败的数量
   - 关注API调用错误
   - 跟踪清理耗时

2. **数据库监控**
   - 监控AuthRecord表的记录数量
   - 设置记录数量告警（如超过1000条）
   - 定期检查是否有长期未清理的记录

3. **API配额监控**
   - 监控Cloudflare API调用次数
   - 监控腾讯云API调用次数
   - 避免超出API限制

## 性能优化

- 清理间隔可以根据实际需求调整
- 大量记录时可以考虑批量处理
- 可以添加并发控制来提高清理效率