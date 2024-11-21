package secrets

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
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
