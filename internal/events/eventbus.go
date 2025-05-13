package events

import (
	"sync"
)

// DefaultEventBus 提供默認的事件總線實現
type DefaultEventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

// NewEventBus 創建新的事件總線實例
func NewEventBus() EventBus {
	return &DefaultEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Publish 發布事件到所有訂閱者
func (b *DefaultEventBus) Publish(event Event) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	handlers, exists := b.handlers[event.Type()]
	if !exists {
		return nil
	}

	var lastErr error
	for _, handler := range handlers {
		if err := handler.Handle(event); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// Subscribe 訂閱特定類型的事件
func (b *DefaultEventBus) Subscribe(eventType string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

// Unsubscribe 取消訂閱特定類型的事件
func (b *DefaultEventBus) Unsubscribe(eventType string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	handlers, exists := b.handlers[eventType]
	if !exists {
		return
	}

	var newHandlers []EventHandler
	for _, h := range handlers {
		if h != handler {
			newHandlers = append(newHandlers, h)
		}
	}

	if len(newHandlers) == 0 {
		delete(b.handlers, eventType)
	} else {
		b.handlers[eventType] = newHandlers
	}
} 