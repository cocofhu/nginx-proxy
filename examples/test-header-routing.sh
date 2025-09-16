#!/bin/bash

# 测试头部路由功能的脚本
# 假设nginx-proxy运行在localhost:8080

BASE_URL="http://localhost:8080"

echo "=== 测试头部路由功能 ==="

echo "1. 测试 X-API-Version: v1"
curl -H "X-API-Version: v1" "$BASE_URL/api/v1/test" -v

echo -e "\n2. 测试 X-API-Version: v2"  
curl -H "X-API-Version: v2" "$BASE_URL/api/v1/test" -v

echo -e "\n3. 测试默认路由（无头部）"
curl "$BASE_URL/api/v1/test" -v

echo -e "\n4. 测试移动端User-Agent"
curl -H "User-Agent: Mobile" "$BASE_URL/api/mobile/test" -v

echo -e "\n5. 测试混合条件（IP + 头部）"
curl -H "X-Internal: true" "$BASE_URL/api/mobile/test" -v

echo -e "\n=== 测试完成 ==="