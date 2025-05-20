package benchmark

import (
	"fmt"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ianwu0915/SettleChat/internal/storage"
)

// 測試WebSocket連接建立性能
func BenchmarkWebSocketConnections(b *testing.B) {
	serverURL := "ws://localhost:8080/ws"
	
	// 避免在基準測試計時期間進行設置
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// 創建唯一用戶ID和用戶名
		userID := fmt.Sprintf("bench_user_%d", i)
		username := fmt.Sprintf("bench_name_%d", i)
		roomID := "benchmark_room"
		
		// 構建URL
		u, _ := url.Parse(serverURL)
		q := u.Query()
		q.Set("room", roomID)
		q.Set("user_id", userID)
		q.Set("username", username)
		u.RawQuery = q.Encode()
		
		// 連接WebSocket
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			b.Fatalf("連接失敗: %v", err)
		}
		
		// 必須關閉連接，避免資源耗盡
		c.Close()
	}
}

// 測試並發WebSocket連接建立
func BenchmarkConcurrentWebSocketConnections(b *testing.B) {
	serverURL := "ws://localhost:8080/ws"
	roomID := "benchmark_room"
	
	// 設置並行性
	b.SetParallelism(100) // 最多100個並行goroutine
	b.ResetTimer()
	
	var counter int32 = 0
	var mu sync.Mutex
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 安全地獲取和增加counter
			mu.Lock()
			userNum := counter
			counter++
			mu.Unlock()
			
			// 創建唯一用戶信息
			userID := fmt.Sprintf("bench_user_%d", userNum)
			username := fmt.Sprintf("bench_name_%d", userNum)
			
			// 構建URL
			u, _ := url.Parse(serverURL)
			q := u.Query()
			q.Set("room", roomID)
			q.Set("user_id", userID)
			q.Set("username", username)
			u.RawQuery = q.Encode()
			
			// 連接WebSocket
			c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				continue // 在基準測試中跳過錯誤
			}
			
			// 關閉連接
			c.Close()
		}
	})
}

// 測試消息吞吐量
func BenchmarkMessageThroughput(b *testing.B) {
	serverURL := "ws://localhost:8080/ws"
	roomID := "benchmark_room"
	
	// 建立與服務器的連接
	// 使用兩個客戶端 - 一個發送，一個接收
	senderID := "bench_sender"
	receiverID := "bench_receiver"
	
	// 建立發送者連接
	senderURL, _ := url.Parse(serverURL)
	sq := senderURL.Query()
	sq.Set("room", roomID)
	sq.Set("user_id", senderID)
	sq.Set("username", "Sender")
	senderURL.RawQuery = sq.Encode()
	
	sender, _, err := websocket.DefaultDialer.Dial(senderURL.String(), nil)
	if err != nil {
		b.Fatalf("發送者連接失敗: %v", err)
	}
	
	// 建立接收者連接
	receiverURL, _ := url.Parse(serverURL)
	rq := receiverURL.Query()
	rq.Set("room", roomID)
	rq.Set("user_id", receiverID)
	rq.Set("username", "Receiver")
	receiverURL.RawQuery = rq.Encode()
	
	receiver, _, err := websocket.DefaultDialer.Dial(receiverURL.String(), nil)
	if err != nil {
		sender.Close()
		b.Fatalf("接收者連接失敗: %v", err)
	}
	
	// 設置發送者的 ping/pong 處理
	sender.SetPingHandler(func(string) error {
		return sender.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(10*time.Second))
	})
	
	// 設置接收者的 ping/pong 處理
	receiver.SetPingHandler(func(string) error {
		return receiver.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(10*time.Second))
	})
	
	// 等待兩個連接都建立完成
	time.Sleep(2 * time.Second)
	
	// 啟動goroutine接收消息
	received := make(chan bool, b.N)
	errors := make(chan error, 1)
	done := make(chan struct{})
	historyDone := make(chan struct{})
	
	// 啟動一個goroutine來處理歷史消息
	go func() {
		defer close(historyDone)
		for {
			var msg storage.ChatMessage
			err := receiver.ReadJSON(&msg)
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					return
				}
				select {
				case errors <- fmt.Errorf("接收歷史消息錯誤: %v", err):
				default:
				}
				return
			}
			// 如果是歷史消息，繼續接收
			if msg.SenderID == "system" {
				continue
			}
			// 如果是普通消息，通知主goroutine
			select {
			case received <- true:
			case <-done:
				return
			}
		}
	}()
	
	// 預分配訊息池
	messagePool := make([]storage.ChatMessage, 1000)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		message := &messagePool[i%1000]
		message.RoomID = roomID
		message.SenderID = senderID
		message.Sender = "Sender"
		message.Content = fmt.Sprintf("Benchmark message %d", i)
		message.Timestamp = time.Now()
		
		if err := sender.WriteJSON(message); err != nil {
			continue
		}
		
		select {
		case <-received:
		case <-time.After(1 * time.Second):
			continue
		}
	}
}

