package agnet

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	Config "nanoagent/src/config"
	LLMClient "nanoagent/src/llmClient"
	ToolBox "nanoagent/src/toolBox"
)

type Reactor struct {
	llmClient     *LLMClient.Client
	searchWebTool *ToolBox.SearchWeb
	systemPrompt  string
	userPrompt    string
	maxEpoch      int
	history       []LLMClient.ChatMessage
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
		llmClient:     llmClient,
		searchWebTool: searchTool,
		systemPrompt:  cfg.Reactor.SystemPrompt,
		userPrompt:    cfg.Reactor.UserPrompt,
		maxEpoch:      maxEpoch,
	}

	return &agent
}

func (r *Reactor) Run(query string) ([]LLMClient.ChatMessage, error) {
	// Inject current date into prompts
	currentDate := time.Now().Format("2006-01-02")
	systemPrompt := strings.Replace(r.systemPrompt, "{current_date}", currentDate, -1)
	userPrompt := strings.Replace(r.userPrompt, "{query}", query, -1)

	// Create initial message list with system + user
	messages := []LLMClient.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	for epoch := 0; epoch < r.maxEpoch; epoch++ {
		// Call LLM and get response
		response, err := r.llmClient.InvokeMessage(messages)
		if err != nil {
			return messages, fmt.Errorf("epoch %d: invoke message error: %w", epoch, err)
		}

		// Add assistant response to messages
		messages = append(messages, LLMClient.ChatMessage{
			Role:    "assistant",
			Content: response,
		})

		// Parse the response to extract thought, action, and final answer
		_, action, finalAnswer, err := r.parseResponse(response)
		if err != nil {
			return messages, fmt.Errorf("epoch %d: parse response error: %w", epoch, err)
		}

		// If we have a final answer, return success
		if finalAnswer != "" {
			return messages, nil
		}

		// If we have an action, execute it
		if action != "" {
			toolName, param, err := r.parseAction(action)
			if err != nil {
				return messages, fmt.Errorf("epoch %d: parse action error: %w", epoch, err)
			}

			// Execute the tool
			toolResult := ""
			if toolName == "searchWeb" {
				toolResult, err = r.searchWebTool.Exec(param)
				if err != nil {
					return messages, fmt.Errorf("epoch %d: search error: %w", epoch, err)
				}
			} else {
				return messages, fmt.Errorf("epoch %d: unknown tool: %s", epoch, toolName)
			}
			// Add observation to messages for next iteration
			observationContent := fmt.Sprintf("[Tool Result]\n%s", toolResult)
			
			messages = append(messages, LLMClient.ChatMessage{
				Role:    "user",
				Content: observationContent,
			})
		} else {
			return messages, fmt.Errorf("epoch %d: no action or final answer in response", epoch)
		}
	}

	return messages, fmt.Errorf("reached max epochs (%d) without final answer", r.maxEpoch)
}

func (r *Reactor) parseResponse(response string) (thought string, action string, finalAnswer string, err error) {
	// Define regex patterns for each component
	thoughtPattern := regexp.MustCompile(`(?i)Thought:\s*(.*?)(?:\n|Action:|Final Answer:|$)`)
	actionPattern := regexp.MustCompile(`(?i)Action:\s*(.*?)(?:\n|Observation:|Final Answer:|$)`)
	finalAnswerPattern := regexp.MustCompile(`(?i)Final Answer:\s*(.*?)(?:\n|$)`)

	// Extract thought
	if thoughtMatch := thoughtPattern.FindStringSubmatch(response); len(thoughtMatch) > 1 {
		thought = strings.TrimSpace(thoughtMatch[1])
	}

	// Extract action
	if actionMatch := actionPattern.FindStringSubmatch(response); len(actionMatch) > 1 {
		action = strings.TrimSpace(actionMatch[1])
	}

	// Extract final answer
	if finalAnswerMatch := finalAnswerPattern.FindStringSubmatch(response); len(finalAnswerMatch) > 1 {
		finalAnswer = strings.TrimSpace(finalAnswerMatch[1])
	}

	return thought, action, finalAnswer, nil
}

func (r *Reactor) parseAction(action string) (name string, param string, err error) {
	// Trim whitespace
	action = strings.TrimSpace(action)

	// Pattern: functionName(parameter)
	// Example: search("query") or search("what is AI")
	actionPattern := regexp.MustCompile(`^(\w+)\s*\((.*)\)$`)

	matches := actionPattern.FindStringSubmatch(action)
	if len(matches) < 3 {
		return "", "", fmt.Errorf("invalid action format: %s", action)
	}

	name = strings.TrimSpace(matches[1])
	param = strings.TrimSpace(matches[2])

	// Remove surrounding quotes if present
	if (strings.HasPrefix(param, `"`) && strings.HasSuffix(param, `"`)) ||
		(strings.HasPrefix(param, "'") && strings.HasSuffix(param, "'")) {
		param = param[1 : len(param)-1]
	}

	return name, param, nil
}