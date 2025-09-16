# 代理配置编辑功能测试

## 功能概述

已完成的编辑功能包括：

### 1. 前端界面

- ✅ 编辑模态框 (`edit-proxy-modal`)
- ✅ 表单字段：域名、路径、分流配置、SSL设置
- ✅ 动态分流规则添加/删除
- ✅ 证书选择下拉框

### 2. JavaScript功能

- ✅ `editRule(ruleId)` - 主编辑函数
- ✅ `showEditProxyModal()` - 显示编辑模态框
- ✅ `closeEditProxyModal()` - 关闭编辑模态框
- ✅ `populateEditForm(rule)` - 填充编辑表单
- ✅ `addEditUpstreamConfig()` - 添加编辑分流配置
- ✅ `loadCertificatesForEditSelect()` - 加载证书到选择框
- ✅ `handleEditProxy(e)` - 处理编辑提交

### 3. 后端API

- ✅ `PUT /api/rules/{id}` - 更新规则接口
- ✅ `UpdateRule` 处理器已实现
- ✅ 路由配置已完成

## 测试步骤

### 1. 基本编辑测试

1. 启动应用：`make run` 或 `./start.sh`
2. 访问：http://localhost:8080
3. 进入"代理配置"页面
4. 点击任意规则的"编辑"按钮
5. 验证编辑模态框正确显示
6. 验证表单字段正确填充

### 2. 表单功能测试

1. 修改域名字段
2. 修改路径字段
3. 添加/删除分流规则
4. 切换SSL开关
5. 选择不同证书
6. 点击"更新"按钮

### 3. API测试

```bash
# 获取规则列表
curl http://localhost:8080/api/rules

# 获取特定规则
curl http://localhost:8080/api/rules/{rule-id}

# 更新规则
curl -X PUT http://localhost:8080/api/rules/{rule-id} \
  -H "Content-Type: application/json" \
  -d '{
    "server_name": "updated.example.com",
    "listen_ports": [80],
    "ssl_cert": "",
    "ssl_key": "",
    "locations": [{
      "path": "/",
      "upstreams": [{
        "condition_ip": "0.0.0.0/0",
        "target": "http://updated-backend:8080"
      }]
    }]
  }'
```

## 预期结果

### 成功场景

- 编辑模态框正确显示
- 表单字段正确填充现有数据
- 分流规则正确显示和编辑
- SSL配置正确显示和切换
- 证书选择正确加载
- 更新成功后模态框关闭
- 代理列表自动刷新显示更新后的数据

### 错误处理

- 无效输入显示错误提示
- 网络错误显示错误消息
- 空分流规则显示警告
- 证书加载失败显示错误

## 已知限制

1. 当前只支持编辑第一个location的配置
2. 多location支持需要进一步开发
3. 高级Nginx配置选项暂未包含

## 下一步改进

1. 支持多location编辑
2. 添加配置验证
3. 支持更多Nginx配置选项
4. 添加配置预览功能