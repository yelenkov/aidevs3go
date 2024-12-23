package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/genai"
)

type GeminiConfig struct {
	GeminiAPIKey string
	Model        string
	System       string `optional:"true"`
	Prompt       string
}

// Read transcription files and asks Gemini a question
func AskGemini(config *GeminiConfig) (string, error) {
	log.Info().Msg("Asking Gemini using genai SDK")

	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  config.GeminiAPIKey,
		Backend: genai.BackendGoogleAI,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Call the GenerateContent method.
	result, err := client.Models.GenerateContent(ctx,
		config.Model,
		genai.Text(config.Prompt),
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: config.System}}},
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}
	log.Debug().Interface("gemini_result", result).Msg("Gemini API Response")

	// Marshal the result to JSON and pretty-print it to a byte array.
	response, err := json.MarshalIndent(*result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	// Log the output.
	fmt.Println(string(response))

	return result.Candidates[0].Content.Parts[0].Text, nil
}
