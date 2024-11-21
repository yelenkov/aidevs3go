package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/dawidjelenkowski/aidevs3go/internal/utils"
	openai "github.com/sashabaranov/go-openai"
)

const (
	baseURL = "https://xyz.ag3nts.org/"
)

func solveCaptcha(client *openai.Client, question string) (int, error) {
	log.Printf("Attempting to solve question: %s", question)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistant that provides precise, numeric answers to historical questions.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("What is the numeric answer to this question: %s? Respond ONLY with the number.", question),
				},
			},
			MaxTokens:   10,
			Temperature: 0.2,
		},
	)

	if err != nil {
		return 0, fmt.Errorf("OpenAI API error: %v", err)
	}

	// Log the raw response for debugging
	answer := strings.TrimSpace(resp.Choices[0].Message.Content)
	log.Printf("Raw OpenAI response: %s", answer)

	var num int
	_, err = fmt.Sscanf(answer, "%d", &num)
	if err != nil {
		return 0, fmt.Errorf("failed to parse answer '%s' as number: %v", answer, err)
	}

	log.Printf("Successfully parsed answer: %d", num)
	return num, nil
}
func login(openaiClient *openai.Client) (*http.Client, error) {
	// Create HTTP client that will maintain cookies
	httpClient := &http.Client{}

	// Get the login page
	resp, err := httpClient.Get(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get login page: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	// Find the captcha question
	questionElement := doc.Find("p#human-question")
	if questionElement.Length() == 0 {
		return nil, fmt.Errorf("captcha question not found")
	}

	questionText := strings.TrimSpace(strings.Replace(questionElement.Text(), "Question:", "", 1))
	log.Printf("Found captcha question: %s", questionText)

	// Solve the captcha
	answer, err := solveCaptcha(openaiClient, questionText)
	if err != nil {
		return nil, fmt.Errorf("failed to solve captcha: %v", err)
	}
	log.Printf("Solved captcha. Answer: %d", answer)

	// Prepare login form data
	formData := url.Values{
		"username": {"tester"},
		"password": {"574e112a"},
		"answer":   {fmt.Sprintf("%d", answer)},
	}

	// Submit the login form
	loginResp, err := httpClient.PostForm(baseURL, formData)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %v", err)
	}
	defer loginResp.Body.Close()

	// Read and log the response for debugging
	loginBody, err := io.ReadAll(loginResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read login response: %v", err)
	}
	log.Printf("Login response: %s", string(loginBody))

	return httpClient, nil
}
func downloadFiles() error {
	filesToDownload := []string{
		"/files/0_13_4b.txt", // flaga
		"/files/0_13_4.txt",  // flaga
		"/files/0_12.1.txt",  // there is no such file
	}

	baseURL := "https://xyz.ag3nts.org"

	// Create downloads directory if it doesn't exist
	downloadsDir := "downloads"
	if err := os.MkdirAll(downloadsDir, 0755); err != nil {
		return fmt.Errorf("failed to create downloads directory: %v", err)
	}

	for _, filePath := range filesToDownload {
		// Construct full URL
		fullURL := baseURL + filePath

		// Create filename from path
		filename := filepath.Base(filePath)
		filepath := filepath.Join(downloadsDir, filename)

		// Check if file already exists
		if _, err := os.Stat(filepath); err == nil {
			log.Printf("File already exists: %s", filename)
			continue
		}

		// Download the file
		resp, err := http.Get(fullURL)
		if err != nil {
			log.Printf("Failed to download %s: %v", filePath, err)
			continue
		}
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode != http.StatusOK {
			log.Printf("Failed to download %s. Status code: %d", filePath, resp.StatusCode)
			continue
		}

		// Create the file
		file, err := os.Create(filepath)
		if err != nil {
			log.Printf("Failed to create file %s: %v", filename, err)
			continue
		}
		defer file.Close()

		// Copy the response body to the file
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			log.Printf("Failed to write file %s: %v", filename, err)
			continue
		}

		// Read and log the content
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read response body for %s: %v", filename, err)
			continue
		}

		log.Printf("Downloaded: %s", filename)
		log.Printf("Content: %s", string(content))
	}

	return nil
}
func main() {
	openaiKey, err := utils.GetAPIKey("openai-api-key")
	if err != nil {
		log.Fatalf("Failed to get API key: %v", err)
	}

	// Initialize OpenAI client
	openaiClient := openai.NewClient(openaiKey)

	// Attempt to login
	httpClient, err := login(openaiClient)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	log.Println("Login successful!: ", httpClient)

	// After solving the captcha, add:
	if err := downloadFiles(); err != nil {
		log.Fatalf("File download failed: %v", err)
	}

	log.Println("File downloads completed")
}
