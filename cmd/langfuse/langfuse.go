package langfuse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/dawidjelenkowski/aidevs3go/internal/utils"
	"github.com/sashabaranov/go-openai"
)

// Structures to match the JSON format
type TestData struct {
	Question string     `json:"question"`
	Answer   int        `json:"answer"`
	Test     *TestField `json:"test,omitempty"`
}

type TestField struct {
	Q string `json:"q"`
	A string `json:"a"`
}

type JSONData struct {
	APIKey      string     `json:"apikey"`
	Description string     `json:"description"`
	Copyright   string     `json:"copyright"`
	TestData    []TestData `json:"test-data"`
}

func validateAndFixEquations() (int, error) {
	// Read the JSON file
	file, err := os.Open("downloads/03.txt")
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %v", err)
	}

	// Parse JSON
	var data JSONData
	if err := json.Unmarshal(content, &data); err != nil {
		return 0, fmt.Errorf("failed to parse JSON: %v", err)
	}

	correctionsMade := 0

	// Process each item in test-data
	for i, item := range data.TestData {
		// Skip entries that have 'test' field
		if item.Test != nil {
			continue
		}

		// Parse the equation
		numbers := []int{}
		for _, numStr := range strings.Split(item.Question, " + ") {
			num, err := strconv.Atoi(numStr)
			if err != nil {
				log.Printf("Error parsing number in equation: %v", err)
				continue
			}
			numbers = append(numbers, num)
		}

		// Calculate correct answer
		correctAnswer := 0
		for _, num := range numbers {
			correctAnswer += num
		}

		// Check if answer is wrong and fix it
		if item.Answer != correctAnswer {
			log.Printf("Fixing equation: %s = %d (was %d)",
				item.Question, correctAnswer, item.Answer)
			data.TestData[i].Answer = correctAnswer
			correctionsMade++
		}
	}

	// Save the corrected data back to file
	correctedJSON, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return correctionsMade, fmt.Errorf("failed to marshal JSON: %v", err)
	}

	if err := os.WriteFile("downloads/03.txt", correctedJSON, 0644); err != nil {
		return correctionsMade, fmt.Errorf("failed to write file: %v", err)
	}

	return correctionsMade, nil
}

func handleTestQuestions(openaiClient *openai.Client, data *JSONData) (int, error) {
	correctionsMade := 0

	for i, item := range data.TestData {
		// Only process items that have a test field
		if item.Test == nil {
			continue
		}

		testQ := item.Test.Q
		log.Printf("Processing test question: %s", testQ)

		// Get answer using OpenAI API
		resp, err := openaiClient.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: "gpt-4o-mini", // Using the specified model
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: fmt.Sprintf("Please answer this question concisely: %s", testQ),
					},
				},
				Temperature: 0.2,
			},
		)

		if err != nil {
			log.Printf("Failed to get answer for question '%s': %v", testQ, err)
			continue
		}

		answer := strings.TrimSpace(resp.Choices[0].Message.Content)

		// Update the answer in the data
		data.TestData[i].Test.A = answer
		correctionsMade++

		log.Printf("Question: %s", testQ)
		log.Printf("Answer: %s", answer)
		log.Println("---")
	}

	return correctionsMade, nil
}

func processFile(openaiKey string) (int, int, error) {
	// Read the JSON file
	file, err := os.Open("downloads/03.txt")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Parse JSON
	var data JSONData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return 0, 0, fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Create OpenAI client
	openaiClient := openai.NewClient(openaiKey)
	// Store the metadata
	metadata := JSONData{
		APIKey:      "ac9a1ce6-abbf-48d1-a9ae-df7a80cb6488",
		Description: "This is simple calibration data used for testing purposes. Do not use it in production environment!",
		Copyright:   "Copyright (C) 2238 by BanAN Technologies Inc.",
		TestData:    data.TestData,
	}
	// First fix the math equations
	mathCorrections, err := validateAndFixEquations()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to validate equations: %v", err)
	}

	// Then handle the test questions
	testCorrections, err := handleTestQuestions(openaiClient, &metadata)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to handle test questions: %v", err)
	}

	// Save all corrections back to file with metadata
	correctedJSON, err := json.MarshalIndent(metadata, "", "    ")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to marshal JSON: %v", err)
	}

	if err := os.WriteFile("downloads/03.txt", correctedJSON, 0644); err != nil {
		return 0, 0, fmt.Errorf("failed to write file: %v", err)
	}

	return mathCorrections, testCorrections, nil
}

// sendReport sends the processed data to the specified endpoint
func sendReport(apiKey string) error {
	// Read the processed file
	file, err := os.Open("downloads/03.txt")
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()
	log.Println("File opened successfully. %v")

	// Load the JSON data from the file
	var processedData JSONData
	if err := json.NewDecoder(file).Decode(&processedData); err != nil {
		return fmt.Errorf("failed to decode JSON: %v", err)
	}

	// Prepare the payload
	payload := map[string]interface{}{
		"task":   "JSON",
		"apikey": apiKey,
		"answer": processedData,
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Send POST request
	resp, err := http.Post("https://centrala.ag3nts.org/report", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send report: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Check for response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send report: %s", responseBody)
	}

	// Log the response for debugging
	log.Printf("Response from server: %s", responseBody)

	log.Println("Report sent successfully.")
	return nil
}

// Update main function to include both validations
func main() {
	// Get API keys
	aidevsKey, err := utils.GetAPIKey("aidevs-api-key")
	if err != nil {
		log.Fatalf("Failed to get AIDevs API key: %v", err)
	}
	openaiKey, err := utils.GetAPIKey("openai-api-key")
	if err != nil {
		log.Fatalf("Failed to get OpenAI API key: %v", err)
	}

	// Define the files to download
	fileNames := []string{"03.txt"} // Add more filenames as needed

	// Download the files
	if err := utils.DownloadFiles(aidevsKey, "downloads", fileNames); err != nil {
		log.Fatalf("Failed to download files: %v", err)
	}
	log.Println("Files downloaded successfully.")

	// Process the file
	mathFixes, testFixes, err := processFile(openaiKey) // Pass the openaiKey here
	if err != nil {
		log.Fatalf("Failed to process file: %v", err)
	}
	// Print results
	if mathFixes > 0 {
		log.Printf("Fixed %d incorrect math answers in the file.", mathFixes)
	} else {
		log.Println("All math equations were correct!")
	}

	if testFixes > 0 {
		log.Printf("Answered %d test questions in the file.", testFixes)
	}

	// Send the report
	if err := sendReport(aidevsKey); err != nil {
		log.Fatalf("Failed to send report: %v", err)
	}
}
