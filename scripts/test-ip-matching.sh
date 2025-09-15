#!/bin/bash

# IP 分流测试脚本
# 用于测试不同 IP 格式的分流配置

echo "=== Nginx Proxy IP 分流测试 ==="

# 服务器地址
SERVER="http://localhost:8080"

echo "1. 测试单个 IP 分流配置..."
curl -X POST "$SERVER/api/rules" \
  -H "Content-Type: application/json" \
  -d @examples/single-ip-rule.json

echo -e "\n2. 测试 IP 段分流配置..."
curl -X POST "$SERVER/api/rules" \
  -H "Content-Type: application/json" \
  -d @examples/subnet-rule.json

echo -e "\n3. 重载 Nginx 配置..."
curl -X POST "$SERVER/api/reload"

echo -e "\n4. 查看生成的配置文件..."
echo "检查 /etc/nginx/conf.d/ 目录下的配置文件"

echo -e "\n5. 测试 Nginx 配置语法..."
nginx -t

echo -e "\n测试完成！"
echo "请检查生成的配置文件中的 IP 匹配条件是否正确。"