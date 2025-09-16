# 纯 Go 构建 Dockerfile - 无 CGO 依赖
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 只安装必要的包
RUN apk add --no-cache git make

# 复制依赖文件
COPY go.mod go.sum ./

RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct

RUN go mod download

# 复制源代码
COPY . .

# 纯 Go 构建，完全禁用 CGO
RUN make build

# 运行阶段 - 使用包含开发工具的 OpenResty 镜像
FROM openresty/openresty:alpine-fat

# 安装必要工具和 Lua 模块
RUN apk --no-cache add ca-certificates curl \
    && opm install ledgetech/lua-resty-http \
    && opm install openresty/lua-cjson

# 复制构建的二进制文件
COPY --from=builder /app/bin/nginx-proxy /usr/local/bin/nginx-proxy

# 创建必要的目录
RUN mkdir -p /app/data /app/config /app/template /app/web/static /etc/nginx/certs \
    /var/log/nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx /var/cache/nginx

# 复制默认配置和模板
COPY config.json /app/config/config.json.default
COPY template/ /app/template/

# 复制优化的 Nginx 配置文件
COPY nginx-docker.conf /etc/nginx/nginx.conf

# 复制静态文件
COPY web/static/ /app/web/static/

# 设置工作目录
WORKDIR /app

# 创建启动脚本
RUN echo '#!/bin/sh' > /start.sh && \
    echo 'echo "启动 nginx-proxy 服务（纯 Go 版本）..."' >> /start.sh && \
    echo '# 检查配置文件' >> /start.sh && \
    echo 'if [ ! -f /app/config/config.json ]; then' >> /start.sh && \
    echo '  echo "复制默认配置文件..."' >> /start.sh && \
    echo '  cp /app/config/config.json.default /app/config/config.json' >> /start.sh && \
    echo 'fi' >> /start.sh && \
    echo '# 修改数据库路径' >> /start.sh && \
    echo 'sed -i "s|\"./nginx-proxy.db\"|\"./data/nginx-proxy.db\"|g" /app/config/config.json' >> /start.sh && \
    echo '# 启动 nginx-proxy（后台）' >> /start.sh && \
    echo 'nginx-proxy -config=/app/config/config.json &' >> /start.sh && \
    echo '# 启动 OpenResty（前台）' >> /start.sh && \
    echo '/usr/local/openresty/bin/openresty -g "daemon off;"' >> /start.sh && \
    chmod +x /start.sh

# 定义卷
VOLUME ["/app/data", "/etc/nginx/conf.d", "/etc/nginx/certs", "/var/log/nginx", "/app/template", "/app/config", "/app/web/static"]

# 暴露端口
EXPOSE 80 443 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/api/rules || exit 1

# 启动命令
CMD ["/start.sh"]