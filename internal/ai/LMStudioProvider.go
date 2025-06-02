package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// LMProvider 實現 LM API
type LMProvider struct {
	Model  string
	Client *http.Client
}

const(
	serverUrl = "http://localhost:1234/v1/chat/completions"
	default_model = "deepseek/deepseek-r1-0528-qwen3-8b"
) 

// NewLMProvider 創建新的 LM 提供者
func NewLMProvider() *LMProvider {
	// if model == "" {
	// 	model_name = model // 默認模型
	// }
	
	return &LMProvider{
		Model:  default_model,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName 返回提供者名稱
func (p *LMProvider) GetName() string {
	return fmt.Sprintf("LM(%s)", p.Model)
}


// /summary 命令 
// GenerateSummary 生成摘要
func (p *LMProvider) GenerateSummary(ctx context.Context, messages []MessageInput, previousSummary string) (string, error) {
	// 構建系統提示
	systemPrompt := "你是一個聊天室助手，負責總結對話內容。請用幽默風趣的方式總結聊天內容，突出重點和有趣的部分。回應請使用繁體中文。"
	
	if previousSummary != "" {
		systemPrompt += fmt.Sprintf("\n\n以下是之前的摘要，請基於此繼續總結新的內容：\n%s", previousSummary)
	}
	
	return p.callLMStudio(ctx, systemPrompt, messages, 500, 0.7)
}

// /prompt 命令
// ProcessPrompt 處理自定義 prompt
func (p *LMProvider) ProcessPrompt(ctx context.Context, prompt string, messages []MessageInput) (string, error) {
	systemPrompt := "你是一個聊天室助手。請根據以下指示處理聊天室使用者向你發出的指示：\n\n" + prompt + "\n\n請用繁體中文回應。"
	maxTokens := 400
	temperature := 0.7
	if len(prompt) > 200 {
		maxTokens = 600
		temperature = 0.6
	}
	return p.callLMStudio(ctx, systemPrompt, messages, maxTokens, temperature)
}

// makeAPICall 執行 API 調用
func (p *LMProvider) callLMStudio(ctx context.Context, systemPrompt string, messages []MessageInput, maxTokens int, temperature float64) (string, error) {
	var chatMessages []map[string]string 
	chatMessages = append(chatMessages, map[string]string{
		"role": "system",
		"content": systemPrompt,
	})

	// summary中的之前聊天訊息
	for _, msg := range messages {
		content := fmt.Sprintf("%s: %s", msg.Name, msg.Content)
		chatMessages = append(chatMessages, map[string]string{
			"role": msg.Role, 
			"content": content,
		})
	}

	reqBody := map[string]interface{} {
		"model": p.Model,
		"messages": chatMessages,
		"temperature": temperature,
		"max_tokens": maxTokens,
		"stream": false,
	}

	print(reqBody)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("Error serialize reqBody: %s", err)
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", serverUrl, strings.NewReader(string(jsonData)))
	if err != nil {
		log.Printf("Error when constructing http reques: %s", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := p.Client.Do(req)
	if err != nil {
		log.Printf("Error when requesting from LM Studio: %s", err)
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error when reading from response Body: %s", err)
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LMStudio API 錯誤 [%d]: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Error deserialize response Body: %s", err)
		return "", err 
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("choices 欄位無效")
	}
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("choice 欄位格式錯誤")
	}
	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("message 欄位格式錯誤")
	}
	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("content 欄位格式錯誤")
	}

	if strings.Contains(content, "</think>") {
		parts := strings.SplitN(content, "</think>", 2)
		content = strings.TrimSpace(parts[1])
	}


	return content, nil

}