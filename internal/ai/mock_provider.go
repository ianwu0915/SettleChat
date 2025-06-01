package ai

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// MockProvider 是一個用於測試的模擬 AI Provider
type MockProvider struct {
	name string
	// 可以添加更多配置選項
	responseDelay time.Duration
}

// NewMockProvider 創建一個新的 MockProvider
func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		name:          name,
		responseDelay: 100 * time.Millisecond, // 模擬網絡延遲
	}
}

// GetName 返回 provider 的名稱
func (m *MockProvider) GetName() string {
	return m.name
}

// GenerateSummary 生成一個模擬的摘要
func (m *MockProvider) GenerateSummary(ctx context.Context, messages []MessageInput, previousSummary string) (string, error) {
	// 模擬處理延遲
	time.Sleep(m.responseDelay)

	// 如果沒有消息，返回錯誤
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages to summarize")
	}

	// 構建一個簡單的摘要
	var summary strings.Builder
	summary.WriteString("📝 聊天摘要：\n\n")

	// 如果有之前的摘要，添加它
	if previousSummary != "" {
		summary.WriteString("之前的摘要：\n")
		summary.WriteString(previousSummary)
		summary.WriteString("\n\n")
	}

	// 添加新消息的摘要
	summary.WriteString("新消息摘要：\n")
	for _, msg := range messages {
		summary.WriteString(fmt.Sprintf("- %s: %s\n", msg.Name, msg.Content))
	}

	// 添加一個總結
	summary.WriteString("\n總結：這是一個模擬的摘要，用於測試目的。")
	summary.WriteString(fmt.Sprintf(" 共處理了 %d 條消息。", len(messages)))

	return summary.String(), nil
}

// ProcessPrompt 處理一個模擬的提示
func (m *MockProvider) ProcessPrompt(ctx context.Context, messages []MessageInput) (string, error) {
	// 模擬處理延遲
	time.Sleep(m.responseDelay)

	// 如果沒有消息，返回錯誤
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages to process")
	}

	// 構建一個模擬的回應
	var response strings.Builder
	response.WriteString("🤖 AI 回應：\n\n")

	// 處理每條消息
	for _, msg := range messages {
		response.WriteString(fmt.Sprintf("收到來自 %s 的消息：\n", msg.Name))
		response.WriteString(fmt.Sprintf("內容：%s\n\n", msg.Content))
	}

	// 添加一個模擬的 AI 分析
	response.WriteString("分析結果：\n")
	response.WriteString("- 這是一個模擬的 AI 回應\n")
	response.WriteString("- 用於測試目的\n")
	response.WriteString("- 實際的 AI 回應會更智能\n")

	return response.String(), nil
}

// SetResponseDelay 設置回應延遲時間
func (m *MockProvider) SetResponseDelay(delay time.Duration) {
	m.responseDelay = delay
} 