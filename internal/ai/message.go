package ai

import (
	"github.com/ianwu0915/SettleChat/internal/storage"
)

// 命令類型常量
const (
    CommandTypeSummary = "/summary"    // 生成聊天內容的幽默摘要
    CommandTypeHelp    = "/help"       // 顯示幫助信息
    CommandTypeStats   = "/stats"      // 顯示 AI 助手統計信息
    CommandTypeClear   = "/clear"      // 清除摘要歷史
    CommandTypePrompt  = "/prompt"     // 自定義 AI 處理
)

// 處理狀態常量
const (
    StatusPending    = "pending"
    StatusProcessing = "processing"
    StatusCompleted  = "completed"
    StatusFailed     = "failed"
)

// AIMessage 表示一個 AI 命令消息
type AIMessage struct {
    storage.ChatMessage //如果是prompt的話 可能會有很多字串
    
    CommandType string    // 命令類型，如 "/summary", "/help" 等
    Status      string    // 處理狀態
    Prompt      string    // 用於 /prompt 命令的自定義提示
}

// NewAIMessage 創建一個新的 AI 消息
func NewAIMessage(baseMsg storage.ChatMessage, commandType string) *AIMessage {
    return &AIMessage{
        ChatMessage: baseMsg,
        CommandType: commandType,
        Status:      StatusPending,
    }
}

// NewPromptMessage 創建一個自定義提示消息
func NewPromptMessage(baseMsg storage.ChatMessage, prompt string) *AIMessage {
    return &AIMessage{
        ChatMessage: baseMsg,
        CommandType: CommandTypePrompt,
        Status:      StatusPending,
        Prompt:      prompt,
    }
}

// IsValidCommandType 檢查命令類型是否有效
func (m *AIMessage) IsValidCommandType() bool {
    switch m.CommandType {
    case CommandTypeSummary, CommandTypeHelp, CommandTypeStats, 
         CommandTypeClear, CommandTypePrompt:
        return true
    default:
        return false
    }
}

// UpdateStatus 更新消息狀態
func (m *AIMessage) UpdateStatus(status string) {
    m.Status = status
}

// IsPromptCommand 檢查是否為自定義提示命令
func (m *AIMessage) IsPromptCommand() bool {
    return m.CommandType == CommandTypePrompt
}

// GetPrompt 獲取自定義提示內容
func (m *AIMessage) GetPrompt() string {
    return m.Prompt
}