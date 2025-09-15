#!/bin/bash

# 部署脚本

set -e

# 配置
ENVIRONMENT=${1:-staging}
VERSION=${2:-latest}
DOCKER_REGISTRY=${DOCKER_REGISTRY:-ccr.ccs.tencentyun.com}

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}Deploying nginx-proxy to ${ENVIRONMENT} environment${NC}"
echo -e "${YELLOW}Version: ${VERSION}${NC}"
echo -e "${YELLOW}Registry: ${DOCKER_REGISTRY}${NC}"

# 检查环境
case $ENVIRONMENT in
    staging|production)
        echo -e "${GREEN}Valid environment: ${ENVIRONMENT}${NC}"
        ;;
    *)
        echo -e "${RED}Invalid environment. Use: staging or production${NC}"
        exit 1
        ;;
esac

# 设置环境变量
export VERSION=$VERSION
export DOCKER_REGISTRY=$DOCKER_REGISTRY

# 创建必要的目录
echo -e "${YELLOW}Creating necessary directories...${NC}"
mkdir -p data nginx-conf nginx-certs logs

# 停止现有服务
echo -e "${YELLOW}Stopping existing services...${NC}"
docker-compose -f docker-compose.prod.yml down || true

# 拉取最新镜像
echo -e "${YELLOW}Pulling latest images...${NC}"
docker-compose -f docker-compose.prod.yml pull

# 启动服务
echo -e "${YELLOW}Starting services...${NC}"
docker-compose -f docker-compose.prod.yml up -d

# 等待服务启动
echo -e "${YELLOW}Waiting for services to start...${NC}"
sleep 30

# 健康检查
echo -e "${YELLOW}Performing health check...${NC}"
if curl -f http://localhost:8080/api/rules > /dev/null 2>&1; then
    echo -e "${GREEN}✅ nginx-proxy is healthy${NC}"
else
    echo -e "${RED}❌ nginx-proxy health check failed${NC}"
    echo -e "${YELLOW}Checking logs...${NC}"
    docker-compose -f docker-compose.prod.yml logs nginx-proxy
    exit 1
fi

# 显示状态
echo -e "${GREEN}✅ Deployment completed successfully!${NC}"
echo -e "${YELLOW}Services status:${NC}"
docker-compose -f docker-compose.prod.yml ps

echo -e "${YELLOW}API endpoint: http://localhost:8080/api/rules${NC}"
echo -e "${YELLOW}To view logs: docker-compose -f docker-compose.prod.yml logs -f${NC}"