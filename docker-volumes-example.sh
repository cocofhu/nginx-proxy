#!/bin/bash

# Docker 卷挂载示例脚本
# 演示如何挂载各种路径进行数据持久化和配置管理

echo "=== Nginx Proxy 容器卷挂载示例 ==="

# 创建本地目录
echo "创建本地目录..."
mkdir -p ./volumes/data
mkdir -p ./volumes/nginx-conf
mkdir -p ./volumes/nginx-certs
mkdir -p ./volumes/nginx-logs
mkdir -p ./volumes/templates
mkdir -p ./volumes/config

# 复制默认配置文件到挂载目录
echo "复制默认配置..."
cp config.json ./volumes/config/
cp -r template/* ./volumes/templates/ 2>/dev/null || echo "模板目录为空，跳过复制"

# 示例 1: 基本挂载
echo ""
echo "=== 示例 1: 基本数据持久化 ==="
echo "docker run -d \\"
echo "  --name nginx-proxy-basic \\"
echo "  -p 80:80 \\"
echo "  -p 443:443 \\"
echo "  -p 8080:8080 \\"
echo "  -v \$(pwd)/volumes/data:/app/data \\"
echo "  -v \$(pwd)/volumes/nginx-conf:/etc/nginx/conf.d \\"
echo "  -v \$(pwd)/volumes/nginx-certs:/etc/nginx/certs \\"
echo "  nginx-proxy"

# 示例 2: 完整挂载（包含日志和配置）
echo ""
echo "=== 示例 2: 完整挂载（推荐生产环境） ==="
echo "docker run -d \\"
echo "  --name nginx-proxy-full \\"
echo "  -p 80:80 \\"
echo "  -p 443:443 \\"
echo "  -p 8080:8080 \\"
echo "  -v \$(pwd)/volumes/data:/app/data \\"
echo "  -v \$(pwd)/volumes/nginx-conf:/etc/nginx/conf.d \\"
echo "  -v \$(pwd)/volumes/nginx-certs:/etc/nginx/certs \\"
echo "  -v \$(pwd)/volumes/nginx-logs:/var/log/nginx \\"
echo "  -v \$(pwd)/volumes/templates:/app/template \\"
echo "  -v \$(pwd)/volumes/config/config.json:/app/config.json \\"
echo "  nginx-proxy"

# 示例 3: 只读挂载（安全配置）
echo ""
echo "=== 示例 3: 只读挂载（安全配置） ==="
echo "docker run -d \\"
echo "  --name nginx-proxy-secure \\"
echo "  -p 80:80 \\"
echo "  -p 443:443 \\"
echo "  -p 8080:8080 \\"
echo "  -v \$(pwd)/volumes/data:/app/data \\"
echo "  -v \$(pwd)/volumes/nginx-conf:/etc/nginx/conf.d \\"
echo "  -v \$(pwd)/volumes/nginx-certs:/etc/nginx/certs:ro \\"
echo "  -v \$(pwd)/volumes/nginx-logs:/var/log/nginx \\"
echo "  -v \$(pwd)/volumes/templates:/app/template:ro \\"
echo "  -v \$(pwd)/volumes/config/config.json:/app/config.json:ro \\"
echo "  nginx-proxy"

# 示例 4: 使用命名卷
echo ""
echo "=== 示例 4: 使用 Docker 命名卷 ==="
echo "# 创建命名卷"
echo "docker volume create nginx-proxy-data"
echo "docker volume create nginx-proxy-certs"
echo "docker volume create nginx-proxy-logs"
echo ""
echo "docker run -d \\"
echo "  --name nginx-proxy-volumes \\"
echo "  -p 80:80 \\"
echo "  -p 443:443 \\"
echo "  -p 8080:8080 \\"
echo "  -v nginx-proxy-data:/app/data \\"
echo "  -v \$(pwd)/volumes/nginx-conf:/etc/nginx/conf.d \\"
echo "  -v nginx-proxy-certs:/etc/nginx/certs \\"
echo "  -v nginx-proxy-logs:/var/log/nginx \\"
echo "  nginx-proxy"

echo ""
echo "=== 卷说明 ==="
echo "/app/data              - SQLite 数据库文件存储"
echo "/etc/nginx/conf.d      - 动态生成的 Nginx 配置文件"
echo "/etc/nginx/certs       - SSL 证书和私钥文件"
echo "/var/log/nginx         - Nginx 访问日志和错误日志"
echo "/app/template          - Nginx 配置模板文件（可自定义）"
echo "/app/config.json       - 应用程序配置文件"

echo ""
echo "=== 使用建议 ==="
echo "1. 生产环境建议挂载所有卷以确保数据持久化"
echo "2. 证书和模板文件可以设置为只读挂载提高安全性"
echo "3. 日志目录建议挂载以便日志分析和监控"
echo "4. 使用命名卷可以更好地管理数据生命周期"

echo ""
echo "脚本执行完成！请根据需要选择合适的挂载方式。"