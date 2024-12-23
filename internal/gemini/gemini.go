package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/genai"
)

// askGemini reads transcription files and asks Gemini a question using the genai SDK.
func AskGemini(ctx context.Context, geminiAPIKey string, system *string, prompt string) (string, error) {
	log.Info().Msg("Asking Gemini using genai SDK")

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  geminiAPIKey,
		Backend: genai.BackendGoogleAI,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Call the GenerateContent method.
	result, err := client.Models.GenerateContent(ctx,
		"gemini-3.0-flash-exp",
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: *system}}},
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	// Marshal the result to JSON and pretty-print it to a byte array.
	response, err := json.MarshalIndent(*result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	// Log the output.
	fmt.Println(string(response))

	return string(response), nil
}
