#!/bin/bash

# Kiro2API 启动脚本
# 支持两种模式：
# 1. OpenAI API兼容模式 (默认)
# 2. Anthropic API代理模式

MODE=${1:-"openai"}
PORT=${2:-"8080"}

echo "Kiro2API 启动脚本"
echo "=================="

case $MODE in
    "openai")
        echo "启动模式: OpenAI API兼容代理"
        echo "端点: http://localhost:$PORT/v1/chat/completions"
        echo "使用方法: 在OpenAI客户端中设置base_url和api_key"
        echo ""
        go run main.go $PORT
        ;;
    "anthropic")
        echo "启动模式: Anthropic API直接代理"
        echo "端点: http://localhost:$PORT"
        echo "环境变量:"
        echo "  ANTHROPIC_BASE_URL: ${ANTHROPIC_BASE_URL:-https://codewhisperer.us-east-1.amazonaws.com/generateAssistantResponse}"
        echo "  ANTHROPIC_API_KEY: ${ANTHROPIC_API_KEY:-未设置}"
        echo ""
        go run anthropic_proxy.go $PORT
        ;;
    *)
        echo "错误: 未知模式 '$MODE'"
        echo "用法: $0 [openai|anthropic] [port]"
        echo ""
        echo "模式说明:"
        echo "  openai    - OpenAI API兼容模式 (默认)"
        echo "  anthropic - Anthropic API直接代理模式"
        echo ""
        echo "示例:"
        echo "  $0 openai 8080     # 启动OpenAI兼容代理"
        echo "  $0 anthropic 8080  # 启动Anthropic直接代理"
        exit 1
        ;;
esac