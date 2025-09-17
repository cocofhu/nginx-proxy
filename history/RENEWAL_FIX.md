# 证书续期问题修复

## 问题描述

用户反馈：点击续期后等待成功，但检查时发现没有更新到新证书，再点击检查说证书不存在。

## 根本原因

在 `handleRenewalCompletion` 函数中，代码错误地将 `originalCert.SourceID = newCertID`，这导致：

1. 原始证书ID失效
2. 用户无法通过原始ID查询证书
3. 证书记录的连续性被破坏

## 修复方案

1. **保持原始证书ID不变**：不修改 `SourceID` 字段
2. **只更新证书文件和内容**：更新文件路径、过期时间等
3. **保持证书记录连续性**：用户始终可以通过原始ID查询证书
4. **不删除腾讯云端老证书**：避免ID失效问题

## 修复内容

### 修复前（有问题的代码）：

```go
originalCert.SourceID = newCertID // 这行导致原始ID失效
```

### 修复后：

```go
// 不更改SourceID，保持证书记录的连续性
// 只更新文件路径和状态
originalCert.CertPath = newCertPath
originalCert.KeyPath = newKeyPath
originalCert.Status = "active"
originalCert.RenewalSourceID = ""
```

## 测试验证

续期完成后：

- ✅ 原始证书ID仍然有效
- ✅ 证书状态显示"已通过"
- ✅ 证书文件已更新为新内容
- ✅ 过期时间已更新
- ✅ 查询不会返回"证书不存在"

## 影响范围

- 修复续期功能的ID连续性问题
- 不影响其他功能
- 向后兼容现有证书记录