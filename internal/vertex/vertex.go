package vertex

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/vertexai/genai"
	"github.com/rs/zerolog/log"
)

type VertexConfig struct {
	Project  string
	Location string
	Model    string
	System   string
	Prompt   string
}

// read transcription files and asks Vertex a question
func AskVertex(config *VertexConfig) (string, error) {
	log.Debug().Interface("vertex_config", config).Msg("Calling AskVertex with config")
	ctx := context.Background()
	client, err := genai.NewClient(ctx, config.Project, config.Location)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create VertexAI client")
		return "", fmt.Errorf("failed to create VertexAI client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(config.Model)
	model.SetTemperature(0.9)
	model.SetTopP(0.5)
	model.SetTopK(20)
	model.SetMaxOutputTokens(100)
	model.SystemInstruction = genai.NewUserContent(genai.Text(config.System))
	log.Debug().Str("prompt", config.Prompt).Str("system_instruction", config.System).Msg("Sending request to Gemini")
	result, err := model.GenerateContent(ctx, genai.Text(config.Prompt))
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate content")
		return "", fmt.Errorf("failed to generate content: %w", err)
	}
	log.Debug().Interface("gemini_result", result).Msg("Gemini API Response")

	response, err := json.MarshalIndent(*result, "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal result")
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	// Log the output.
	fmt.Println(string(response))

	return string(response), nil
}
