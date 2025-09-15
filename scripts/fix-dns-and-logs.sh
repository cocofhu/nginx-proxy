#!/bin/bash

# DNS 解析和日志问题修复脚本

echo "=== Nginx DNS 解析和日志问题修复 ==="

# 检查是否在 Docker 环境中
if [ -f /.dockerenv ]; then
    echo "检测到 Docker 环境"
    DOCKER_ENV=true
else
    echo "检测到宿主机环境"
    DOCKER_ENV=false
fi

echo "1. 备份当前配置..."
cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.backup.$(date +%Y%m%d_%H%M%S)

echo "2. 修复 DNS 解析问题..."
if [ "$DOCKER_ENV" = true ]; then
    echo "使用 Docker 环境配置"
    cp nginx-docker.conf /etc/nginx/nginx.conf
else
    echo "使用宿主机环境配置"
    cp nginx.conf /etc/nginx/nginx.conf
fi

echo "3. 创建日志目录和缓存目录..."
mkdir -p /var/log/nginx
mkdir -p /var/cache/nginx
chown -R nginx:nginx /var/log/nginx /var/cache/nginx

echo "4. 测试 DNS 解析..."
echo "测试解析 git.service.arpa:"
if command -v nslookup >/dev/null 2>&1; then
    nslookup git.service.arpa
elif command -v dig >/dev/null 2>&1; then
    dig git.service.arpa
else
    echo "未找到 DNS 查询工具，请手动检查 DNS 解析"
fi

echo "5. 检查 Nginx 配置语法..."
nginx -t

if [ $? -eq 0 ]; then
    echo "✅ Nginx 配置语法正确"
    
    echo "6. 重载 Nginx 配置..."
    nginx -s reload
    
    echo "7. 检查 Nginx 进程..."
    ps aux | grep nginx
    
    echo "8. 测试日志写入..."
    echo "测试访问日志写入..." > /var/log/nginx/test.log
    if [ -f /var/log/nginx/test.log ]; then
        echo "✅ 日志目录可写"
        rm /var/log/nginx/test.log
    else
        echo "❌ 日志目录不可写"
    fi
    
    echo "9. 显示当前日志文件状态..."
    ls -la /var/log/nginx/
    
    echo "10. 测试日志读取..."
    echo "最近的访问日志（最后 5 行）："
    tail -n 5 /var/log/nginx/access.log 2>/dev/null || echo "访问日志为空或不存在"
    
    echo "最近的错误日志（最后 5 行）："
    tail -n 5 /var/log/nginx/error.log 2>/dev/null || echo "错误日志为空或不存在"
    
else
    echo "❌ Nginx 配置语法错误，请检查配置文件"
    exit 1
fi

echo ""
echo "=== 修复完成 ==="
echo ""
echo "📋 解决方案说明："
echo "1. DNS 解析问题："
echo "   - 添加了 Docker 内置 DNS (127.0.0.11) 和公共 DNS"
echo "   - 设置了合理的解析超时时间"
echo "   - 使用变量方式进行动态域名解析"
echo ""
echo "2. 日志卡顿问题："
echo "   - 启用了日志缓冲 (buffer=64k flush=1s)"
echo "   - 使用异步日志写入"
echo "   - 优化了日志格式"
echo ""
echo "🧪 测试建议："
echo "1. 测试域名解析："
echo "   curl -H 'Host: fff.com' http://localhost/"
echo ""
echo "2. 实时查看日志："
echo "   tail -f /var/log/nginx/access.log"
echo "   tail -f /var/log/nginx/error.log"
echo ""
echo "3. 检查分流效果："
echo "   从 192.168.2.45 访问应该路由到 git.service.arpa"
echo "   从其他 IP 访问应该路由到 192.168.2.1"