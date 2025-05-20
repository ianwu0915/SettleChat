package benchmark

import (
	"context"
	"testing"
	"time"

	"github.com/ianwu0915/SettleChat/internal/storage" // 調整為你的包路徑
)

// 初始化測試數據庫連接
func setupTestStore() *storage.PostgresStore {
	// 使用測試數據庫配置
	dsn := "postgres://postgres:secret@localhost:5432/settlechat_test"
	store, err := storage.NewPostgresStore(dsn)
	if err != nil {
		panic(err)
	}
	return store
}

// 測試消息存儲性能
func BenchmarkSaveMessage(b *testing.B) {
	store := setupTestStore()
	defer store.Close()
	
	ctx := context.Background()
	roomID := "benchmark_room"
	senderID := "benchmark_sender"
	
	// 重置計時器
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		msg := storage.ChatMessage{
			RoomID:    roomID,
			SenderID:  senderID,
			Sender:    "Benchmark User",
			Content:   "This is a benchmark test message",
			Timestamp: time.Now(),
		}
		
		err := store.SaveMessage(ctx, msg)
		if err != nil {
			b.Fatalf("保存消息失敗: %v", err)
		}
	}
}

// 測試消息查詢性能
func BenchmarkGetRecentMessages(b *testing.B) {
	store := setupTestStore()
	defer store.Close()
	
	ctx := context.Background()
	roomID := "benchmark_room"
	
	// 預先插入一些消息用於測試
	preloadCount := 100
	for i := 0; i < preloadCount; i++ {
		msg := storage.ChatMessage{
			RoomID:    roomID,
			SenderID:  "preload_sender",
			Sender:    "Preload User",
			Content:   "Preloaded message for benchmark",
			Timestamp: time.Now().Add(-time.Duration(i) * time.Minute),
		}
		if err := store.SaveMessage(ctx, msg); err != nil {
			b.Fatalf("預載消息失敗: %v", err)
		}
	}
	
	// 重置計時器
	b.ResetTimer()
	
	// 測試查詢性能
	for i := 0; i < b.N; i++ {
		_, err := store.GetRecentMessages(ctx, roomID, 50)
		if err != nil {
			b.Fatalf("獲取消息失敗: %v", err)
		}
	}
}