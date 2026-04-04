package main

import (
	"fmt"
	Agnet "nanoagent/src/agnet"
	Config "nanoagent/src/config"
)

func main() {
	// Load configuration
	_, err := Config.LoadConfig("src/resource/config.yaml")
	if err != nil {
		fmt.Println("load config error:", err)
		return
	}

	fmt.Println("=== NanoAgent Reactor Test ===")

	// Create reactor instance
	reactor := Agnet.NewReactor()
	fmt.Println("Reactor created successfully")

	// Test queries
	testQueries := []string{
		// "What is the capital of France?",
		// "What is the newest version of iphone",
		"原神卡池新角色值得抽吗？",
	}

	for i, query := range testQueries {
		fmt.Printf("\n========== Test Query %d ==========\n", i+1)
		fmt.Printf("Query: %s\n", query)

		// Run the reactor
		messages, err := reactor.Run(query)

		// Format and print the messages
		fmt.Println("\n--- Execution History ---")
		for idx, msg := range messages {
			contentLen := len(msg.Content)
			if contentLen > 200 {
				fmt.Printf("\n[%d] %s (length: %d):\n%s...\n", idx, msg.Role, contentLen, msg.Content[:200])
			} else {
				fmt.Printf("\n[%d] %s:\n%s\n", idx, msg.Role, msg.Content)
			}
		}

		// Print final error or success status
		if err != nil {
			fmt.Printf("\n[Status] Error: %v\n", err)
		} else {
			fmt.Println("\n[Status] Success")
		}

		fmt.Println("\n========== End Test ==========")
	}

	fmt.Println("\n=== All tests completed ===")
}
