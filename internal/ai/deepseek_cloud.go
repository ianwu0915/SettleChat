package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DeepSeekProvider 實現 DeepSeek API
type DeepSeekProvider struct {
	APIKey string
	Model  string
	Client *http.Client
}

// NewDeepSeekProvider 創建新的 DeepSeek 提供者
func NewDeepSeekProvider(apiKey string, model string) *DeepSeekProvider {
	if model == "" {
		model = "deepseek-chat" // 默認模型
	}
	
	return &DeepSeekProvider{
		APIKey: apiKey,
		Model:  model,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName 返回提供者名稱
func (p *DeepSeekProvider) GetName() string {
	return fmt.Sprintf("DeepSeek(%s)", p.Model)
}

// CanHandle 檢查是否能處理特定任務
func (p *DeepSeekProvider) CanHandle(taskType TaskType, complexity TaskComplexity) bool {
	// DeepSeek 適合處理簡單到中等複雜度的任務
	switch taskType {
	case TaskTypeSummary, TaskTypePrompt:
		return complexity <= TaskMedium // 只處理簡單和中等任務
	default:
		return false
	}
}

// EstimateCost 估算任務成本
func (p *DeepSeekProvider) EstimateCost(taskType TaskType, complexity TaskComplexity, inputLength int) float64 {
	// DeepSeek 的價格通常比 OpenAI 便宜
	baseCostPer1K := 0.0007 // 假設 $0.0007 per 1K tokens
	
	// 估算 token 數
	estimatedTokens := float64(inputLength) / 4
	
	// 根據任務複雜度調整
	complexityMultiplier := 1.0
	switch complexity {
	case TaskSimple:
		complexityMultiplier = 0.3
	case TaskMedium:
		complexityMultiplier = 0.7
	case TaskComplex:
		complexityMultiplier = 1.0
	}
	
	// 考慮輸出 tokens
	outputTokensRatio := 0.3
	totalTokens := estimatedTokens * (1 + outputTokensRatio)
	
	return (totalTokens / 1000) * baseCostPer1K * complexityMultiplier
}

// GenerateSummary 生成摘要
func (p *DeepSeekProvider) GenerateSummary(ctx context.Context, messages []MessageInput, previousSummary string) (string, error) {
	// 構建系統提示
	systemPrompt := "你是一個AI聊天室助手，負責總結對話內容。請用諷刺的方式去總結聊天內容，突出重點和有趣的部份，並嘲笑發訊息的人類。回應請使用英文。"
	
	if previousSummary != "" {
		systemPrompt += fmt.Sprintf("\n\n以下是之前的摘要，請基於此繼續總結新的內容：\n%s", previousSummary)
	}
	
	// 構建請求消息
	var deepSeekMessages []map[string]string
	
	// 添加系統消息
	deepSeekMessages = append(deepSeekMessages, map[string]string{
		"role":    "system",
		"content": systemPrompt,
	})
	
	// 添加聊天消息
	for _, msg := range messages {
		content := fmt.Sprintf("%s: %s", msg.Name, msg.Content)
		deepSeekMessages = append(deepSeekMessages, map[string]string{
			"role":    msg.Role,
			"content": content,
		})
	}
	
	// 構建請求體
	requestBody := map[string]interface{}{
		"model":       p.Model,
		"messages":    deepSeekMessages,
		"temperature": 0.7,
		"max_tokens":  500,
		"stream":      false,
	}
	
	return p.makeAPICall(ctx, "https://api.deepseek.com/v1/chat/completions", requestBody)
}

// ProcessPrompt 處理自定義 prompt
func (p *DeepSeekProvider) ProcessPrompt(ctx context.Context, prompt string, messages []MessageInput) (string, error) {
	// 構建系統提示
	systemPrompt := fmt.Sprintf("你是一個聊天室助手。請根據以下指示處理聊天內容：\n\n%s\n\n請用英文回應。", prompt)
	
	// 構建請求消息
	var deepSeekMessages []map[string]string
	
	// 添加系統消息
	deepSeekMessages = append(deepSeekMessages, map[string]string{
		"role":    "system",
		"content": systemPrompt,
	})
	
	// 添加聊天消息
	for _, msg := range messages {
		content := fmt.Sprintf("%s: %s", msg.Name, msg.Content)
		deepSeekMessages = append(deepSeekMessages, map[string]string{
			"role":    msg.Role,
			"content": content,
		})
	}
	
	// 根據 prompt 長度調整參數
	maxTokens := 400
	temperature := 0.7
	
	if len(prompt) > 200 {
		maxTokens = 600
		temperature = 0.6
	}
	
	// 構建請求體
	requestBody := map[string]interface{}{
		"model":       p.Model,
		"messages":    deepSeekMessages,
		"temperature": temperature,
		"max_tokens":  maxTokens,
		"stream":      false,
	}
	
	return p.makeAPICall(ctx, "https://api.deepseek.com/v1/chat/completions", requestBody)
}

// makeAPICall 執行 API 調用
func (p *DeepSeekProvider) makeAPICall(ctx context.Context, url string, requestBody map[string]interface{}) (string, error) {
	// 序列化請求
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("序列化請求失敗: %w", err)
	}
	
	// 創建請求
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(requestJSON)))
	if err != nil {
		return "", fmt.Errorf("創建請求失敗: %w", err)
	}
	
	// 設置請求頭
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	req.Header.Set("Accept", "application/json")
	
	// 發送請求
	resp, err := p.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("發送請求失敗: %w", err)
	}
	defer resp.Body.Close()
	
	// 讀取響應
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("讀取響應失敗: %w", err)
	}
	
	// 檢查 HTTP 狀態碼
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DeepSeek API 錯誤 [%d]: %s", resp.StatusCode, string(body))
	}
	
	// 解析響應
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析響應失敗: %w", err)
	}
	
	// 檢查是否有錯誤
	if errorInfo, exists := result["error"]; exists {
		return "", fmt.Errorf("DeepSeek API 返回錯誤: %v", errorInfo)
	}
	
	// 獲取生成的文本
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("無效的 DeepSeek API 響應格式: choices 欄位錯誤")
	}
	
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("無效的 DeepSeek API 響應格式: choice 欄位錯誤")
	}
	
	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("無效的 DeepSeek API 響應格式: message 欄位錯誤")
	}
	
	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("無效的 DeepSeek API 響應格式: content 欄位錯誤")
	}
	
	return content, nil
}