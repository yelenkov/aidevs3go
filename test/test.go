package main

import (
	"log"
	"time"

	"github.com/dawidjelenkowski/aidevs3go/internal/chatopenaiservice"
	"github.com/dawidjelenkowski/aidevs3go/internal/langfuseservice"
	"github.com/dawidjelenkowski/aidevs3go/internal/utils"
)

func main() {
	// Get all API keys at once
	keys, err := utils.GetAPIKeys(
		"openai-api-key",
		"langfuse-public-key",
		"langfuse-secret-key",
	)
	if err != nil {
		log.Fatalf("Failed to get API keys: %v", err)
	}

	// Initialize configuration
	langfuseHost := "http://localhost:3001"

	langfuseConfig := langfuseservice.LangfuseConfig{
		BaseURL:   langfuseHost,
		PublicKey: keys["langfuse-public-key"],
		SecretKey: keys["langfuse-secret-key"],
	}

	// Create chat service
	chatService := chatopenaiservice.NewChatService(keys["openai-api-key"])

	// Prepare messages
	messages := []chatopenaiservice.ChatMessage{
		{Role: "system", Content: "You are a helpful comedian. Tell a short, clean joke."},
		{Role: "user", Content: "Make me laugh!"},
	}

	// Create generation with start time
	generation := langfuseservice.Generation{
		Name:      "Joke Generation",
		Model:     "gpt-4o-mini",
		StartTime: time.Now(),
		Input:     messages,
		ModelParameters: map[string]string{
			"temperature": "0.7",
		},
	}

	// Get response from OpenAI
	response, err := chatService.Completion(messages, generation.Model)
	if err != nil {
		log.Fatalf("Completion error: %v", err)
	}

	// Update generation with results
	generation.Output = response.Choices[0].Message.Content
	generation.EndTime = time.Now()
	generation.Usage = langfuseservice.ExtractUsageFromResponse(response)

	// Send to Langfuse
	generationID, err := langfuseConfig.CreateGeneration(generation)
	if err != nil {
		log.Printf("Failed to create generation: %v", err)
	}
	log.Printf("Created generation with ID: %s", generationID)
}
