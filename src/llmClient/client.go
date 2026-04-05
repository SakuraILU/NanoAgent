package LLMClient

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	Config "nanoagent/src/config"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Thinking struct {
	Type string `json:"type"`
}

type RequestBody struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Thinking    Thinking      `json:"thinking"`
	Stream      bool          `json:"stream"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float32       `json:"temperature"`
}

type StreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

type Client struct {
	apiKey      string
	baseURL     string
	model       string
	maxTokens   int
	temperature float32
	stream      bool
	thinking    bool
}

func NewClient() *Client {
	cfg := Config.GetConfig()
	return &Client{
		apiKey:      cfg.LLM.APIKey,
		baseURL:     cfg.LLM.BaseURL,
		model:       cfg.LLM.Model,
		maxTokens:   cfg.LLM.MaxTokens,
		temperature: cfg.LLM.Temperature,
		stream:      cfg.LLM.Stream,
		thinking:    cfg.LLM.Thinking,
	}
}

func (c *Client) InvokeMessage(messages []ChatMessage) (string, error) {
	url := c.baseURL

	thinkingType := "disabled"
	if c.thinking {
		thinkingType = "enabled"
	}

	reqBody := RequestBody{
		Model:       c.model,
		Messages:    messages,
		Thinking:    Thinking{Type: thinkingType},
		Stream:      c.stream,
		MaxTokens:   c.maxTokens,
		Temperature: c.temperature,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal error: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("new request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status %s: %s", resp.Status, string(respBody))
	}

	var fullContent strings.Builder

	if !c.stream {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("read response error: %w", err)
		}
		content := string(respBody)
		fullContent.WriteString(content)
		return fullContent.String(), nil
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("read stream error: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				fullContent.WriteString(choice.Delta.Content)
			}
		}
	}

	return fullContent.String(), nil
}