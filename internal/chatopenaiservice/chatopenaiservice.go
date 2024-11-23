package chatopenaiservice

import (
	"context"
	"fmt"
	"log"

	"github.com/sashabaranov/go-openai"
)

// OpenAIService handles interactions with the OpenAI API
type OpenAIService struct {
	client *openai.Client
}

// ChatMessage represents a message in the chat completion
type ChatMessage struct {
	Role    string
	Content string
}

// CompletionOptions allows configuring the completion request
type CompletionOptions struct {
	Model       string
	Stream      bool
	JSONMode    bool
	MaxTokens   int
	Temperature float32
}

// NewOpenAIService creates a new instance of OpenAIService
func NewOpenAIService(apiKey string) *OpenAIService {
	client := openai.NewClient(apiKey)
	return &OpenAIService{
		client: client,
	}
}

// Completion performs a chat completion request
func (s *OpenAIService) Completion(messages []ChatMessage, opts CompletionOptions) (*openai.ChatCompletionResponse, error) {
	// Convert our ChatMessage to openai.ChatCompletionMessage
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Prepare request
	request := openai.ChatCompletionRequest{
		Model:       opts.Model,
		Messages:    openaiMessages,
		Stream:      opts.Stream,
		MaxTokens:   opts.MaxTokens,
		Temperature: float32(opts.Temperature),
	}

	// Add JSON mode if specified
	if opts.JSONMode {
		request.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: "json_object",
		}
	}

	// Perform completion
	resp, err := s.client.CreateChatCompletion(context.Background(), request)
	if err != nil {
		log.Printf("OpenAI completion error: %v", err)
		return nil, fmt.Errorf("failed to get completion: %w", err)
	}

	return &resp, nil
}

// ChatService provides a higher-level interface for chat completions
type ChatService struct {
	openaiService *OpenAIService
}

// NewChatService creates a new ChatService
func NewChatService(apiKey string) *ChatService {
	return &ChatService{
		openaiService: NewOpenAIService(apiKey),
	}
}

// Completion performs a standard chat completion
func (s *ChatService) Completion(messages []ChatMessage, model string) (*openai.ChatCompletionResponse, error) {
	opts := CompletionOptions{
		Model:       model,
		Stream:      false,
		JSONMode:    false,
		MaxTokens:   0,   // Use default
		Temperature: 0.7, // Default temperature
	}

	return s.openaiService.Completion(messages, opts)
}
