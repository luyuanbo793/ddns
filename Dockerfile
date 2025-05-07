# 使用官方 Golang 镜像作为基础镜像
FROM golang:1.20-alpine AS builder

# 设置工作目录
WORKDIR /app

# 将 Go 源代码复制到容器中
COPY . .

# 构建 Go 程序，并指定输出文件名为 youddns
RUN go build -o youddns .

# 使用最小化的 Alpine 镜像运行程序
FROM alpine:latest

# 安装必要的工具
RUN apk add --no-cache bash

# 从 builder 阶段复制编译好的二进制文件，并重命名为 youddns
COPY --from=builder /app/youddns /usr/local/bin/youddns

# 设置默认环境变量
ENV DOMAIN="test.luqh.dpdns.org" \
    TOKEN="fbdd26343facbf9ac81538c32328e721" \
    API_URL="https://dns.cngames.site/ddnsapi.php"

# 启动命令
CMD ["youddns"]