#!/bin/bash

# Docker 构建和测试脚本
set -e

echo "🐳 开始构建 Docker 镜像..."
docker build -t nginx-proxy:test .

echo "🔍 检查镜像中的静态文件..."
docker run --rm nginx-proxy:test ls -la /app/web/static/

echo "📁 检查静态文件内容..."
docker run --rm nginx-proxy:test find /app/web/static -type f -exec ls -lh {} \;

echo "🧪 启动容器进行快速测试..."
CONTAINER_ID=$(docker run -d -p 8080:8080 nginx-proxy:test)

echo "⏳ 等待服务启动..."
sleep 5

echo "🌐 测试静态文件访问..."
if curl -f http://localhost:8080/ > /dev/null 2>&1; then
    echo "✅ 静态文件访问成功！"
else
    echo "❌ 静态文件访问失败！"
fi

echo "🔧 测试 API 接口..."
if curl -f http://localhost:8080/api/health > /dev/null 2>&1; then
    echo "✅ API 接口访问成功！"
else
    echo "❌ API 接口访问失败！"
fi

echo "🧹 清理测试容器..."
docker stop $CONTAINER_ID
docker rm $CONTAINER_ID

echo "🎉 Docker 测试完成！"