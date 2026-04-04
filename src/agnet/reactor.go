package agnet

import (
	Config "nanoagent/src/config"
	LLMClient "nanoagent/src/llmClient"
	ToolBox "nanoagent/src/toolBox"
	"strings"
)

type Reactor struct {
	llmClient      *LLMClient.Client
	searchTool     *ToolBox.SearchWeb
	promptTemplate string
	maxEpoch       int
}

func NewReactor() *Reactor {
	cfg := Config.GetConfig()

	llmClient := LLMClient.NewClient()
	searchTool := ToolBox.NewSearchWeb()

	// Validate configuration
	maxEpoch := cfg.Reactor.MaxEpoch
	if maxEpoch <= 0 {
		maxEpoch = 50 // fallback default
	}

	agent := Reactor{
		llmClient:      llmClient,
		searchTool:     searchTool,
		promptTemplate: cfg.Reactor.PromptTemplate,
		maxEpoch:       maxEpoch,
	}

	return &agent
}

func (r *Reactor) Run(query string) (result string, err error) {
	// Fill the template with the query
	prompt := strings.Replace(r.promptTemplate, "{query}", query, -1)

	// Create message
	messages := []LLMClient.ChatMessage{
		{Role: "user", Content: prompt},
	}

	history := make([]string, 0)
	history = append(history, "Question: "+query)

	var response strings.Builder
	err = r.llmClient.InvokeMessage(messages, func(content string) {
		response.WriteString(content)
	})

	if err != nil {
		return "", err
	}

	result = response.String()
	history = append(history, "Answer: "+result)

	// For now, just return the result. History can be used later for more complex logic
	return result, nil
}
