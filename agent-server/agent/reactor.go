package agnet

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	Config "agent-server/config"
	llm "agent-server/client"
	ToolBox "agent-server/toolBox"
)

// LLM JSON 响应结构
type LLMResponse struct {
	Thought     string      `json:"thought"`
	Action      *ActionCall `json:"action"`
	FinalAnswer string      `json:"final_answer"`
	MemoryScore float64     `json:"memory_score"`
}

// 工具调用结构
type ActionCall struct {
	Tool       string                 `json:"tool"`
	Parameters map[string]interface{} `json:"parameters"`
}

type Reactor struct {
	llmClient     *llm.Client
	searchWebTool *ToolBox.SearchWeb
	systemPrompt  string
	userPrompt    string
	maxEpoch      int
	history       []llm.ChatMessage
}

func NewReactor() *Reactor {
	cfg := Config.GetConfig()

	llmClient := llm.NewClient()
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

func (r *Reactor) Run(query string) ([]llm.ChatMessage, error) {
	// Inject current date into prompts
	currentDate := time.Now().Format("2006-01-02")
	systemPrompt := strings.Replace(r.systemPrompt, "{current_date}", currentDate, -1)
	userPrompt := strings.Replace(r.userPrompt, "{query}", query, -1)

	// Create initial message list with system + user
	messages := []llm.ChatMessage{
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
		messages = append(messages, llm.ChatMessage{
			Role:    "assistant",
			Content: response,
		})

		// Parse JSON response
		resp, err := r.parseResponse(response)
		if err != nil {
			return messages, fmt.Errorf("epoch %d: parse response error: %w", epoch, err)
		}

		// If we have a final answer, return success
		if resp.FinalAnswer != "" {
			return messages, nil
		}

		// If we have an action, execute it
		if resp.Action != nil {
			toolResult, err := r.executeAction(resp.Action)
			if err != nil {
				return messages, fmt.Errorf("epoch %d: execute action error: %w", epoch, err)
			}

			messages = append(messages, llm.ChatMessage{
				Role:    "tool",
				Content: toolResult,
			})
		} else {
			return messages, fmt.Errorf("epoch %d: no action or final answer in response", epoch)
		}
	}

	return messages, fmt.Errorf("reached max epochs (%d) without final answer", r.maxEpoch)
}

// parseResponse 解析 LLM 的 JSON 响应
func (r *Reactor) parseResponse(response string) (*LLMResponse, error) {
	response = strings.TrimSpace(response)

	// 1. 尝试直接解析整个响应
	var resp LLMResponse
	if err := json.Unmarshal([]byte(response), &resp); err == nil {
		return &resp, nil
	}

	// 2. 尝试提取 ```json ... ``` 代码块
	jsonBlock := extractJSONBlock(response)
	if jsonBlock != "" {
		// 预处理：替换中文引号
		jsonBlock = sanitizeJSON(jsonBlock)
		if err := json.Unmarshal([]byte(jsonBlock), &resp); err == nil {
			return &resp, nil
		}
	}

	// 3. 兜底：提取第一个完整的 JSON 对象
	jsonObj := extractFirstJSONObject(response)
	if jsonObj != "" {
		jsonObj = sanitizeJSON(jsonObj)
		if err := json.Unmarshal([]byte(jsonObj), &resp); err == nil {
			return &resp, nil
		}
	}

	// 4. 全部失败，返回错误
	return nil, fmt.Errorf("failed to parse JSON response: %s", response)
}

// sanitizeJSON 预处理 JSON 字符串，修复常见问题
func sanitizeJSON(jsonStr string) string {
	// 替换中文引号为英文引号（需要转义）
	jsonStr = strings.ReplaceAll(jsonStr, "\u201c", "\\\"") // "
	jsonStr = strings.ReplaceAll(jsonStr, "\u201d", "\\\"") // "
	jsonStr = strings.ReplaceAll(jsonStr, "\u2018", "\\\"") // '
	jsonStr = strings.ReplaceAll(jsonStr, "\u2019", "\\\"") // '
	return jsonStr
}

// extractFirstJSONObject 提取第一个完整的 JSON 对象
func extractFirstJSONObject(response string) string {
	start := strings.Index(response, "{")
	if start == -1 {
		return ""
	}

	depth := 0
	inString := false
	for i := start; i < len(response); i++ {
		c := response[i]
		if c == '\\' && i+1 < len(response) {
			i++ // 跳过转义字符
			continue
		}
		if c == '"' {
			inString = !inString
		} else if !inString {
			if c == '{' {
				depth++
			} else if c == '}' {
				depth--
				if depth == 0 {
					return response[start : i+1]
				}
			}
		}
	}

	return ""
}

// extractJSONBlock 提取 ```json 代码块
func extractJSONBlock(response string) string {
	// 匹配 ```json ... ``` (允许多个换行)
	re := regexp.MustCompile("(?s)```json\\s*(.*?)\\s*```")
	if match := re.FindStringSubmatch(response); len(match) > 1 {
		return strings.TrimSpace(match[1])
	}

	// 匹配 ``` ... ``` (可能没写 json 标识)
	re = regexp.MustCompile("(?s)```\\s*(.*?)\\s*```")
	if match := re.FindStringSubmatch(response); len(match) > 1 {
		content := strings.TrimSpace(match[1])
		// 确保是 JSON 格式
		if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
			return content
		}
	}

	return ""
}

// executeAction 执行工具调用
func (r *Reactor) executeAction(action *ActionCall) (string, error) {
	switch action.Tool {
	case "searchWeb":
		// 兼容 "query" 和 "param" 两种参数名
		var query string
		if p, ok := action.Parameters["param"].(string); ok {
			query = p
		} else {
			return "", fmt.Errorf("invalid query parameter for searchWeb")
		}
		return r.searchWebTool.Exec(query)
	default:
		return "", fmt.Errorf("unknown tool: %s", action.Tool)
	}
}