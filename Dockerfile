# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# 运行阶段
FROM alpine:latest

# 安装 nginx 和其他必要的包
RUN apk --no-cache add nginx ca-certificates sqlite

# 创建必要的目录
RUN mkdir -p /etc/nginx/conf.d /etc/nginx/certs /var/log/nginx /var/lib/nginx/tmp

# 创建 nginx 用户
RUN adduser -D -s /bin/sh nginx

# 复制构建的二进制文件
COPY --from=builder /app/main /usr/local/bin/nginx-proxy

# 复制配置文件和模板
COPY config.json /app/
COPY template/ /app/template/

# 设置工作目录
WORKDIR /app

# 创建数据目录
RUN mkdir -p /app/data

# 修改配置文件中的路径
RUN sed -i 's|"./nginx-proxy.db"|"/app/data/nginx-proxy.db"|g' config.json

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["nginx-proxy"]