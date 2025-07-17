package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Anthropic API 结构定义
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []AnthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature float64            `json:"temperature,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
}

type AnthropicResponse struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type,omitempty"`
	Role    string `json:"role,omitempty"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content,omitempty"`
	Model        string `json:"model,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage,omitempty"`
}

func main() {
	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	// 检查环境变量
	baseURL := os.Getenv("ANTHROPIC_BASE_URL")
	apiKey := os.Getenv("ANTHROPIC_API_KEY")

	if baseURL == "" {
		baseURL = "https://codewhisperer.us-east-1.amazonaws.com/generateAssistantResponse"
		fmt.Println("使用默认的 ANTHROPIC_BASE_URL:", baseURL)
	}

	if apiKey == "" {
		fmt.Println("警告: 未设置 ANTHROPIC_API_KEY 环境变量")
		fmt.Println("请设置环境变量或在请求中提供 Authorization 头")
	}

	startAnthropicProxy(port, baseURL, apiKey)
}

func startAnthropicProxy(port, baseURL, defaultAPIKey string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 启用CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "只支持 POST 请求", http.StatusMethodNotAllowed)
			return
		}

		// 获取API Key
		apiKey := defaultAPIKey
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				apiKey = authHeader
			}
		}

		if apiKey == "" {
			http.Error(w, "缺少API Key。请设置 ANTHROPIC_API_KEY 环境变量或提供 Authorization 头", http.StatusUnauthorized)
			return
		}

		// 读取请求体
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("读取请求体失败: %v", err), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		// 解析请求以验证格式
		var anthropicReq AnthropicRequest
		if err := json.Unmarshal(body, &anthropicReq); err != nil {
			http.Error(w, fmt.Sprintf("解析请求失败: %v", err), http.StatusBadRequest)
			return
		}

		// 创建代理请求
		proxyReq, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(body))
		if err != nil {
			http.Error(w, fmt.Sprintf("创建代理请求失败: %v", err), http.StatusInternalServerError)
			return
		}

		// 设置请求头
		proxyReq.Header.Set("Content-Type", "application/json")
		proxyReq.Header.Set("Authorization", "Bearer "+apiKey)
		proxyReq.Header.Set("User-Agent", "Kiro2API-AnthropicProxy/1.0")
		proxyReq.Header.Set("Accept", "application/json")
		proxyReq.Header.Set("X-Amz-Target", "CodeWhispererService.GenerateAssistantResponse")
		proxyReq.Header.Set("X-Amz-Date", time.Now().UTC().Format("20060102T150405Z"))

		// 发送请求
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, fmt.Sprintf("代理请求失败: %v", err), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// 读取响应
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("读取响应失败: %v", err), http.StatusInternalServerError)
			return
		}

		// 设置响应头
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)

		// 返回响应
		w.Write(respBody)
	})

	// 健康检查端点
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"service": "kiro2api-anthropic-proxy",
		})
	})

	fmt.Printf("Kiro2API Anthropic代理服务器启动，端口: %s\n", port)
	fmt.Printf("代理端点: http://localhost:%s\n", port)
	fmt.Printf("目标API: %s\n", os.Getenv("ANTHROPIC_BASE_URL"))
	fmt.Printf("使用方法: 发送标准的Anthropic API请求到 http://localhost:%s\n", port)
	
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}