package ai

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	messaging "github.com/ianwu0915/SettleChat/internal/nats_messaging"
	"github.com/ianwu0915/SettleChat/internal/storage"
)

// Facade for AI services
type Manager struct {
	agents map[string]*Agent // Each room has one agent
	store *storage.PostgresStore
	provider Provider
	eventBus *messaging.EventBus
	mu sync.RWMutex

	config ManagerConfig 
}	

type ManagerConfig struct {
	MaxAgentsPerRoom int

	// How long to keep an agent idle before cleaning it up
	AgentIdleTimeout time.Duration

	// How often to clean up idle agents
	CleanupInterval time.Duration

	// Whether to enable periodic summary
	EnablePeriodicSummary bool 

	// How often to generate a summary
	PeriodicSummaryInterval time.Duration

	// Whether to enable concurrent processing for each room
	EnableConcurrentProcessing bool

	// Maximum number of concurrent processes
	MaxConcurrentProcesses int
}

func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		MaxAgentsPerRoom: 1,
		AgentIdleTimeout: 30 * time.Minute,
		CleanupInterval: 10 * time.Minute,
		EnablePeriodicSummary: true,
		PeriodicSummaryInterval: 10 * time.Minute,
		EnableConcurrentProcessing: true,
		MaxConcurrentProcesses: 5,
	}
}

func NewManager(store *storage.PostgresStore, provider Provider, eventBus *messaging.EventBus, options ...ManagerOption) *Manager {

	config := DefaultManagerConfig()

	for _, option := range options {
		option(&config)
	}

	manager :=  &Manager{
		agents: make(map[string]*Agent),
		store: store,
		eventBus: eventBus,
		config: config,
		provider: provider,
	}

	// Start the cleanup goroutine
	go manager.startCleanupRoutine()

	log.Printf("AI manager started with config: %+v", config)
	return manager
}

type ManagerOption func(*ManagerConfig) 

func WithMaxAgentsPerRoom(max int) ManagerOption {
	return func(c *ManagerConfig) {
		c.MaxAgentsPerRoom = max
	}
}

func WithAgentIdleTimeout(timeout time.Duration) ManagerOption {
	return func(c *ManagerConfig) {
		c.AgentIdleTimeout = timeout
	}
}

func WithCleanupInterval(interval time.Duration) ManagerOption {
	return func(c *ManagerConfig) {
		c.CleanupInterval = interval
	}
}

func WithEnablePeriodicSummary(enable bool) ManagerOption {
	return func(c *ManagerConfig) {
		c.EnablePeriodicSummary = enable
	}
}

func WithPeriodicSummaryInterval(interval time.Duration) ManagerOption {
	return func(c *ManagerConfig) {
		c.PeriodicSummaryInterval = interval
	}
}

func WithEnableConcurrentProcessing(enable bool) ManagerOption {
	return func(c *ManagerConfig) {
		c.EnableConcurrentProcessing = enable
	}
}

func WithMaxConcurrentProcesses(max int) ManagerOption {
	return func(c *ManagerConfig) {
		c.MaxConcurrentProcesses = max
	}
}

// HandleAIMessage 處理 AI 命令消息
func (m *Manager) HandleAIMessage(ctx context.Context, msg storage.ChatMessage) (bool, string, error) {
	// 1. 解析命令
	command, args := ParseCommandWithArgs(msg.Content)
	if command == "" {
		return false, "", nil // 不是命令，返回 false
	}

	// 2. 創建 AIMessage
	aiMsg := NewAIMessage(msg, command)
	if args != nil && command == CommandTypePrompt {
		aiMsg = NewPromptMessage(msg, strings.Join(args, " "))
	}

	// 3. 獲取對應的 Agent 並處理
	agent := m.GetOrCreateAgent(msg.RoomID)
	return agent.HandleAIMessage(ctx, aiMsg)
}

func (m *Manager) GetOrCreateAgent(roomID string) *Agent {
	m.mu.RLock()
	agent, exist := m.agents[roomID]
	m.mu.RUnlock()

	if exist {
		agent.updateLastUsed()
		return agent
	}

	// Create a new agent if it doesn't exist
	m.mu.Lock()
	defer m.mu.Unlock()

	if agent, exist := m.agents[roomID]; exist {
		agent.updateLastUsed()
		return agent
	}

	agent = NewAgent(roomID, m.provider, m.store, m)
	m.agents[roomID] = agent

	log.Printf("Created new agent for room %s", roomID)
	if m.config.EnablePeriodicSummary {
		go m.startPeriodicSummaryForRoom(roomID)
	}

	return agent
}

// GetManagerStats 獲取 Manager 統計信息
func (m *Manager) GetManagerStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	stats := make(map[string]interface{})
	stats["total_agents"] = len(m.agents)
	stats["config"] = m.config
	
	// 統計每個 Agent 的信息
	agentStats := make(map[string]interface{})
	for roomID, agent := range m.agents {
		agentStats[roomID] = agent.GetSummaryStats()
	}
	stats["agents"] = agentStats
	
	return stats
}

func (m *Manager) startCleanupRoutine() {
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		m.cleanupIdleAgents()
	}
}

func (m *Manager) cleanupIdleAgents() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check agents that have been idle for too long
	now := time.Now()
	var toRemove []string 

	for roomId, agent := range m.agents {
		if now.Sub(agent.LastUsedTime) > m.config.AgentIdleTimeout {
			toRemove = append(toRemove, roomId)
		}
	}

	// remove agents from the list
	for _, roomID := range toRemove {
		agent := m.agents[roomID]
		// Clear all the resources associated with agent
		agent.ClearSummaryCache()
		delete(m.agents, roomID)
		log.Printf("Clean Idle AI Agents for room: %s", roomID)
	}

	if len(toRemove) > 0 {
		log.Printf("Clean %d Idle AI Agent", len(toRemove))
	}
}

func (m *Manager) startPeriodicSummaryForRoom(roomID string) {
	ticker := time.NewTicker(m.config.PeriodicSummaryInterval)

	go func () {
		defer ticker.Stop()
		// Check if the client still exists
		m.mu.RLock()
		_, exist := m.agents[roomID]
		m.mu.RUnlock()
		if !exist {
			log.Printf("Stop Periodic Summary As Agent do not exist")
			return 
		}

		// Execute Periodic Summary
		

	}()
}

// Shutdown graceully shutdown the manganger 
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, agent := range m.agents {
		agent.ClearSummaryCache()
		log.Printf("Clean Agent: %s", agent.RoomID)
	}

	m.agents = make(map[string]*Agent)

	log.Println("AI Manganer Shut Down")
	return nil
}


