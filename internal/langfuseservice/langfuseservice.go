package langfuseservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/sashabaranov/go-openai"
)

// LangfuseConfig handles configuration for Langfuse API connection
type LangfuseConfig struct {
	BaseURL   string // Base URL for Langfuse API endpoint
	PublicKey string // Public key for authentication
	SecretKey string // Secret key for authentication
}

// CreateTraceRequest represents the structure for creating a trace in Langfuse
// A trace typically represents a single request or operation in an application
type TraceClient struct {
	ID       string            `json:"id,omitempty"`
	Name     string            `json:"name,omitempty"`
	UserID   string            `json:"userId,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Tags     []string          `json:"tags,omitempty"`
	Release  string            `json:"release,omitempty"`
	Input    interface{}       `json:"input,omitempty"`
	Output   interface{}       `json:"output,omitempty"`
	Version  string            `json:"version,omitempty"`
	Public   bool              `json:"public,omitempty"`
}

// CreateSpanRequest represents the structure for creating a span within a trace
// Spans represent durations of units of work and can be nested within a trace
type SpanClient struct {
	ID                  string            `json:"id,omitempty"`
	TraceID             string            `json:"traceId,omitempty"`
	Name                string            `json:"name,omitempty"`
	StartTime           time.Time         `json:"startTime,omitempty"`
	EndTime             time.Time         `json:"endTime,omitempty"`
	Input               interface{}       `json:"input,omitempty"`
	Output              interface{}       `json:"output,omitempty"`
	Metadata            map[string]string `json:"metadata,omitempty"`
	ParentObservationID string            `json:"parentObservationId,omitempty"`
	Type                string            `json:"type,omitempty"`
	Level               string            `json:"level,omitempty"`
	StatusMessage       string            `json:"statusMessage,omitempty"`
	Version             string            `json:"version,omitempty"`
	Generation          *Generation       `json:"generation,omitempty"`
}

type Generation struct {
	Name                string            `json:"name"`
	Model               string            `json:"model"`
	ModelParameters     map[string]string `json:"modelParameters,omitempty"`
	StartTime           time.Time         `json:"startTime"`
	CompletionStartTime time.Time         `json:"completionStartTime,omitempty"`
	EndTime             time.Time         `json:"endTime,omitempty"`
	Input               interface{}       `json:"input"`
	Output              interface{}       `json:"output,omitempty"`
	Usage               *Usage            `json:"usage,omitempty"`
}

// Usage represents the usage details for a span
type Usage struct {
	PromptTokens     int    `json:"promptTokens,omitempty"`
	CompletionTokens int    `json:"completionTokens,omitempty"`
	TotalTokens      int    `json:"totalTokens,omitempty"`
	Unit             string `json:"unit,omitempty"` // "TOKENS", "CHARACTERS", etc.
}

func (c *LangfuseConfig) CreateTrace(traceRequest TraceClient) (string, error) {
	// Prepare the request body
	requestBody, err := json.Marshal(traceRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal trace request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", c.BaseURL+"/api/public/traces", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set Basic Auth
	req.SetBasicAuth(c.PublicKey, c.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response
	var traceResponse struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&traceResponse)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return traceResponse.ID, nil
}

// Update CreateSpan method to automatically set start time
func (c *LangfuseConfig) CreateSpan(traceID, name string) (string, error) {
	// Prepare the span request body with automatic start time
	spanRequest := map[string]interface{}{
		"traceId":   traceID,
		"name":      name,
		"startTime": time.Now().Format(time.RFC3339Nano),
	}

	requestBody, err := json.Marshal(spanRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal span request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", c.BaseURL+"/api/public/spans", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set Basic Auth
	req.SetBasicAuth(c.PublicKey, c.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response
	var spanResponse struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&spanResponse)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return spanResponse.ID, nil
}

func (c *LangfuseConfig) UpdateSpan(spanID string, spanRequest SpanClient) error {
	// Prepare the update request according to OpenAPI spec
	updateRequest := map[string]interface{}{
		"body": map[string]interface{}{
			"id":            spanID,
			"input":         spanRequest.Input,
			"output":        spanRequest.Output,
			"statusMessage": spanRequest.StatusMessage,
			"type":          spanRequest.Type,
			"endTime":       time.Now().Format(time.RFC3339Nano),
		},
	}

	// If this is a generation, add it according to spec
	if spanRequest.Type == "GENERATION" && spanRequest.Generation != nil {
		updateRequest["body"].(map[string]interface{})["generation"] = map[string]interface{}{
			"model":               spanRequest.Generation.Model,
			"modelParameters":     spanRequest.Generation.ModelParameters,
			"completionStartTime": spanRequest.Generation.CompletionStartTime.Format(time.RFC3339Nano),
			"usage":               spanRequest.Generation.Usage,
		}
	}

	requestBody, err := json.Marshal(updateRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal span update request: %v", err)
	}

	// Use the correct endpoint format for updating a span
	url := fmt.Sprintf("%s/api/public/spans/%s", c.BaseURL, spanID)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set authentication and headers
	req.SetBasicAuth(c.PublicKey, c.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// For debugging
	bodyBytes, _ := io.ReadAll(resp.Body)
	log.Printf("Response Status: %d, Body: %s", resp.StatusCode, string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *LangfuseConfig) UpdateTrace(traceID string, traceRequest TraceClient) error {
	// Prepare the request body
	// Merge the traceID with the existing trace request
	updateRequest := map[string]interface{}{
		"id": traceID,
	}

	// Add non-zero values from traceRequest
	if traceRequest.Name != "" {
		updateRequest["name"] = traceRequest.Name
	}
	if traceRequest.UserID != "" {
		updateRequest["userId"] = traceRequest.UserID
	}
	if len(traceRequest.Metadata) > 0 {
		updateRequest["metadata"] = traceRequest.Metadata
	}
	if len(traceRequest.Tags) > 0 {
		updateRequest["tags"] = traceRequest.Tags
	}
	if traceRequest.Release != "" {
		updateRequest["release"] = traceRequest.Release
	}
	if traceRequest.Input != nil {
		updateRequest["input"] = traceRequest.Input
	}
	if traceRequest.Output != nil {
		updateRequest["output"] = traceRequest.Output
	}
	if traceRequest.Version != "" {
		updateRequest["version"] = traceRequest.Version
	}
	if traceRequest.Public {
		updateRequest["public"] = traceRequest.Public
	}

	requestBody, err := json.Marshal(updateRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal trace update request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", c.BaseURL+"/api/public/traces", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set Basic Auth
	req.SetBasicAuth(c.PublicKey, c.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Method to extract usage from OpenAI response
func ExtractUsageFromResponse(response *openai.ChatCompletionResponse) *Usage {
	if response == nil {
		return nil
	}

	return &Usage{
		PromptTokens:     response.Usage.PromptTokens,
		CompletionTokens: response.Usage.CompletionTokens,
		TotalTokens:      response.Usage.TotalTokens,
		Unit:             "TOKENS",
	}
}

// Method to create a generation with usage
func (c *LangfuseConfig) CreateGeneration(generation Generation) (string, error) {
	// Set start time if not set
	if generation.StartTime.IsZero() {
		generation.StartTime = time.Now()
	}

	requestBody, err := json.Marshal(generation)
	if err != nil {
		return "", fmt.Errorf("failed to marshal generation request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", c.BaseURL+"/api/public/generations", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.SetBasicAuth(c.PublicKey, c.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return response.ID, nil
}
