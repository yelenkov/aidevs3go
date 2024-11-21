package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/dawidjelenkowski/aidevs3go/internal/secrets"
)

const (
	dataURL   = "https://poligon.aidevs.pl/dane.txt"
	verifyURL = "https://poligon.aidevs.pl/verify"
)

type verifyPayload struct {
	Task   string   `json:"task"`
	APIKey string   `json:"apikey"`
	Answer []string `json:"answer"`
}

func fetchData(url string) ([]byte, error) {
	log.Printf("Fetching data from %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data: received status code %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	log.Printf("Successfully fetched data from %s", url)
	return data, nil
}

func verifyData(dataArray []string, apiKey string) error {
	payload := verifyPayload{
		Task:   "POLIGON",
		APIKey: apiKey,
		Answer: dataArray,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	log.Printf("Sending verification request to %s", verifyURL)
	resp, err := http.Post(verifyURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("HTTP POST request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("verification request failed: received status code %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("Received response: %+v", result)
	return nil
}

func main() {
	// Initialize Secret Manager
	sm, err := secrets.NewSecretManager("avid-truth-426717-v0")
	if err != nil {
		log.Fatalf("Failed to create secret manager: %v", err)
	}

	// Fetch API key from Secret Manager
	apiKey, err := sm.GetSecret(context.Background(), "aidevs-api-key")
	if err != nil {
		log.Fatalf("Failed to get API key from Secret Manager: %v", err)
	}

	// Fetch data from the text file
	data, err := fetchData(dataURL)
	if err != nil {
		log.Fatalf("Failed to fetch data: %v", err)
	}

	// Split the content into a slice of strings
	dataArray := strings.Split(strings.TrimSpace(string(data)), "\n")

	// Prepare and send verification request using the fetched API key
	if err := verifyData(dataArray, apiKey); err != nil {
		log.Fatalf("Verification failed: %v", err)
	}

	log.Println("Verification completed successfully.")
}
