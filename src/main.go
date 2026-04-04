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
		"What is the capital of France?",
		"What is the newest version of iphone",
	}

	for i, query := range testQueries {
		fmt.Printf("\n--- Test Query %d ---\n", i+1)
		fmt.Printf("Query: %s\n", query)

		// Run the reactor
		result, err := reactor.Run(query)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("Result: %s\n", result)
		fmt.Println("--- End Test ---\n")
	}

	fmt.Println("=== All tests completed ===")
}
