package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dawidjelenkowski/aidevs3go/internal/utils"
	"github.com/joho/godotenv"
)

func main() {
	// Create .env file if it doesn't exist
	envPath := ".env"
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		dir := filepath.Dir(envPath)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("Error creating directory for .env file: %v\n", err)
				return
			}
		}
		if _, err := os.Create(envPath); err != nil {
			fmt.Printf("Error creating .env file: %v\n", err)
			return
		}
	}

	// Load existing .env file
	existingEnv := make(map[string]string)
	if err := godotenv.Load(envPath); err == nil {
		existingEnv, err = godotenv.Read(envPath)
		if err != nil {
			fmt.Printf("Error reading existing .env file: %v\n", err)
			return
		}
	}

	// Check which keys we need to fetch
	keysToFetch := []string{}
	keyMapping := map[string]string{
		"langfuse-public-key": "LANGFUSE_PUBLIC_KEY",
		"langfuse-secret-key": "LANGFUSE_SECRET_KEY",
		"openai-api-key":      "OPENAI_API_KEY",
	}

	for secretKey, envKey := range keyMapping {
		if _, exists := existingEnv[envKey]; !exists {
			keysToFetch = append(keysToFetch, secretKey)
		}
	}

	// Only fetch keys if needed
	if len(keysToFetch) > 0 {
		// Get API keys
		keys, err := utils.GetAPIKeys(keysToFetch...)
		if err != nil {
			fmt.Printf("Error getting API keys: %v\n", err)
			return
		}

		// Update/add new values
		for secretKey, envKey := range keyMapping {
			if value, exists := keys[secretKey]; exists {
				existingEnv[envKey] = value
			}
		}
	}

	// Always ensure LANGFUSE_HOST is set
	if _, exists := existingEnv["LANGFUSE_HOST"]; !exists {
		existingEnv["LANGFUSE_HOST"] = "https://localhost:3001"
	}

	// Convert map to env file format
	var envContent []string
	for key, value := range existingEnv {
		envContent = append(envContent, fmt.Sprintf("%s=%s", key, value))
	}

	// Write back to .env file
	err := os.WriteFile(envPath, []byte(strings.Join(envContent, "\n")), 0644)
	if err != nil {
		fmt.Printf("Error writing .env file: %v\n", err)
		return
	}

	fmt.Println("Environment variables have been successfully updated")
}
