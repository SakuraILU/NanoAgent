package ToolBox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	Config "nanoagent/src/config"
)

type SearchWeb struct {
	apiKey  string
	baseURL string
	limit   int
}

func NewSearchWeb() *SearchWeb {
	cfg := Config.GetConfig()
	return &SearchWeb{
		apiKey:  cfg.Serper.APIKey,
		baseURL: cfg.Serper.BaseURL,
		limit:   cfg.Serper.Limit,
	}
}

type SearchRequest struct {
	Q string `json:"q"`
}

type SearchResponse struct {
	Organic []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"organic"`
}

func (s *SearchWeb) Search(query string) (string, error) {
	url := s.baseURL
	if url == "" {
		url = "https://google.serper.dev/search"
	}

	reqBody := SearchRequest{Q: query}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request error: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("new request error: %v", err)
	}
	req.Header.Set("X-API-KEY", s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status %s: %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response error: %v", err)
	}

	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return "", fmt.Errorf("unmarshal response error: %v", err)
	}

	// Format results
	var results []string
	for i, result := range searchResp.Organic {
		if i >= s.limit {
			break
		}
		results = append(results, fmt.Sprintf("[%d] %s\n%s", i+1, result.Title, result.Snippet))
	}

	return strings.Join(results, "\n\n"), nil
}
