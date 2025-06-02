package handlers

import (
	"strings"

	"github.com/ianwu0915/SettleChat/internal/ai"
	"github.com/ianwu0915/SettleChat/internal/chat"
	messaging "github.com/ianwu0915/SettleChat/internal/nats_messaging"

	"github.com/ianwu0915/SettleChat/internal/storage"
	"github.com/ianwu0915/SettleChat/internal/types"
)

// HandlerManager 負責管理所有NATS事件處理器
type HandlerManager struct {
	store      *storage.PostgresStore
	publisher  types.NATSPublisher
	topics     types.TopicFormatter
	env        string
	hub        *chat.Hub
	aiManager  *ai.Manager
	handlers   map[string]types.MessageHandler
}

// NewHandlerManager 創建一個新的 HandlerManager 實例
func NewHandlerManager(
	store *storage.PostgresStore,
	publisher types.NATSPublisher,
	topics types.TopicFormatter,
	env string,
	hub *chat.Hub,
	aiManager *ai.Manager,
) *HandlerManager {
	return &HandlerManager{
		store:     store,
		publisher: publisher,
		topics:    topics,
		env:       env,
		hub:       hub,
		aiManager: aiManager,
		handlers:  make(map[string]types.MessageHandler),
	}
}

// Initialize 初始化所有處理器
func (m *HandlerManager) Initialize() {
	m.handlers["user.joined"] = NewUserJoinedHandler(m.store, m.publisher, m.topics, m.env)
	m.handlers["user.left"] = NewUserLeftHandler(m.store, m.publisher, m.topics, m.env)
	m.handlers["user.presence"] = NewPresenceHandler(m.store, m.topics, m.env)
	m.handlers["message.chat"] = NewChatMessageHandler(m.store, m.publisher, m.topics)
	m.handlers["message.history.request"] = NewHistoryHandler(m.store, m.publisher, m.topics, m.env)
	m.handlers["message.history.response"] = NewHistoryResponseHandler(m.hub)
	m.handlers["message.broadcast"] = NewBroadcastHandler(m.hub)
	m.handlers["system.message"] = NewSystemMessageHandler(m.publisher, m.topics, m.env)
	m.handlers["connection.event"] = NewConnectionEventHandler(m.store, m.publisher, m.topics)
	m.handlers["ai.command"] = NewAICommandHandler(m.publisher, m.topics, m.env, m.aiManager)
}

// Register 註冊所有處理器到NATS訂閱器
func (m *HandlerManager) Register(subscriber *messaging.Subscriber) {
	for topic, handler := range m.handlers {
		parts := strings.Split(topic, ".")
		if len(parts) >= 2 {
			subscriber.RegisterHandler(parts[0], strings.Join(parts[1:], "."), handler)
		}
	}
}

// GetHandler 獲取指定主題的處理器
func (m *HandlerManager) GetHandler(topic string) (types.MessageHandler, bool) {
	handler, exists := m.handlers[topic]
	return handler, exists
} 