package messaging

import (
	"fmt"
	"log"
	"sync"
	"time"
	"github.com/nats-io/nats.go"
)

type NATSManager struct {
	conn *nats.Conn
	reconnect bool 
	url string
	mutex sync.RWMutex
	options []nats.Option
}

func NewNATSManger(url string, reconnect bool) *NATSManager {
	if url == "" {
		url = nats.DefaultURL //nats://localhost:4222
	}

	return &NATSManager{
		url: url,
		reconnect: reconnect,
	}
}

func (m *NATSManager) Connect() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.conn != nil && m.conn.IsConnected() {
		return nil
	}

	// Default Setup Options
	opts := []nats.Option{
		nats.Name("SettleChat"),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("NATS disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NASTS reconnected to %s", nc.ConnectedUrl() )
		}),
		nats.ErrorHandler(func(nc *nats.Conn, s *nats.Subscription, err error) {
			log.Printf("NATS error: %v", err)
		}),
		nats.MaxReconnects(-1), //Retry forever (every 2 secs)

		// nats.ReconnectWait(time.Second * 5) // 每 5 秒嘗試一次 reconnect
		// nats.Timeout(time.Second * 2)       // 初始連線 timeout
		// nats.PingInterval(time.Minute * 2)  // 心跳 ping 頻率
	}

	// Handle Customized options 
	if len(m.options) > 0 {
		opts = append(opts, m.options...)
	}

	// Build the Connection
	conn, err := nats.Connect(m.url, opts...)

	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	m.conn = conn
	log.Printf("Connected to NATS server at %s", m.url)
	return nil
}

// Disconnect the NATS server
func (m *NATSManager) Disconnect() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.conn != nil {
		m.conn.Drain() // Blocking call 
		m.conn = nil 
		log.Println("Disconnected from NATS server")

	}
}

// Get the Server conn 
// If it's disconnected, will try to reconncet with nats server
func (m *NATSManager) GetConn() (*nats.Conn, error) {
	m.mutex.RLock()
	if m.conn != nil && m.conn.IsConnected() {
		defer m.mutex.RUnlock()
		return m.conn, nil
	}
	m.mutex.RUnlock()

	if err := m.Connect(); err != nil {
		return nil, err
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.conn, nil 
}

// Publish data to the Subject
func (m *NATSManager) Publish(subject string, data []byte) error {
	conn, err := m.GetConn();
	if err != nil {
		log.Printf("Couldn't Publish since its Disconnected with the server: %s", err)
	}
	return conn.Publish(subject, data)
}

// Publish data to the Subject
// Non-Blocking: 會自己開一個Goroutine在Background
// NATS 客戶端已處理並發：NATS 客戶端庫已經在內部使用 goroutine 處理訂閱
// 多訂閱共享連接：所有訂閱共享同一個 NATS 連接，客戶端庫高效地管理這些訂閱
// 回調函數是並發執行的：不同訂閱的回調函數會在 NATS 客戶端的不同 goroutine 中並發執行
// 只有取消訂閱後，會結束
func (m *NATSManager) Subscribe(subject string, msgHandler nats.MsgHandler) (*nats.Subscription, error) {
	conn, err := m.GetConn();
	if err != nil {
		log.Printf("Couldn't Publish since its Disconnected with the server: %s", err)
	}
	
	return conn.Subscribe(subject, msgHandler)
}

// WithOptions 設置自定義的NATS連接選項
func (m *NATSManager) WithOptions(options ...nats.Option) *NATSManager {
	m.options = append(m.options, options...)
	return m
}

// IsConnected 檢查是否已連接到NATS服務器
func (m *NATSManager) IsConnected() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.conn != nil && m.conn.IsConnected()
}

// WaitForConnection 等待NATS連接建立，直到超時
func (m *NATSManager) WaitForConnection(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if m.IsConnected() {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}