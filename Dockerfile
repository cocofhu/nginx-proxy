# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装必要的包（不需要 gcc 和 sqlite-dev，使用纯 Go 驱动）
RUN apk add --no-cache git make

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码和 Makefile
COPY . .

# 使用 make 构建应用（纯 Go 构建，无需 CGO）
RUN make build-no-cgo

# 运行阶段 - 基于官方 nginx 镜像
FROM nginx:alpine

# 安装必要的包
RUN apk --no-cache add ca-certificates curl

# 复制构建的二进制文件
COPY --from=builder /app/bin/nginx-proxy /usr/local/bin/nginx-proxy

# 创建必要的目录
RUN mkdir -p /app/data /app/config /app/template /etc/nginx/certs

# 复制默认配置文件和模板到默认位置
COPY config.json /app/config/config.json.default
COPY template/ /app/template/
COPY nginx.conf /etc/nginx/nginx.conf

# 设置工作目录
WORKDIR /app

# 创建启动脚本，支持配置文件挂载
RUN echo '#!/bin/sh' > /start.sh && \
    echo '# 如果没有挂载配置文件，使用默认配置' >> /start.sh && \
    echo 'if [ ! -f /app/config/config.json ]; then' >> /start.sh && \
    echo '  cp /app/config/config.json.default /app/config/config.json' >> /start.sh && \
    echo 'fi' >> /start.sh && \
    echo '# 修改配置文件中的数据库路径' >> /start.sh && \
    echo 'sed -i "s|\"./nginx-proxy.db\"|\"./data/nginx-proxy.db\"|g" /app/config/config.json' >> /start.sh && \
    echo '# 启动应用，指定配置文件路径' >> /start.sh && \
    echo 'nginx-proxy -config=/app/config/config.json &' >> /start.sh && \
    echo 'nginx -g "daemon off;"' >> /start.sh && \
    chmod +x /start.sh

# 定义可挂载的卷
VOLUME ["/app/data"]
# 数据库文件存储
VOLUME ["/etc/nginx/conf.d"]
# Nginx 配置文件目录
VOLUME ["/etc/nginx/certs"]
# SSL 证书存储目录
VOLUME ["/var/log/nginx"]
# Nginx 日志目录
VOLUME ["/app/template"]
# 模板文件目录（可自定义模板）
VOLUME ["/app/config"]
# 应用配置目录（而不是单个文件）

# 暴露端口
EXPOSE 80 443 8080

# 启动命令 - 同时启动 nginx-proxy 和 nginx
CMD ["/start.sh"]