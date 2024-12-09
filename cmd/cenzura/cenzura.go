package main

import (
	"context"
	"path/filepath"

	"github.com/dawidjelenkowski/aidevs3go/internal/logging"
	"github.com/dawidjelenkowski/aidevs3go/internal/utils"
	"github.com/rs/zerolog/log"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	logging.Setup()
	// Get API keys
	aidevsKey, err := utils.GetAPIKey("aidevs-api-key")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get AIDevs API key")
	}
	openaiKey, err := utils.GetAPIKey("openai-api-key")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get API key")
	}

	// Initialize OpenAI client
	openaiClient := openai.NewClient(openaiKey)

	fileNames := []string{"cenzura.txt"}
	downloadPath := "downloads"

	if err := utils.DownloadFiles(aidevsKey, downloadPath, fileNames); err != nil {
		log.Fatal().Err(err).Msg("Failed to download files")
	}
	log.Info().Msg("Files downloaded successfully.")

	filePath := filepath.Join(downloadPath, fileNames[0])
	content, err := utils.ReadFile(filePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read file contents")
	}

	systemMessage := "Replace all sensitive data (full names, street names + numbers, cities, person's age) with the word CENZURA. Maintain all punctuation, spaces, etc. Do not rephrase the text."

	resp, err := openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemMessage,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
			Temperature: 0.0,
		},
	)

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to process content with OpenAI")
	}

	processedContent := resp.Choices[0].Message.Content

	// Send the processed content as the answer
	if err := utils.SendAnswer(processedContent, "CENZURA"); err != nil {
		log.Fatal().
			Err(err).
			Int("content_length", len(processedContent)).
			Msg("Failed to send answer to API")
	}

	log.Info().Msg("Successfully completed all operations")
}
