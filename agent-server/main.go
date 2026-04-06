package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"agent-server/agent"
	"agent-server/config"
)

func main() {
	// Load configuration
	_, err := config.LoadConfig("resource/config.yaml")
	if err != nil {
		fmt.Println("load config error:", err)
		return
	}

	fmt.Println("=== NanoAgent Interactive Mode ===")
	fmt.Println("Type your question and press Enter to chat.")
	fmt.Println("Type 'exit' or 'quit' to stop.")
	fmt.Println("Type 'clear' to clear memory.")
	fmt.Println("==================================")

	// Create reactor instance (memory persists across conversations)
	reactor := agent.NewReactor()

	scanner := bufio.NewScanner(os.Stdin)
	round := 0

	for {
		round++
		fmt.Printf("\n[%d] You: ", round)

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			round-- // don't count empty input
			continue
		}

		// Check for exit commands
		if input == "exit" || input == "quit" {
			fmt.Println("\nGoodbye!")
			break
		}

		// Check for clear memory command
		if input == "clear" {
			reactor = agent.NewReactor()
			fmt.Println("Memory cleared. Starting fresh conversation.")
			round = 0
			continue
		}

		// Run the reactor
		fmt.Println("\n--- Thinking... ---")
		messages, err := reactor.Run(input)

		// Print the final response
		fmt.Println("\n--- Response ---")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		// Print all messages for debugging
		for i, msg := range messages {
			content := msg.Content
			fmt.Printf("\n[%d] %s:\n%s\n", i, msg.Role, content)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
	}
}