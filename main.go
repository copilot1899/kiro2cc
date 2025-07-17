package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// OpenAI API 结构定义
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type OpenAIChoice struct {
	Index   int           `json:"index"`
	Message OpenAIMessage `json:"message"`
}

type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAIResponse struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage   `json:"usage"`
}

// Kiro API 结构定义
type KiroMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type KiroRequest struct {
	ConversationId string        `json:"conversationId,omitempty"`
	Messages       []KiroMessage `json:"messages"`
	Model          string        `json:"model,omitempty"`
	MaxTokens      int           `json:"maxTokens,omitempty"`
	Temperature    float64       `json:"temperature,omitempty"`
	Stream         bool          `json:"stream,omitempty"`
}

type KiroResponse struct {
	ConversationId string `json:"conversationId,omitempty"`
	Message        string `json:"message,omitempty"`
	Content        string `json:"content,omitempty"`
	Text           string `json:"text,omitempty"`
	Model          string `json:"model,omitempty"`
}

func main() {
	port := "8080" // 默认端口
	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	
	startServer(port)
}

// convertOpenAIToKiro 将OpenAI API请求转换为Kiro API请求
func convertOpenAIToKiro(openaiReq OpenAIRequest) KiroRequest {
	kiroMessages := make([]KiroMessage, len(openaiReq.Messages))
	for i, msg := range openaiReq.Messages {
		kiroMessages[i] = KiroMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	return KiroRequest{
		ConversationId: fmt.Sprintf("conv_%d", time.Now().UnixNano()),
		Messages:       kiroMessages,
		Model:          "claude-sonnet-4-20250514", // 默认使用Claude Sonnet 4
		MaxTokens:      openaiReq.MaxTokens,
		Temperature:    openaiReq.Temperature,
	}
}

// convertKiroToOpenAI 将Kiro API响应转换为OpenAI API响应
func convertKiroToOpenAI(kiroResp KiroResponse, originalModel string) OpenAIResponse {
	// 尝试从不同的字段获取响应内容
	content := kiroResp.Message
	if content == "" {
		content = kiroResp.Content
	}
	if content == "" {
		content = kiroResp.Text
	}
	
	return OpenAIResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   originalModel,
		Choices: []OpenAIChoice{
			{
				Index: 0,
				Message: OpenAIMessage{
					Role:    "assistant",
					Content: content,
				},
			},
		},
		Usage: OpenAIUsage{
			PromptTokens:     100, // 估算值，实际应该从Kiro响应中获取
			CompletionTokens: 200, // 估算值
			TotalTokens:      300, // 估算值
		},
	}
}

// startServer 启动OpenAI API兼容的代理服务器
func startServer(port string) {
	// 设置CORS中间件
	corsHandler := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next(w, r)
		}
	}

	// OpenAI API兼容的聊天完成端点
	http.HandleFunc("/v1/chat/completions", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "只支持POST请求", http.StatusMethodNotAllowed)
			return
		}

		// 从Authorization头中提取accessToken
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			errorMsg := map[string]interface{}{
				"error": map[string]interface{}{
					"message": "缺少Authorization头。请在OpenAI客户端的api_key参数中设置你的Kiro access token。",
					"type":    "authentication_error",
					"code":    "missing_authorization",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(errorMsg)
			return
		}

		// 提取Bearer token
		accessToken := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			accessToken = authHeader[7:]
		} else {
			accessToken = authHeader
		}

		if accessToken == "" {
			errorMsg := map[string]interface{}{
				"error": map[string]interface{}{
					"message": "无效的Authorization头格式。请确保在OpenAI客户端的api_key参数中设置了有效的Kiro access token。",
					"type":    "authentication_error",
					"code":    "invalid_authorization_format",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(errorMsg)
			return
		}

		// 读取OpenAI API请求
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("读取请求体失败: %v", err), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var openaiReq OpenAIRequest
		if err := json.Unmarshal(body, &openaiReq); err != nil {
			http.Error(w, fmt.Sprintf("解析OpenAI请求失败: %v", err), http.StatusBadRequest)
			return
		}

		// 转换为Kiro API请求
		kiroReq := convertOpenAIToKiro(openaiReq)
		
		// 序列化Kiro请求
		kiroReqBody, err := json.Marshal(kiroReq)
		if err != nil {
			http.Error(w, fmt.Sprintf("序列化Kiro请求失败: %v", err), http.StatusInternalServerError)
			return
		}

		// 创建到Kiro API的请求
		proxyReq, err := http.NewRequest(
			http.MethodPost,
			"https://codewhisperer.us-east-1.amazonaws.com/generateAssistantResponse",
			bytes.NewBuffer(kiroReqBody),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("创建Kiro请求失败: %v", err), http.StatusInternalServerError)
			return
		}

		// 设置请求头
		proxyReq.Header.Set("Content-Type", "application/json")
		proxyReq.Header.Set("Authorization", "Bearer "+accessToken)
		proxyReq.Header.Set("User-Agent", "Kiro2API/1.0")
		proxyReq.Header.Set("Accept", "application/json")
		proxyReq.Header.Set("X-Amz-Target", "CodeWhispererService.GenerateAssistantResponse")

		// 发送请求到Kiro API
		client := &http.Client{}
		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, fmt.Sprintf("发送Kiro请求失败: %v", err), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// 读取Kiro响应
		kiroRespBody, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("读取Kiro响应失败: %v", err), http.StatusInternalServerError)
			return
		}

		if resp.StatusCode != http.StatusOK {
			// 如果是认证错误，提供更有用的错误信息
			if resp.StatusCode == 403 || resp.StatusCode == 401 {
				errorMsg := map[string]interface{}{
					"error": map[string]interface{}{
						"message": "认证失败。请检查api_key中的Kiro access token是否正确且未过期。",
						"type":    "authentication_error",
						"code":    "invalid_token",
						"details": "请从Kiro IDE中获取最新的access token，并在OpenAI客户端的api_key参数中使用它。",
					},
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(errorMsg)
				return
			}
			
			http.Error(w, fmt.Sprintf("Kiro API错误，状态码: %d", resp.StatusCode), resp.StatusCode)
			return
		}

		var kiroResp KiroResponse
		if err := json.Unmarshal(kiroRespBody, &kiroResp); err != nil {
			http.Error(w, fmt.Sprintf("解析Kiro响应失败: %v", err), http.StatusInternalServerError)
			return
		}

		// 转换为OpenAI API响应
		openaiResp := convertKiroToOpenAI(kiroResp, openaiReq.Model)

		// 返回OpenAI API格式的响应
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(openaiResp)
	}))

	// 健康检查端点
	http.HandleFunc("/health", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))

	// 模型列表端点（OpenAI API兼容）
	http.HandleFunc("/v1/models", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		models := map[string]interface{}{
			"object": "list",
			"data": []map[string]interface{}{
				{
					"id":      "claude-sonnet-4-20250514",
					"object":  "model",
					"created": time.Now().Unix(),
					"owned_by": "kiro",
				},
				{
					"id":      "claude-3-7-sonnet-20250219",
					"object":  "model", 
					"created": time.Now().Unix(),
					"owned_by": "kiro",
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models)
	}))

	// 启动服务器
	fmt.Printf("Kiro2API 代理服务器启动，端口: %s\n", port)
	fmt.Printf("OpenAI API 端点: http://localhost:%s/v1/chat/completions\n", port)
	fmt.Printf("使用方法: 在 OpenAI 客户端的 api_key 中设置你的 Kiro access token\n")

	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		fmt.Printf("启动服务器失败: %v\n", err)
		os.Exit(1)
	}
}