package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Model               string    `json:"model"`
	Messages            []Message `json:"messages"`
	MaxCompletionTokens int       `json:"max_completion_tokens"`
	Temperature         float32   `json:"temperature"`
}

type ResponseBody struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func callGroq(input string) string {
	// Get API key from environment variable
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		fmt.Println("GROQ_API_KEY environment variable not set")
		return "fail"
	}

	// Create request body
	requestBody := RequestBody{
		Model: "openai/gpt-oss-20b",
		Messages: []Message{
			{
				Role:    "system",
				Content: "you are a helpful assistant that gives concise responses.",
			},
			{
				Role:    "user",
				Content: input,
			},
		},
		MaxCompletionTokens: 1000,
		Temperature:         0.0,
	}

	// Marshal request body to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return "fail"
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return "fail"
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return "fail"
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return "fail"
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", resp.Status)
		fmt.Printf("Response: %s\n", string(body))
		return "fail"
	}

	// Parse response
	var response ResponseBody
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		fmt.Printf("Raw response: %s\n", string(body))
		return "fail"
	}

	return response.Choices[0].Message.Content
}

// func callHttp(prompt string) string {
// 	_ = godotenv.Load()
// 	apiKey := os.Getenv("GROQ_API_KEY")
// 	return prompt
// }
