package ai

import (
	"context"
)

type TaskType string 

const (
	TaskTypeSummary TaskType = "summary"
	TaskTypePrompt TaskType = "prompt"
)

type TaskComplexity string

const (
	TaskSimple  TaskComplexity = "simple" 
	TaskMedium  TaskComplexity = "medium" 
	TaskComplex TaskComplexity = "complex" 
)


// Provider is an interface for AI providers: OpenAI, DeepSeek, llama can all implement this interface
type Provider interface {

	GetName() string

	// GenerateSummary generates a summary of the conversation based on the previous summary for context
	GenerateSummary(ctx context.Context, messages []MessageInput, previousSummary string) (string, error) 

	// ProcessPrompt processes a prompt and returns a response from the AI provider
	ProcessPrompt(ctx context.Context, messages []MessageInput) (string, error)
}

type MessageInput struct {
	Role string
	Content string
	Name string 
}
