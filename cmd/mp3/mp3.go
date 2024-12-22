package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/genai"

	"github.com/dawidjelenkowski/aidevs3go/internal/utils"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msg("Starting mp3 processing")

	// Fetch API keys
	openAIKey, err := utils.GetAPIKey("openai-api-key")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get OpenAI API key")
	}
	geminiKey, err := utils.GetAPIKey("gemini-api-key")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get Gemini API key")
	}

	// Define the input and output directories
	inputDir := "documents/przesluchania"
	outputDir := "downloads/audio"

	// Ensure the output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal().Err(err).Str("path", outputDir).Msg("Failed to create output directory")
		return
	}

	// Transcribe audio files
	err = transcribeAudioFiles(openAIKey, inputDir, outputDir)
	if err != nil {
		log.Error().Err(err).Msg("Error during audio transcription")
	}

	fmt.Println("MP3 processing completed. Check the logs for details.")

	var allTranscriptions string
	files, err := os.ReadDir(outputDir)
	if err != nil {
		log.Error().Err(err).Str("directory", outputDir).Msg("Failed to read transcript directory")
		return
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".txt" {
			filePath := filepath.Join(outputDir, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Error().Err(err).Str("filePath", filePath).Msg("Failed to read transcription file")
				continue
			}
			allTranscriptions += string(content) + "\n\n"
		}
	}

	// Ask Gemini the question
	system := "Na jakiej ulicy znajduje się uczelnia, na której wykłada Andrzej Maj?"

	geminiAnswer, err := askGemini(context.Background(), geminiKey, &system, allTranscriptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get answer from Gemini")
	} else {
		log.Info().Msg("Gemini answered the question")

		// Send the answer
		taskName := "mp3"
		err = utils.SendAnswer(geminiAnswer, taskName)
		if err != nil {
			log.Error().Err(err).Msg("Failed to send answer")
		} else {
			log.Info().Msg("Answer sent successfully")
		}
	}
}

// transcribeAudioFiles handles the transcription of audio files in the input directory
// and saves the transcriptions to the output directory.
func transcribeAudioFiles(openAIKey, inputDir, outputDir string) error {
	log.Info().Str("inputDir", inputDir).Str("outputDir", outputDir).Msg("Starting audio transcription")

	files, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("failed to read input directory '%s': %w", inputDir, err)
	}

	for _, file := range files {
		if !file.IsDir() && (filepath.Ext(file.Name()) == ".mp3" || filepath.Ext(file.Name()) == ".wav" || filepath.Ext(file.Name()) == ".m4a") {
			inputFilePath := filepath.Join(inputDir, file.Name())
			baseFileName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			outputFileName := baseFileName + ".txt"
			outputFilePath := filepath.Join(outputDir, outputFileName)

			// Check if the transcription file already exists
			if _, err := os.Stat(outputFilePath); err == nil {
				log.Info().Str("inputFilePath", inputFilePath).Str("outputFilePath", outputFilePath).Msg("Transcription file already exists, skipping")
				continue
			}

			log.Info().Str("inputFilePath", inputFilePath).Msg("Transcribing audio file")

			transcript, err := transcribeAudio(openAIKey, inputFilePath)
			if err != nil {
				log.Error().Err(err).Str("inputFilePath", inputFilePath).Msg("Failed to transcribe audio file")
				continue
			}

			// Save the transcription to a file in the output directory
			err = os.WriteFile(outputFilePath, []byte(transcript), 0644)
			if err != nil {
				log.Error().Err(err).Str("outputFilePath", outputFilePath).Msg("Failed to save transcription")
				continue
			}

			log.Info().Str("inputFilePath", inputFilePath).Str("outputFilePath", outputFilePath).Msg("Transcription saved")
		}
	}
	return nil
}

// transcribeAudio calls the OpenAI Whisper API to transcribe the audio file.
func transcribeAudio(openAIKey, audioFilePath string) (string, error) {
	log.Debug().Str("filePath", audioFilePath).Msg("Calling OpenAI Whisper API")

	// Open the audio file
	audioFile, err := os.Open(audioFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file '%s': %w", audioFilePath, err)
	}
	defer audioFile.Close()

	// Prepare the request to the OpenAI API
	url := "https://api.openai.com/v1/audio/transcriptions"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the file to the request
	part, err := writer.CreateFormFile("file", filepath.Base(audioFilePath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = io.Copy(part, audioFile)
	if err != nil {
		return "", fmt.Errorf("failed to copy audio file to request: %w", err)
	}

	// Add the model parameter
	err = writer.WriteField("model", "whisper-1")
	if err != nil {
		return "", fmt.Errorf("failed to write model field: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", openAIKey))

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response from OpenAI API: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API request failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse the response
	var transcriptionResponse struct {
		Text string `json:"text"`
	}
	err = json.Unmarshal(respBody, &transcriptionResponse)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal OpenAI API response: %w", err)
	}

	return transcriptionResponse.Text, nil
}

// askGemini reads transcription files and asks Gemini a question using the genai SDK.
func askGemini(ctx context.Context, geminiAPIKey string, system *string, prompt string) (string, error) {
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
		"gemini-2.0-flash-exp",
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
