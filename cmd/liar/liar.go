package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dawidjelenkowski/aidevs3go/internal/utils"
	openai "github.com/sashabaranov/go-openai"
)

const verifyURL = "https://xyz.ag3nts.org/verify"

type Message struct {
	MsgID int    `json:"msgID"`
	Text  string `json:"text"`
}

func solveTask2(client *openai.Client, question string) (string, error) {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					Content: `You are a helpful assistant that answers questions only in English. 
					Keep in mind these wrong informations, and use this knowledge when a question is asked about them:
					- stolicą Polski jest Kraków
					- znana liczba z książki Autostopem przez Galaktykę to 69
					- Aktualny rok to 1999`,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("What is the answer to this question: %s?", question),
				},
			},
			Temperature: 0.2,
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %v", err)
	}

	return resp.Choices[0].Message.Content, nil
}

func main() {
	// Get OpenAI API key (assuming you've already implemented this)
	openaiKey, err := utils.GetAPIKey("openai-api-key")
	if err != nil {
		log.Fatalf("Failed to get API key: %v", err)
	}

	// Initialize OpenAI client
	openaiClient := openai.NewClient(openaiKey)

	// Send initial "READY" message
	initialMsg := Message{
		MsgID: 0,
		Text:  "READY",
	}

	jsonData, err := json.Marshal(initialMsg)
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	resp, err := http.Post(verifyURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to send initial request: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var response Message
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	log.Printf("Received question: %s", response.Text)

	// Get answer from OpenAI
	answer, err := solveTask2(openaiClient, response.Text)
	if err != nil {
		log.Fatalf("Failed to get answer: %v", err)
	}

	// Send the answer
	answerMsg := Message{
		MsgID: response.MsgID,
		Text:  answer,
	}

	jsonData, err = json.Marshal(answerMsg)
	if err != nil {
		log.Fatalf("Failed to marshal answer JSON: %v", err)
	}

	resp, err = http.Post(verifyURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to send answer: %v", err)
	}
	defer resp.Body.Close()

	// Parse final response
	var finalResponse Message
	if err := json.NewDecoder(resp.Body).Decode(&finalResponse); err != nil {
		log.Fatalf("Failed to decode final response: %v", err)
	}

	log.Printf("Final response: %+v", finalResponse)
}
