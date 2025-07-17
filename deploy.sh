#!/bin/bash

# Kiro2API 部署脚本

set -e

echo "🚀 开始部署 Kiro2API..."

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    exit 1
fi

# 检查Docker Compose是否安装
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose 未安装，请先安装 Docker Compose"
    exit 1
fi

# 停止现有服务
echo "🛑 停止现有服务..."
docker-compose down || true

# 构建镜像
echo "🔨 构建 Docker 镜像..."
docker-compose build

# 启动服务
echo "▶️ 启动服务..."
docker-compose up -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 5

# 检查服务状态
echo "🔍 检查服务状态..."
docker-compose ps

# 测试健康检查
echo "🏥 测试健康检查..."
if curl -s http://localhost:8080/health | grep -q "ok"; then
    echo "✅ 服务启动成功！"
    echo "📡 API 端点: http://localhost:8080/v1/chat/completions"
    echo "📋 模型列表: http://localhost:8080/v1/models"
    echo "💡 使用方法: 在 OpenAI 客户端的 api_key 中设置你的 Kiro access token"
else
    echo "❌ 服务启动失败，请检查日志:"
    docker-compose logs
    exit 1
fi

echo "🎉 部署完成！"