# 使用官方 Go 镜像作为构建环境
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY main.go ./

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o kiro2api main.go

# 使用轻量级的 alpine 作为运行环境
FROM alpine:latest

# 安装 ca-certificates 用于 HTTPS 请求
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/kiro2api .

# 设置环境变量
ENV PORT=8080

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./kiro2api", "8080"]