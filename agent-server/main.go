package main

import (
	"fmt"
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

	fmt.Println("=== NanoAgent Reactor Test ===")

	// Create reactor instance
	reactor := agent.NewReactor()
	fmt.Println("Reactor created successfully")

	// Test queries
	testQueries := []string{
		// "What is the capital of France?",
		// "What is the next main version of iphone",
		// "原神下个版本卡池值得抽取么，我没抽兹白，但是有少女？",
		// "接下来还有多久才放假啊",
		// "长江电力未来三年的净利润预测",
		// "我比较喜欢木偶，目前有少女，下个版本的新角色好像也是岩系的，但没抽6.3的兹白...下个版本应该抽这个新角色么，它在未来能适配木偶的队伍么？还是说抽了就只能等队友兹白复刻了",
		"你好呀",
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
				fmt.Printf("\n[%d] %s (length: %d):\n%s...\n", idx, msg.Role, contentLen, msg.Content)
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