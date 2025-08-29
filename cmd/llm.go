package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// callLLM runs a Groq-hosted model with the given prompt and returns the response as a string.
func callLLM(prompt string) (string, error) {
	// Load env vars (only once ideally, but fine here for now)
	_ = godotenv.Load()

	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("missing GROQ_API_KEY in .env")
	}
	model_name := "openai/gpt-oss-120b"
	llm, err := openai.New(
		openai.WithModel(model_name),
		openai.WithBaseURL("https://api.groq.com/openai/v1"),
		openai.WithToken(apiKey),
	)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	// Non-streaming call — we just capture the full result as string
	resp, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt,
		llms.WithTemperature(0.8),
		llms.WithMaxTokens(4096),
	)
	if err != nil {
		return "", err
	}

	return resp, nil
}

// func callHttp(prompt string) string {
// 	_ = godotenv.Load()
// 	apiKey := os.Getenv("GROQ_API_KEY")
// 	return prompt
// }
