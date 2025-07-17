# Kiro2API - OpenAI API兼容的Kiro代理服务器

这是一个Go程序，将标准的OpenAI API请求转发到Kiro IDE后端的Claude模型。它提供了一个OpenAI API兼容的接口，让你可以使用任何支持OpenAI API的客户端来访问Kiro的Claude模型。

## 功能

- 🔄 将OpenAI API格式的请求转换为Kiro API格式
- 🌐 提供OpenAI API兼容的HTTP端点
- 🚀 支持Claude Sonnet 4和Claude 3.7 Sonnet模型
- 🔒 使用Kiro IDE提取的accessToken进行认证
- 📊 提供模型列表和健康检查端点
- 🌍 支持CORS，可在Web应用中使用
- 🐳 支持Docker和Docker Compose部署
- 📦 轻量级Alpine Linux镜像

## 编译

```bash
go build -o kiro2api main.go
```

## 使用方法

### 方式一：直接运行

#### 1. 启动服务器

```bash
# 使用默认端口8080
./kiro2api

# 指定自定义端口
./kiro2api 9000
```

### 方式二：使用Docker

#### 1. 构建Docker镜像

```bash
docker build -t kiro2api .
```

#### 2. 运行Docker容器

```bash
docker run -d -p 8080:8080 --name kiro2api kiro2api
```

#### 3. 使用Docker Compose（推荐）

```bash
docker-compose up -d
```

#### 4. 查看容器状态

```bash
docker ps
docker logs kiro2api
```

#### 5. 停止服务

```bash
docker-compose down
# 或者
docker stop kiro2api && docker rm kiro2api
```

## 在客户端中使用Kiro Token

不需要设置环境变量！直接在OpenAI客户端的`api_key`参数中使用你的Kiro access token。

## API端点

### 聊天完成 (OpenAI兼容)
```
POST /v1/chat/completions
```

### 模型列表
```
GET /v1/models
```

### 健康检查
```
GET /health
```

## 使用示例

### 使用curl

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-kiro-access-token-here" \
  -d '{
    "model": "claude-sonnet-4-20250514",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ],
    "max_tokens": 1000,
    "temperature": 0.7
  }'
```

### 使用Python OpenAI客户端

```python
import openai

# 直接在api_key参数中使用你的Kiro access token
client = openai.OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="your-kiro-access-token-here"  # 在这里设置你的Kiro token
)

response = client.chat.completions.create(
    model="claude-sonnet-4-20250514",
    messages=[
        {"role": "user", "content": "Hello, how are you?"}
    ],
    max_tokens=1000,
    temperature=0.7
)

print(response.choices[0].message.content)
```

### 使用JavaScript

```javascript
const response = await fetch('http://localhost:8080/v1/chat/completions', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer your-kiro-access-token-here'
  },
  body: JSON.stringify({
    model: 'claude-sonnet-4-20250514',
    messages: [
      { role: 'user', content: 'Hello, how are you?' }
    ],
    max_tokens: 1000,
    temperature: 0.7
  })
});

const data = await response.json();
console.log(data.choices[0].message.content);
```

## 支持的模型

- `claude-sonnet-4-20250514` (默认)
- `claude-3-7-sonnet-20250219`

## 认证方式

不需要设置环境变量！直接在OpenAI客户端的`api_key`参数中使用你的Kiro access token，或在HTTP请求的`Authorization`头中使用`Bearer your-token`格式。

## 获取Access Token

你可以从Kiro IDE中提取access token。通常可以通过以下方式获取：

1. 打开Kiro IDE的开发者工具
2. 查看网络请求中的Authorization头
3. 复制Bearer token后面的值

## 注意事项

- 这个代理服务器将请求转发到 `https://codewhisperer.us-east-1.amazonaws.com/generateAssistantResponse`
- Access token可能会过期，需要定期更新
- 服务器默认监听所有接口 (0.0.0.0)，支持CORS请求

## 故障排除

如果遇到认证错误，请检查：
1. 是否在OpenAI客户端的api_key参数中正确设置了Kiro access token
2. Access token是否已过期
3. 网络连接是否正常
4. Authorization头格式是否正确 (Bearer your-token)

## 项目文件

```
kiro2cc/
├── main.go              # 主程序文件
├── go.mod              # Go模块文件
├── Dockerfile          # Docker镜像构建文件
├── docker-compose.yml  # Docker Compose配置
├── .dockerignore       # Docker构建忽略文件
├── deploy.sh           # 一键部署脚本
└── README.md           # 项目说明文档
```

## 快速部署

使用提供的部署脚本可以一键部署：

```bash
chmod +x deploy.sh
./deploy.sh
```

## 许可证

MIT License
