package agent

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	client "agent-server/client"
	"agent-server/config"
	"agent-server/memory"
	"agent-server/tool_box"
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
	llmClient     *client.Client
	searchWebTool *tool_box.SearchWeb
	systemPrompt  string
	userPrompt    string
	maxEpoch      int
	memory        *memory.ShortTermMemory
}

func NewReactor() *Reactor {
	cfg := config.GetConfig()

	llmClient := client.NewClient()
	searchTool := tool_box.NewSearchWeb()

	// Validate configuration
	maxEpoch := cfg.Reactor.MaxEpoch
	if maxEpoch <= 0 {
		maxEpoch = 50
	}

	reactor := Reactor{
		llmClient:     llmClient,
		searchWebTool: searchTool,
		systemPrompt:  cfg.Reactor.SystemPrompt,
		userPrompt:    cfg.Reactor.UserPrompt,
		maxEpoch:      maxEpoch,
		memory:        memory.NewShortTermMemory(),
	}

	return &reactor
}

func (r *Reactor) Run(query string) ([]client.ChatMessage, error) {
	// Inject current date into prompts
	currentDate := time.Now().Format("2006-01-02")
	systemPrompt := strings.Replace(r.systemPrompt, "{current_date}", currentDate, -1)
	userPrompt := strings.Replace(r.userPrompt, "{query}", query, -1)

	// 基于原始 query 召回相关历史记忆 (只召回一次)
	recallTopK := config.GetConfig().Memory.RecallTopK
	recalledItems, err := r.memory.Recall(query, recallTopK)
	if err != nil {
		return nil, fmt.Errorf("recall memory error: %w", err)
	}

	messages := make([]client.ChatMessage, 0)

	// 历史记忆作为独立的 user message，注意不要重复加入 memory
	memoryContext := r.buildMemoryContext(recalledItems)
	if memoryContext != "" {
		messages = append(messages, client.ChatMessage{
			Role:    "user",
			Content: "## 相关历史记忆:\n" + memoryContext,
		})
	}
	systemMsg := client.ChatMessage{
		Role:    "system",
		Content: systemPrompt,
	}
	userMsg := client.ChatMessage{
		Role:    "user",
		Content: userPrompt,
	}
	messages = append(messages, systemMsg, userMsg)
	r.memory.Add(userMsg, 1.0) // 系统提示词不计入记忆，每轮对话都会自动构建

	for epoch := 0; epoch < r.maxEpoch; epoch++ {
		// Call LLM and get response
		response, err := r.llmClient.InvokeMessage(messages)
		if err != nil {
			return messages, fmt.Errorf("epoch %d: invoke message error: %w", epoch, err)
		}

		// Add assistant response to messages
		messages = append(messages, client.ChatMessage{
			Role:    "assistant",
			Content: response,
		})

		// Parse JSON response
		resp, err := r.parseResponse(response)
		if err != nil {
			return messages, fmt.Errorf("epoch %d: parse response error: %w", epoch, err)
		}

		// If we have a final answer, save to memory and return
		if resp.FinalAnswer != "" {
			// 保存最终答案到记忆
			importance := float32(resp.MemoryScore)
			err = r.memory.Add(client.ChatMessage{Role: "assistant", Content: resp.FinalAnswer}, importance)
			if err != nil {
				fmt.Printf("Warning: failed to add final answer to memory: %v\n", err)
			}
			messages = append(messages, client.ChatMessage{
				Role:    "assistant",
				Content: resp.FinalAnswer,
			})
			return messages, nil
		}

		// 记录本轮对话的 LLM 回复到记忆 (importance = resp.MemoryScore)
		err = r.memory.Add(client.ChatMessage{Role: "assistant", Content: response}, float32(resp.MemoryScore))
		if err != nil {
			fmt.Printf("Warning: failed to add LLM response to memory: %v\n", err)
		}
		// If we have an action, execute it
		if resp.Action != nil {
			toolResult, err := r.executeAction(resp.Action)
			if err != nil {
				return messages, fmt.Errorf("epoch %d: execute action error: %w", epoch, err)
			}

			// 工具结果使用 LLM 返回的 memory_score
			importance := float32(resp.MemoryScore)
			messages = append(messages, client.ChatMessage{
				Role:    "tool",
				Content: toolResult,
			})
			err = r.memory.Add(client.ChatMessage{Role: "tool", Content: toolResult}, importance)
			if err != nil {
				fmt.Printf("Warning: failed to add tool result to memory: %v\n", err)
			}
		} else {
			return messages, fmt.Errorf("epoch %d: no action or final answer in response", epoch)
		}
	}

	return messages, fmt.Errorf("reached max epochs (%d) without final answer", r.maxEpoch)
}

// buildMemoryContext 构建记忆上下文字符串
func (r *Reactor) buildMemoryContext(items []memory.MemoryItem) string {
	if len(items) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("- [%s] %s\n", item.Role, item.Content))
	}
	return sb.String()
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
