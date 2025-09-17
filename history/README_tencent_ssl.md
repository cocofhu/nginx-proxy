# 腾讯云SSL证书管理API

本模块提供腾讯云SSL证书的完整管理功能，包括证书申请、状态查询、下载和删除等操作。**已集成真实的腾讯云SDK**，支持生产环境使用。

## 功能特性

- ✅ 申请腾讯云免费SSL证书（真实API调用）
- ✅ 查询证书申请状态（实时同步腾讯云状态）
- ✅ 下载已签发的证书（真实证书文件）
- ✅ 获取证书列表（同步腾讯云证书到本地）
- ✅ 删除证书记录（本地记录和文件管理）
- ✅ 支持DNS自动验证、DNS手动验证和文件验证
- ✅ 自动解析证书过期时间
- ✅ **智能续期功能**：自动检测即将过期的证书并重新申请
- ✅ **批量续期**：一键批量续期所有即将过期的证书
- ✅ **续期状态检查**：检查证书是否需要续期，支持自定义过期天数
- ✅ 完整的错误处理和腾讯云API错误映射
- ✅ 中文错误提示和日志记录

## API接口

### 1. 申请证书

```http
POST /api/certificates/tencent/apply
Content-Type: application/json

{
  "domain": "example.com",
  "validate_type": "DNS_AUTO",
  "cert_alias": "我的证书"
}
```

**响应示例：**

```json
{
  "certificate_id": "cert_1234567890",
  "status": "申请中",
  "validate_info": {
    "type": "DNS",
    "record": "_dnsauth.example.com",
    "value": "verification_value_cert_1234567890"
  }
}
```

### 2. 查询证书状态

```http
GET /api/certificates/tencent/{certificate_id}/status
```

**响应示例：**

```json
{
  "certificate_id": "cert_1234567890",
  "status": "已通过",
  "domain": "example.com",
  "expires_at": "2025-01-15 23:59:59"
}
```

### 3. 下载证书

```http
POST /api/certificates/tencent/{certificate_id}/download
```

**响应示例：**

```json
{
  "message": "证书下载成功"
}
```

### 4. 获取证书列表

```http
GET /api/certificates/tencent/list
```

**响应示例：**

```json
{
  "certificates": [
    {
      "id": "cert_1234567890",
      "name": "我的证书",
      "domain": "example.com",
      "cert_path": "./certs/cert_1234567890.crt",
      "key_path": "./certs/cert_1234567890.key",
      "expires_at": "2025-01-15T00:00:00Z",
      "source": "tencent_cloud",
      "source_id": "cert_1234567890",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1
}
```

### 5. 删除证书

```http
DELETE /api/certificates/tencent/{certificate_id}
```

**响应示例：**

```json
{
  "message": "证书删除成功"
}
```

### 6. 续期证书（随时可操作）

```http
POST /api/certificates/tencent/{certificate_id}/renew
```

**功能说明：**

- 可以在任何时候操作续期，无需等待证书即将过期
- 通过重新申请新证书来替换现有证书
- 旧证书会被标记为已续期，新证书将使用相同的域名

**响应示例：**

```json
{
  "message": "证书续期申请成功",
  "old_cert_id": "cert_1234567890",
  "new_cert_id": "cert_9876543210",
  "status": "申请中",
  "validate_info": {
    "type": "DNS",
    "record": "_dnsauth.example.com",
    "value": "verification_value_cert_9876543210"
  }
}
```

### 7. 检查续期状态

```http
GET /api/certificates/tencent/{certificate_id}/renewal-status?days=30
```

**响应示例：**

```json
{
  "certificate_id": "cert_1234567890",
  "need_renewal": true,
  "days_before_expiry": 30
}
```

### 8. 批量续期

```http
POST /api/certificates/tencent/batch-renew
```

**功能说明：**

- 批量续期所有腾讯云证书，无时间限制
- 可选参数 `days` 用于筛选即将过期的证书（如不指定则续期所有证书）

**请求示例：**

```http
# 续期所有证书
POST /api/certificates/tencent/batch-renew

# 只续期30天内过期的证书
POST /api/certificates/tencent/batch-renew?days=30
```

**响应示例：**

```json
{
  "message": "批量续期完成",
  "renewed_count": 3,
  "total_count": 5
}
```

## 配置要求

在 `config.json` 中配置腾讯云访问凭证：

```json
{
  "tencent_cloud": {
    "secret_id": "your_secret_id_here",
    "secret_key": "your_secret_key_here",
    "region": "ap-beijing"
  },
  "ssl": {
    "cert_dir": "./certs"
  }
}
```

## 依赖要求

需要安装腾讯云SDK依赖：

```bash
go mod tidy
```

主要依赖包：

- `github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common`
- `github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl`

## 腾讯云权限配置

确保您的腾讯云账号具有以下权限：

1. **SSL证书管理权限**:
    - `ssl:ApplyCertificate` - 申请证书
    - `ssl:DescribeCertificates` - 查询证书列表
    - `ssl:DescribeCertificateDetail` - 查询证书详情
    - `ssl:DownloadCertificate` - 下载证书

2. **建议使用子账号**:
    - 创建专门的子账号用于API调用
    - 只授予必要的SSL证书管理权限
    - 定期轮换访问密钥

## 验证方式

支持以下验证方式：

1. **DNS_AUTO**: 自动DNS验证（推荐，腾讯云自动完成验证）
2. **DNS**: 手动DNS验证（需要手动添加DNS记录）
3. **FILE_VALIDATION**: 文件验证（需要在网站根目录放置验证文件）

## 证书状态映射

腾讯云证书状态与中文描述的映射：

| 状态码 | 中文描述         | 说明           |
|-----|--------------|--------------|
| 0   | 审核中          | 证书申请正在审核     |
| 1   | 已通过          | 证书已签发，可以下载   |
| 2   | 审核失败         | 证书申请被拒绝      |
| 3   | 已过期          | 证书已过期        |
| 4   | DNS记录添加中     | 正在添加DNS验证记录  |
| 5   | 企业证书，待提交     | 企业证书需要提交额外资料 |
| 6   | 订单取消中        | 正在取消证书订单     |
| 7   | 已取消          | 证书订单已取消      |
| 8   | 已提交资料，待上传确认函 | 需要上传确认函      |
| 9   | 证书吊销中        | 正在吊销证书       |
| 10  | 已吊销          | 证书已被吊销       |
| 11  | 重颁发中         | 正在重新颁发证书     |
| 12  | 待上传吊销确认函     | 需要上传吊销确认函    |

## 错误处理

系统提供详细的错误处理：

- **腾讯云API错误**: 自动解析并返回腾讯云的错误码和消息
- **网络错误**: 处理网络连接问题
- **权限错误**: 处理API权限不足的情况
- **参数错误**: 验证请求参数的有效性
- **文件操作错误**: 处理证书文件读写问题

常见错误示例：

```json
{
  "error": "腾讯云API错误: InvalidParameter.Domain - 域名格式不正确"
}
```

## 注意事项

1. **真实腾讯云集成**: 已集成腾讯云SSL证书管理API，支持真实的证书申请、查询、下载功能
2. **API凭证配置**: 需要在配置文件中正确设置腾讯云的 `secret_id` 和 `secret_key`
3. **证书文件管理**: 证书文件存储在配置的 `cert_dir` 目录中，自动管理文件生命周期
4. **状态同步**: 自动同步腾讯云证书状态到本地数据库
5. **验证方式**: 支持DNS自动验证、DNS手动验证和文件验证
6. **智能续期**: 续期功能通过重新申请新证书来替换现有证书，可随时操作，无需等待证书即将过期
7. **灵活续期策略**:
    - 单证书续期：任何时候都可以操作，无时间限制
    - 批量续期：可选择续期所有证书或仅续期即将过期的证书
    - 支持自定义过期时间阈值（默认30天）
8. **无缝替换**: 续期过程中旧证书会被标记，新证书使用相同域名，确保服务连续性
9. **权限要求**: 确保腾讯云账号具有SSL证书管理的相关权限
10. **免费证书限制**: 腾讯云免费证书通常不支持API删除，只能在控制台操作
11. **域名验证**: 申请证书前确保域名已正确解析并可访问

## 文件结构

```
internal/api/tencent_ssl.go     # API处理器
internal/core/tencent_ssl.go    # 核心服务逻辑（真实腾讯云SDK集成）
internal/db/models.go           # 数据模型
config.example.json             # 配置文件示例
go.mod                          # 依赖管理（包含腾讯云SDK）
README_tencent_ssl.md           # 本文档
```

## 使用示例

1. **配置腾讯云凭证**：
   ```bash
   cp config.example.json config.json
   # 编辑 config.json，填入真实的腾讯云凭证
   ```

2. **启动服务**：
   ```bash
   go run cmd/server/main.go
   ```

3. **申请证书**：
   ```bash
   curl -X POST http://localhost:8080/api/certificates/tencent/apply \
     -H "Content-Type: application/json" \
     -d '{"domain":"example.com","validate_type":"DNS_AUTO","cert_alias":"测试证书"}'
   ```

4. **查询状态**：
   ```bash
   curl http://localhost:8080/api/certificates/tencent/cert_1234567890/status
   ```

5. **下载证书**（状态为"已通过"后）：
   ```bash
   curl -X POST http://localhost:8080/api/certificates/tencent/cert_1234567890/download
   ```

6. **检查证书是否需要续期**：
   ```bash
   curl http://localhost:8080/api/certificates/tencent/cert_1234567890/renewal-status?days=30
   ```

7. **续期单个证书**：
   ```bash
   curl -X POST http://localhost:8080/api/certificates/tencent/cert_1234567890/renew
   ```

8. **批量续期**（续期所有证书，无时间限制）：
   ```bash
   curl -X POST http://localhost:8080/api/certificates/tencent/batch-renew
   ```

9. **批量续期**（只续期30天内过期的证书）：
   ```bash
   curl -X POST http://localhost:8080/api/certificates/tencent/batch-renew?days=30
   ```