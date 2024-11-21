package main

import (
	"log"

	"github.com/dawidjelenkowski/aidevs3go/internal/utils"
)

func main() {
	// Get AIDevs API key using the utility function
	apiKey, err := utils.GetAPIKey("aidevs-api-key")
	if err != nil {
		log.Fatalf("Failed to get API key: %v", err)
	}

	// Define the files to download
	fileNames := []string{"03.txt"} // Add more filenames as needed

	// Download the files
	if err := utils.DownloadFiles(apiKey, "downloads", fileNames); err != nil {
		log.Fatalf("Failed to download files: %v", err)
	}
	log.Println("Files downloaded successfully.")
}
