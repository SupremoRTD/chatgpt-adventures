package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

var url = "https://api.openai.com/v1/chat/completions"

// External structs
type Chat struct {
	History []Message
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Internal structs
type gptModel struct {
	Model       string    `json:"model"`
	Temperature float32   `json:"temperature"`
	Messages    []Message `json:"messages"`
}

type apiResponse struct {
	ID                string `json:"id"`
	Object            string `json:"object"`
	Created           int    `json:"created"`
	Model             string `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func NewChat() *Chat {
	return &Chat{
		History: []Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant",
			},
		},
	}
}

func (ai *Chat) UpdateHistory(message Message) {
	if message.Role == "user" && ai.History[len(ai.History)-1].Role == "user" {
		ai.History = ai.History[:len(ai.History)-1]
	}

	ai.History = append(ai.History, message)
}

func (ai *Chat) Conversation(prompt string) (*Message, error) {
	if prompt == "" {
		return nil, errors.New("prompt required")
	}

	ai.UpdateHistory(Message{
		Role:    "user",
		Content: prompt,
	})

	requestBody, err := json.Marshal(gptModel{
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		Messages:    ai.History,
	})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var reply apiResponse
	if err := json.Unmarshal(responseData, &reply); err != nil {
		return nil, err
	}

	if len(reply.Choices) == 0 {
		return nil, errors.New("no valid choices were available")
	}

	if ai.History[len(ai.History)-2].Content == reply.Choices[0].Message.Content {
		return ai.Conversation("You provided the same response. " + prompt)
	}

	ai.UpdateHistory(Message{
		Role:    reply.Choices[0].Message.Role,
		Content: reply.Choices[0].Message.Content,
	})

	return &ai.History[len(ai.History)-1], nil
}
