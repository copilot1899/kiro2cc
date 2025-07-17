# Kiro2API - 双模式API代理服务器

这是一个Go程序，提供两种模式来访问Kiro IDE后端的Claude模型：
1. **OpenAI API兼容模式** - 将OpenAI API请求转换为Kiro API格式
2. **Anthropic API直接代理模式** - 直接代理Anthropic API请求到Kiro后端

## 功能特点

- 🔄 **双模式支持**: OpenAI兼容模式和Anthropic直接代理模式
- 🔑 **灵活认证**: 支持环境变量和请求头认证
- 🚀 **即插即用**: 无需修改现有代码，只需更改配置
- 🐳 **Docker支持**: 提供完整的Docker部署方案
- 🌐 **CORS支持**: 支持跨域请求，可在Web应用中使用
- 📊 **健康检查**: 提供健康检查端点

## 快速开始

### 方式1: 使用启动脚本

```bash
# OpenAI API兼容模式 (默认)
./start.sh openai 8080

# Anthropic API直接代理模式
export ANTHROPIC_BASE_URL="https://codewhisperer.us-east-1.amazonaws.com/generateAssistantResponse"
export ANTHROPIC_API_KEY="your-kiro-access-token"
./start.sh anthropic 8080
```

### 方式2: 直接运行

```bash
# OpenAI兼容模式
go run main.go 8080

# Anthropic代理模式
export ANTHROPIC_BASE_URL="https://codewhisperer.us-east-1.amazonaws.com/generateAssistantResponse"
export ANTHROPIC_API_KEY="your-kiro-access-token"
go run anthropic_proxy.go 8080
```

## 使用模式详解

### 模式1: OpenAI API兼容模式

**特点:**
- 提供标准的OpenAI API端点
- 支持现有的OpenAI客户端库
- 自动转换请求和响应格式

**端点:**
- `POST /v1/chat/completions` - 聊天完成
- `GET /v1/models` - 模型列表
- `GET /health` - 健康检查

**使用示例:**

```bash
# 使用curl
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-kiro-access-token" \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ],
    "max_tokens": 1000
  }'
```

```python
# 使用Python OpenAI客户端
import openai

client = openai.OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="your-kiro-access-token"
)

response = client.chat.completions.create(
    model="claude-sonnet-4-20250514",
    messages=[{"role": "user", "content": "Hello!"}],
    max_tokens=1000
)
```

### 模式2: Anthropic API直接代理模式

**特点:**
- 直接代理Anthropic API请求
- 支持环境变量配置
- 更接近原生Anthropic API体验

**环境变量:**
- `ANTHROPIC_BASE_URL`: 目标API端点 (默认: https://codewhisperer.us-east-1.amazonaws.com/generateAssistantResponse)
- `ANTHROPIC_API_KEY`: Kiro access token

**端点:**
- `POST /` - 直接代理请求
- `GET /health` - 健康检查

**使用示例:**

```bash
# 设置环境变量
export ANTHROPIC_BASE_URL="https://codewhisperer.us-east-1.amazonaws.com/generateAssistantResponse"
export ANTHROPIC_API_KEY="your-kiro-access-token"

# 发送请求
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-opus-20240229",
    "messages": [
      {"role": "user", "content": "Hello"}
    ],
    "max_tokens": 50
  }'
```

```python
# 使用Anthropic客户端
import anthropic

client = anthropic.Anthropic(
    base_url="http://localhost:8080",
    api_key="your-kiro-access-token"  # 或从环境变量读取
)

response = client.messages.create(
    model="claude-3-opus-20240229",
    messages=[{"role": "user", "content": "Hello"}],
    max_tokens=50
)
```

## Docker部署

### 构建镜像

```bash
docker build -t kiro2api .
```

### 运行容器

```bash
# OpenAI兼容模式
docker run -p 8080:8080 kiro2api

# Anthropic代理模式
docker run -p 8080:8080 \
  -e ANTHROPIC_BASE_URL="https://codewhisperer.us-east-1.amazonaws.com/generateAssistantResponse" \
  -e ANTHROPIC_API_KEY="your-kiro-access-token" \
  kiro2api anthropic
```

### 使用Docker Compose

```bash
# 编辑docker-compose.yml中的环境变量
docker-compose up -d
```

## 支持的模型

- `claude-sonnet-4-20250514` (推荐)
- `claude-3-7-sonnet-20250219`
- `claude-3-opus-20240229`

## 获取Access Token

从Kiro IDE中提取access token：

1. 打开Kiro IDE的开发者工具
2. 查看网络请求中的Authorization头
3. 复制Bearer token后面的值

## 故障排除

### 常见问题

1. **认证失败**
   - 确保token未过期
   - 检查token格式是否正确
   - 验证token权限

2. **连接问题**
   - 检查网络连接
   - 确认API端点是否正确
   - 验证防火墙设置

### 已知问题

⚠️ **当前状态**: 代码已完成，但存在认证问题：

- Token认证可能失败
- 需要验证正确的API端点和权限
- 建议从Kiro IDE抓包获取正确的API调用格式

## 项目结构

```
kiro2cc/
├── main.go              # OpenAI兼容模式主程序
├── anthropic_proxy.go   # Anthropic代理模式程序
├── start.sh            # 启动脚本
├── go.mod              # Go模块文件
├── Dockerfile          # Docker镜像构建文件
├── docker-compose.yml  # Docker Compose配置
├── deploy.sh           # 部署脚本
└── README.md           # 项目说明文档
```

## 开发

### 编译

```bash
# OpenAI兼容模式
go build -o kiro2api main.go

# Anthropic代理模式
go build -o anthropic-proxy anthropic_proxy.go
```

### 测试

```bash
# 测试健康检查
curl http://localhost:8080/health

# 测试OpenAI模式
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{"model": "claude-sonnet-4-20250514", "messages": [{"role": "user", "content": "test"}]}'

# 测试Anthropic模式
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{"model": "claude-3-opus-20240229", "messages": [{"role": "user", "content": "test"}]}'
```

## 许可证

MIT License