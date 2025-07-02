package ai

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ianwu0915/SettleChat/internal/storage"
)
// Summary 工作流程說明：
// 1. 當用戶發送 /summary 命令時，系統會：
//    - 檢查是否有新的消息需要摘要
//    - 獲取上次摘要作為上下文
//    - 將新消息轉換為 AI Provider 可處理的格式
//    - 使用 AI Provider 生成新的摘要
//    - 更新摘要緩存
//
// 2. 摘要緩存 (SummaryCache) 包含：
//    - LastSummaryTime: 上次生成摘要的時間
//    - LastSummaryText: 上次生成的摘要內容
//    - SummarizedMsgIDs: 已經被摘要過的消息 ID 集合
//
// 3. 消息處理流程：
//    - 從數據庫獲取最近的 100 條消息
//    - 過濾出上次摘要後的新消息
//    - 確保每條消息只被摘要一次
//    - 將消息轉換為 AI Provider 需要的格式
//
// 4. 錯誤處理：
//    - 處理數據庫查詢錯誤
//    - 處理 AI Provider 生成摘要失敗的情況
//    - 提供友好的錯誤提示



// HandleSummary 處理摘要生成
func (a *Agent) HandleSummary(ctx context.Context, msg *AIMessage) (string, error) {
	// 獲取新的消息
	newMessages, err := a.getNewMessagesForSummary(ctx)
	if err != nil {
		return "", fmt.Errorf("獲取新消息失敗: %w", err)
	}

	if len(newMessages) == 0 {
		return "沒有新的消息需要摘要", nil
	}

	// 獲取上次摘要
	previousSummary := a.summaryCache.LastSummaryText

	// 將消息轉換為 AI Provider 需要的格式
	messageInputs := a.preprocessMessagesForAI(newMessages)

	// 生成新的摘要
	summary, err := a.Provider.GenerateSummary(ctx, messageInputs, previousSummary)
	if err != nil {
		return "", fmt.Errorf("生成摘要失敗: %w", err)
	}

	// 更新摘要緩存
	a.updateSummaryCache(newMessages, summary)

	return summary, nil
}

// getNewMessagesForSummary 獲取需要摘要的新消息
func (a *Agent) getNewMessagesForSummary(ctx context.Context) ([]storage.ChatMessage, error) {
	// 獲取最近的100條消息
	allMessages, err := a.store.GetRecentMessages(ctx, a.RoomID, 100)
	if err != nil {
		return nil, fmt.Errorf("從數據庫獲取消息失敗: %w", err)
	}

	// 如果是第一次摘要，返回所有消息
	if a.summaryCache.LastSummaryTime.IsZero() {
		return allMessages, nil
	}

	// 過濾出新的消息
	var newMessages []storage.ChatMessage
	for _, msg := range allMessages {
		if msg.Timestamp.After(a.summaryCache.LastSummaryTime) && !a.summaryCache.SummarizedMsgIDs[msg.ID] {
			newMessages = append(newMessages, msg)
		}
	}

	return newMessages, nil
}

// updateSummaryCache 更新摘要緩存
func (a *Agent) updateSummaryCache(messages []storage.ChatMessage, newSummary string) {
	a.summaryCache.mu.Lock()
	defer a.summaryCache.mu.Unlock()

	a.summaryCache.LastSummaryTime = time.Now()
	a.summaryCache.LastSummaryText = newSummary

	// 更新已摘要的消息ID
	for _, msg := range messages {
		a.summaryCache.SummarizedMsgIDs[msg.ID] = true
	}

	// 如果緩存太大，清理舊的ID
	if len(a.summaryCache.SummarizedMsgIDs) > 1000 {
		a.summaryCache.SummarizedMsgIDs = make(map[int]bool)
		for _, msg := range messages {
			a.summaryCache.SummarizedMsgIDs[msg.ID] = true
		}
		
	}
}

// preprocessMessagesForAI 將消息轉換為 AI Provider 需要的格式
func (a *Agent) preprocessMessagesForAI(messages []storage.ChatMessage) []MessageInput {
	var inputs []MessageInput
	for _, msg := range messages {
		role := "user"
		if msg.Sender == "system" {
			role = "system"
		}

		input := MessageInput{
			Role:    role,
			Content: msg.Content,
			Name:    msg.Sender,
		}
		inputs = append(inputs, input)
	}
	return inputs
}

// GetSummaryStats 獲取摘要統計信息
func (a *Agent) GetSummaryStats() map[string]interface{} {
	a.summaryCache.mu.Lock()
	defer a.summaryCache.mu.Unlock()

	return map[string]interface{}{
		"last_summary_time":        a.summaryCache.LastSummaryTime,
		"has_previous_summary":     a.summaryCache.LastSummaryText != "",
		"summarized_message_count": len(a.summaryCache.SummarizedMsgIDs),
		"summary_text_length":      len(a.summaryCache.LastSummaryText),
	}
}

// ClearSummaryCache 清除摘要緩存
func (a *Agent) ClearSummaryCache() {
	a.summaryCache.mu.Lock()
	defer a.summaryCache.mu.Unlock()

	a.summaryCache.LastSummaryText = ""
	a.summaryCache.SummarizedMsgIDs = make(map[int]bool)
	a.summaryCache.LastSummaryTime = time.Time{}

	log.Println("摘要緩存已清除")
}
