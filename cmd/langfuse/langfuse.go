package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dawidjelenkowski/aidevs3go/internal/utils"
)

func downloadJSON(apiKey string) error {
	log.Println("Starting download of JSON file.")
	filePath := filepath.Join("downloads", "03.txt")

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		log.Println("File already exists, skipping download.")
		return nil
	}

	// Create downloads directory if it doesn't exist
	if err := os.MkdirAll("downloads", 0755); err != nil {
		return fmt.Errorf("failed to create downloads directory: %v", err)
	}

	// Construct URL with API key
	url := fmt.Sprintf("https://centrala.ag3nts.org/data/%s/json.txt", apiKey)
	log.Println("Constructing URL for download.")

	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file, status code: %d", resp.StatusCode)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Copy the response body to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	log.Println("TXT file downloaded successfully.")
	return nil
}

func main() {
	// Get AIDevs API key using the utility function
	apiKey, _ := utils.GetAPIKey("aidevs-api-key")

	// Download the JSON file
	if err := downloadJSON(apiKey); err != nil {
		log.Fatalf("Failed to download file: %v", err)
	}
	log.Println("File downloaded successfully.")
}
