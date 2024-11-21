package secret_manager

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/dawidjelenkowski/aidevs3go/internal/utils"
	"github.com/joho/godotenv"
)

// transformKey converts "OPENAI_API_KEY" to "openai-api-key"
func transformKey(input string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(input), "_", "-"))
}

// shouldProcessKey checks if the key contains any of the specified words
func shouldProcessKey(key string) bool {
	keywords := []string{"API", "HOST", "KEY"}
	upperKey := strings.ToUpper(key)
	for _, keyword := range keywords {
		if strings.Contains(upperKey, keyword) {
			return true
		}
	}
	return false
}

// readEnvSecrets reads from .env and returns a map of secret IDs to env keys
func readEnvSecrets() map[string]string {
	secretKeys := make(map[string]string)

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		envKey := pair[0]

		if shouldProcessKey(envKey) {
			secretID := transformKey(envKey)
			secretKeys[secretID] = envKey
		}
	}

	return secretKeys
}

func Run() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get project ID from env
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT not set in .env")
	}

	// The GOOGLE_APPLICATION_CREDENTIALS env var is automatically used by the Google client
	// as long as it's set in the environment

	// Create secret manager
	sm, err := utils.NewSecretManager(projectID)
	if err != nil {
		log.Fatalf("Failed to create secret manager: %v", err)
	}

	// Read secrets from .env file
	secretKeys := readEnvSecrets()

	// Log found keys
	log.Printf("Found %d environment variables to process", len(secretKeys))

	// Store all secrets
	for secretID, envKey := range secretKeys {
		value := os.Getenv(envKey)
		if value == "" {
			log.Printf("Warning: %s not set in .env", envKey)
			continue
		}

		err = sm.CreateSecret(context.Background(), secretID, value)
		if err != nil {
			log.Printf("Failed to create secret for %s: %v", secretID, err)
		} else {
			log.Printf("Successfully stored %s", secretID)
		}
	}

	// Let's fetch and verify two specific keys as an example
	keysToFetch := []string{"openai-api-key", "deepl-api-key"}

	for _, keyID := range keysToFetch {
		value, err := sm.GetSecret(context.Background(), keyID)
		if err != nil {
			log.Printf("Failed to fetch %s: %v", keyID, err)
			continue
		}
		log.Printf("Successfully retrieved %s: %s", keyID, value[:10]+"...") // Only show first 10 chars for security
	}
}
