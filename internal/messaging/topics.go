package messaging

import "fmt"

// TopicFormatter 實現了 types.TopicFormatter 接口
type TopicFormatter struct {
	// basePrefix 是所有主題的基礎前綴
	basePrefix string
}

// NewTopicFormatter 創建一個新的 TopicFormatter 實例
func NewTopicFormatter(env string) *TopicFormatter {
	return &TopicFormatter{
		basePrefix: fmt.Sprintf("settlechat%s", env),
	}
}

// formatTopic 是一個輔助方法，用於格式化主題字符串
func (t *TopicFormatter) formatTopic(category, action, roomID string) string {
	return fmt.Sprintf("%s.%s.%s.%s", t.basePrefix, category, action, roomID)
}

// GetMessageTopic 返回聊天消息的主題
func (t *TopicFormatter) GetMessageTopic(roomID string) string {
	return t.formatTopic("message", "chat", roomID)
}

// GetPresenceTopic 返回在線狀態的主題
func (t *TopicFormatter) GetPresenceTopic(roomID string) string {
	return t.formatTopic("user", "presence", roomID)
}

// GetHistoryTopic 返回歷史消息請求的主題
func (t *TopicFormatter) GetHistoryTopic(roomID string) string {
	return t.formatTopic("message", "history", roomID)
}

// GetSystemMessageTopic 返回系統消息的主題
func (t *TopicFormatter) GetSystemMessageTopic(roomID string) string {
	return t.formatTopic("system", "message", roomID)
}

// GetUserJoinedTopic 返回用戶加入的主題
func (t *TopicFormatter) GetUserJoinedTopic(roomID string) string {
	return t.formatTopic("user", "joined", roomID)
}

// GetUserLeftTopic 返回用戶離開的主題
func (t *TopicFormatter) GetUserLeftTopic(roomID string) string {
	return t.formatTopic("user", "left", roomID)
}

// GetBroadcastTopic 返回廣播消息的主題
func (t *TopicFormatter) GetBroadcastTopic(roomID string) string {
	return t.formatTopic("message", "broadcast", roomID)
}

// GetHistoryRequestTopic 返回歷史消息請求的主題
func (t *TopicFormatter) GetHistoryRequestTopic(roomID string) string {
	return t.formatTopic("message", "history.request", roomID)
}

// GetHistoryResponseTopic 返回歷史消息響應的主題
func (t *TopicFormatter) GetHistoryResponseTopic(roomID, userID string) string {
	return fmt.Sprintf("%s.%s", t.formatTopic("message", "history.response", roomID), userID)
} 