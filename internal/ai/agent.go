package ai

import (
	"context"
	"sync"
	"time"

	"github.com/ianwu0915/SettleChat/internal/storage"
)

// All command handlers will follow this signature
// ex: /summary, /help, /prompt
// will be func(context.Context, []storage.ChatMessage) (string, error)
type CommandHandler func(context.Context, *AIMessage) (string, error)

// 
type Agent struct {
	RoomID string
	Provider Provider
	store *storage.PostgresStore
	commandMap map[string]CommandHandler
	summaryCache *SummaryCache
	maxPromptLength int
	manager *Manager
	LastUsedTime time.Time
	mu sync.RWMutex
}

type SummaryCache struct {
	LastSummaryTime time.Time
	LastSummaryText string
	SummarizedMsgIDs map[int]bool
	mu sync.Mutex
}

func NewAgent(roomID string, provider Provider, store *storage.PostgresStore, manager *Manager) *Agent {
	agent := &Agent{
		RoomID: roomID,
		Provider: provider,
		store: store,
		commandMap: make(map[string]CommandHandler),
		summaryCache: &SummaryCache{
			SummarizedMsgIDs: make(map[int]bool),
		},
		maxPromptLength: 4000,
		manager: manager,
		LastUsedTime: time.Now(),
	}

	agent.registerCommands()
	return agent
}

func (a *Agent) registerCommands() {
	a.commandMap[CommandTypeHelp] = a.HandleHelpCommand
	a.commandMap[CommandTypeSummary] = a.HandleSummaryCommand
	a.commandMap[CommandTypeStats] = a.HandleStatsCommand
	a.commandMap[CommandTypeClear] = a.HandleClearCommand
	a.commandMap[CommandTypePrompt] = a.HandlePromptCommand
}



func (a *Agent) updateLastUsed() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.LastUsedTime = time.Now()
}

func (a *Agent) Shutdown(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	return nil
}

