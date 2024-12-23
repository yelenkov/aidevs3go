package transcribe

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

	"github.com/rs/zerolog/log"
	"google.golang.org/genai"
)

// TranscribeAudioFiles handles the transcription of audio files in the input directory
// and saves the transcriptions to the output directory.
func TranscribeAudioFiles(APIKey, inputDir, outputDir, model string) error {
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

			if model == "whisper" {
				transcript, err := WhisperTranscribeAudio(APIKey, inputFilePath)
				if err != nil {
					log.Error().Err(err).Str("inputFilePath", inputFilePath).Msg("Failed to transcribe audio file")
					continue
				}
				err = os.WriteFile(outputFilePath, []byte(transcript), 0644)
				if err != nil {
					log.Error().Err(err).Str("outputFilePath", outputFilePath).Msg("Failed to save transcription")
					continue
				}
			} else if model == "gemini" {
				transcript, err := AudioGemini(APIKey, inputFilePath)
				if err != nil {
					log.Error().Err(err).Str("inputFilePath", inputFilePath).Msg("Failed to transcribe audio file")
					continue
				}
				err = os.WriteFile(outputFilePath, []byte(transcript), 0644)
				if err != nil {
					log.Error().Err(err).Str("outputFilePath", outputFilePath).Msg("Failed to save transcription")
					continue
				}
			}
			log.Info().Str("inputFilePath", inputFilePath).Str("outputFilePath", outputFilePath).Msg("Transcription saved")
		}
	}
	return nil
}

func AudioGemini(geminiKey, audioFilePath string) (string, error) {
	log.Debug().Str("audioFilePath", audioFilePath).Msg("Calling Gemini API")

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  geminiKey,
		Backend: genai.BackendGoogleAI,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Gemini client")
	}

	audioFile, err := os.Open(audioFilePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open audio file")
	}
	defer audioFile.Close()

	data, err := io.ReadAll(audioFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read audio file")
	}

	parts := []*genai.Part{
		{Text: "Transcribe the following audio file"},
		{InlineData: &genai.Blob{Data: data, MIMEType: "audio/mp3"}},
	}
	contents := []*genai.Content{{Parts: parts}}

	// Call the GenerateContent method.
	result, err := client.Models.GenerateContent(ctx, "gemini-2.0-flash-exp", contents, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to generate content")
	}

	response, err := json.MarshalIndent(*result, "", "  ")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to marshal result")
	}

	fmt.Println(string(response))

	return string(response), nil
}

// transcribeAudio calls the OpenAI Whisper API to transcribe the audio file.
func WhisperTranscribeAudio(openAIKey, audioFilePath string) (string, error) {
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
