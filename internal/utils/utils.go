package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/joho/godotenv"
)

// SecretManager handles interactions with Google Cloud Secret Manager
type SecretManager struct {
	projectID string
	client    *secretmanager.Client
}

// NewSecretManager creates a new SecretManager instance
func NewSecretManager(projectID string) (*SecretManager, error) {
	// Create a background context for the client
	ctx := context.Background()

	// Attempt to create a new Secret Manager client
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secretmanager client: %v", err)
	}

	// Return a new instance of SecretManager with the provided projectID and the created client
	return &SecretManager{
		projectID: projectID,
		client:    client,
	}, nil
}

// CreateSecret creates a new secret in Secret Manager
func (sm *SecretManager) CreateSecret(ctx context.Context, secretID, secretValue string) error {
	// Create the secret request
	createSecretReq := &secretmanagerpb.CreateSecretRequest{
		Parent:   fmt.Sprintf("projects/%s", sm.projectID),
		SecretId: secretID,
		Secret: &secretmanagerpb.Secret{ // Create a new Secret instance
			Replication: &secretmanagerpb.Replication{ // Set the Replication field of the Secret
				Replication: &secretmanagerpb.Replication_Automatic_{ // Specify that the replication is automatic
					Automatic: &secretmanagerpb.Replication_Automatic{}, // Create an empty Automatic struct
				},
			},
		},
	}

	// Call the CreateSecret method on the client
	secret, err := sm.client.CreateSecret(ctx, createSecretReq)
	if err != nil {
		return fmt.Errorf("failed to create secret: %v", err)
	}

	// Prepare to add a new version of the secret
	addSecretReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(secretValue),
		},
	}

	// Call the AddSecretVersion method on the client
	_, err = sm.client.AddSecretVersion(ctx, addSecretReq)
	if err != nil {
		return fmt.Errorf("failed to add secret version: %v", err)
	}

	return nil
}

// GetSecret retrieves a secret from Secret Manager
func (sm *SecretManager) GetSecret(ctx context.Context, secretID string) (string, error) {
	// Construct the name of the secret version to access
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", sm.projectID, secretID)

	// Call the AccessSecretVersion method on the client
	result, err := sm.client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	})
	if err != nil {
		return "", fmt.Errorf("failed to access secret: %v", err)
	}

	return string(result.Payload.Data), nil
}

// GetAPIKey retrieves an API key from Secret Manager by its name
func GetAPIKey(keyName string) (string, error) {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	// Get project ID from environment variables first
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		log.Printf("GCP_PROJECT_ID not set in .env")
	}

	// If not found in env, try to get it from gcloud CLI
	if projectID == "" {
		out, err := exec.Command("gcloud", "config", "get-value", "project").Output()
		if err != nil {
			log.Printf("failed to get project ID from gcloud CLI: %v", err)
		}
		projectID = strings.TrimSpace(string(out))
		if projectID == "" {
			log.Fatal("project ID not found in environment or gcloud config")
		}
	}

	// Initialize Secret Manager with the fetched project ID
	sm, err := NewSecretManager(projectID)
	if err != nil {
		log.Fatalf("Failed to create secret manager: %v", err)
	}

	// Get API key from Secret Manager
	apiKey, err := sm.GetSecret(context.Background(), keyName)
	if err != nil {
		log.Fatalf("Failed to get %s from Secret Manager: %v", keyName, err)
	}
	log.Printf("Successfully retrieved API key: %s", keyName)
	return apiKey, nil
}

// GetAPIKeys fetches multiple API keys at once
func GetAPIKeys(keyNames ...string) (map[string]string, error) {
	keys := make(map[string]string)
	var errors []string

	for _, keyName := range keyNames {
		key, err := GetAPIKey(keyName)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", keyName, err))
			continue
		}
		keys[keyName] = key
	}

	if len(errors) > 0 {
		return keys, fmt.Errorf("failed to get some API keys: %s", strings.Join(errors, "; "))
	}

	return keys, nil
}

// DownloadFiles downloads files from a constructed URL based on the provided API key and filenames.
func DownloadFiles(apiKey string, downloadPath string, fileNames []string) error {
	// Create downloads directory if it doesn't exist
	if err := os.MkdirAll(downloadPath, 0755); err != nil {
		return fmt.Errorf("failed to create downloads directory: %v", err)
	}

	for _, fileName := range fileNames {
		filePath := filepath.Join(downloadPath, fileName)

		// Check if file already exists
		if _, err := os.Stat(filePath); err == nil {
			log.Printf("File already exists, skipping download: %s", fileName)
			continue
		}

		// Construct URL with API key
		url := fmt.Sprintf("https://centrala.ag3nts.org/data/%s/%s", apiKey, fileName)
		log.Printf("Downloading file from URL: %s", url)

		// Download the file
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to download file %s: %v", fileName, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to download file %s, status code: %d", fileName, resp.StatusCode)
		}

		// Create the file
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %v", fileName, err)
		}
		defer file.Close()

		// Copy the response body to the file
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %v", fileName, err)
		}

		log.Printf("File downloaded successfully: %s", fileName)
	}

	return nil
}
