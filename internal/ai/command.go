package ai

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ianwu0915/SettleChat/internal/storage"
)

// HandleAIMessage 處理 AI 消息
func (a *Agent) HandleAIMessage(ctx context.Context, msg *AIMessage) (bool, string, error) {
    // 從命令映射中查找處理器
    handler, exists := a.commandMap[msg.CommandType]
    if !exists {
        return true, "不支持的命令", nil
    }
    
    // 調用對應的處理器
    response, err := handler(ctx, msg)
    if err != nil {
        return true, "處理命令時發生錯誤", fmt.Errorf("處理命令失敗: %w", err)
    }
    
    return true, response, nil
}

// HandleHelp 處理 /help 命令
func (a *Agent) HandleHelpCommand(ctx context.Context, message *AIMessage) (string, error) {
	helpText := `🤖 SettleChat AI 助手可用命令：
		/summary - 生成聊天內容的幽默摘要
		/help - 顯示此幫助信息
		/stats - 顯示 AI 助手統計信息
		/clear - 清除摘要歷史（重新開始摘要）
		/prompt <提示> - 自定義 AI 處理（例如：/prompt 分析聊天中提到的技術問題並列出解決方案）

		使用示例：直接輸入命令即可，例如輸入 /summary
		`
	return helpText, nil
}

// HandleSummaryCommand
func (a *Agent) HandleSummaryCommand(ctx context.Context, message *AIMessage) (string, error) {
	summary, err := a.HandleSummary(ctx, message) // 之後要加有options的選項
	if err != nil {
		log.Printf("Error Processing Summary Command: %v", err)
		return "Can't Generate Summary by AI Agent, Try again later", err
	}

	return summary, nil
}


// HandleStatsCommand handles /stats command
// It returns a stats message for the user
func (a *Agent) HandleStatsCommand(ctx context.Context, message *AIMessage) (string, error) {
	stats := a.GetSummaryStats()
	statsText := fmt.Sprintf(`📊 AI 助手統計信息：
		聊天室 ID: %s
		上次摘要時間: %v
		是否有歷史摘要: %v
		已摘要消息數量: %d
		摘要文本長度: %d 字元
		最大提示長度: %d 字元`,
		a.RoomID,
		stats["last_summary_time"],
		stats["has_previous_summary"],
		stats["summarized_message_count"],
		stats["summary_text_length"],
		a.maxPromptLength,
	)
	
	return statsText, nil

}


// HandleClearCommand handles /clear command
// It clears the summary cache
func (a *Agent) HandleClearCommand(ctx context.Context, message *AIMessage) (string, error) {
	a.ClearSummaryCache()
	return "✅ 已清除所有摘要歷史，下次 /summary 將重新開始完整摘要。", nil
}

// HandlePromptCommand handles /prompt command
// It returns a prompt message for the user
func (a *Agent) HandlePromptCommand(ctx context.Context, msg *AIMessage) (string, error) {
	if msg.Prompt == "" {
		return "請提供提示內容", nil
	}
	
	// 將消息轉換為 AI Provider 需要的格式
	messageInputs := a.preprocessMessagesForAI([]storage.ChatMessage{msg.ChatMessage})
	
	// 使用 Provider 處理提示
	return a.Provider.ProcessPrompt(ctx, messageInputs)
}


// ParseCommandWithArgs parses a command string and its arguments.
// It expects the command to start with a forward slash ('/').
// Returns the command and any arguments as separate strings.
// If the input is not a valid command (doesn't start with '/'), returns empty strings.
//
// Example:
//   "/summary" -> command: "/summary", args: nil
//   "/prompt hello world" -> command: "/prompt", args: ["hello", "world"]
//   "not a command" -> command: "", args: nil
func ParseCommandWithArgs(content string) (command string, args []string) {
    if len(content) == 0 || content[0] != '/' {
        return "", nil
    }
    
    parts := strings.Fields(content)
    if len(parts) == 0 {
        return "", nil
    }
    
    command = parts[0]
    if len(parts) > 1 {
        args = parts[1:]
    }
    
    return command, args
}

