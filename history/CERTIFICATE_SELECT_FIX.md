# 证书选择功能修复

## 问题描述

用户反馈：在添加代理时无法选择证书，证书下拉框为空或显示"暂无证书"。

## 问题分析

原始代码中存在以下问题：

1. **数据加载时机错误**：`loadCertificatesForSelect()`函数依赖`currentCertificates`变量，但该变量只在访问证书管理页面时才被加载
2. **异步加载未处理**：证书数据加载是异步的，但界面更新是同步的
3. **调试信息不足**：无法确定是API问题还是前端逻辑问题

## 修复方案

### 1. 添加专门的证书数据加载函数

```javascript
function loadCertificatesData() {
    return Promise.all([
        fetch('/api/certificates/tencent/list').then(response => response.json()),
        fetch('/api/certificates').then(response => response.json())
    ])
    .then(([tencentData, uploadData]) => {
        // 合并证书数据
        const allCertificates = [
            ...tencentCerts.map(cert => ({...cert, source: 'tencent_cloud'})),
            ...uploadCerts.filter(cert => cert.source === 'upload').map(cert => ({...cert, source: 'upload'}))
        ];
        currentCertificates = allCertificates;
        return allCertificates;
    });
}
```

### 2. 修复添加代理按钮事件

```javascript
addProxyBtn.addEventListener('click', function () {
    // 显示模态框
    document.getElementById('add-proxy-modal').classList.remove('hidden');
    document.getElementById('add-proxy-modal').classList.add('flex');
    
    // 异步加载证书数据
    loadCertificatesData().then(() => {
        loadCertificatesForSelect();
    });
});
```

### 3. 增强SSL复选框事件

```javascript
sslCheckbox.addEventListener('change', function () {
    if (this.checked) {
        sslConfig.classList.remove('hidden');
        // 确保证书数据是最新的
        if (currentCertificates && currentCertificates.length > 0) {
            loadCertificatesForSelect();
        } else {
            loadCertificatesData().then(() => {
                loadCertificatesForSelect();
            });
        }
    }
});
```

### 4. 添加详细的调试信息

- 控制台日志显示证书加载过程
- 区分有效证书和无效证书
- 显示具体的错误信息

## 调试工具

创建了 `debug_certificates.html` 调试页面，可以：

- 测试腾讯云证书API
- 测试上传证书API
- 测试证书数据合并逻辑
- 查看详细的API响应

## 使用方法

### 调试步骤：

1. 访问 `http://localhost:8080/debug_certificates.html`
2. 点击各个测试按钮
3. 查看API响应和数据格式
4. 确认证书数据是否正确

### 正常使用：

1. 确保已有可用证书（上传或申请腾讯云证书）
2. 点击"添加代理"按钮
3. 启用SSL选项
4. 从下拉框选择证书

## 预期结果

修复后，证书选择功能应该：

- ✅ 自动加载所有可用证书
- ✅ 显示证书名称和域名
- ✅ 区分腾讯云证书和上传证书
- ✅ 只显示有效的证书（有cert_path的）
- ✅ 提供清晰的状态提示

## 故障排除

### 如果仍然无法选择证书：

1. **检查证书是否存在**：
    - 访问证书管理页面
    - 确认有可用的证书

2. **检查证书状态**：
    - 腾讯云证书需要是"已通过"状态
    - 上传证书需要有有效的文件路径

3. **检查API响应**：
    - 使用调试页面测试API
    - 查看浏览器控制台的错误信息

4. **检查证书文件**：
    - 确认cert_path和key_path字段不为空
    - 确认证书文件实际存在

现在证书选择功能应该可以正常工作了！