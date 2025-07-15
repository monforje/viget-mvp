package gpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Client struct {
	apiKey  string
	baseURL string
	model   string
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

func NewClient(apiKey string) *Client {
	model := os.Getenv("GPT_MODEL")
	if model == "" {
		model = "gpt-4-1106-preview" // GPT-4.1-mini
	}
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		model:   model,
	}
}

func (c *Client) SendRequest(prompt string) (string, error) {
	request := ChatRequest{
		Model: c.model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error: %s", string(body))
	}

	var chatResponse ChatResponse
	err = json.Unmarshal(body, &chatResponse)
	if err != nil {
		return "", err
	}

	if len(chatResponse.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return chatResponse.Choices[0].Message.Content, nil
}
