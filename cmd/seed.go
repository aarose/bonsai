package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aarose/bonsai/db"
	"github.com/aarose/bonsai/pkg/config"
	"github.com/aarose/bonsai/pkg/llm"
	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "seed <content>",
	Short: "Create a new root node with the given content",
	Long:  `Create a new root node (no parent) with the provided content. The node type will be set to "user".`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		content := args[0]

		// Get LLM flag value
		llmModel, err := cmd.Flags().GetString("llm")
		if err != nil {
			fmt.Printf("Failed to get llm flag: %v\n", err)
			os.Exit(1)
		}

		// Get user's home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Failed to get home directory: %v\n", err)
			os.Exit(1)
		}

		// Create database path in user's home directory
		dbPath := filepath.Join(homeDir, ".bonsai", "bonsai.db")

		// Create and initialize database
		database, err := db.NewDatabase(dbPath)
		if err != nil {
			fmt.Printf("Failed to create database: %v\n", err)
			os.Exit(1)
		}
		defer database.Close()

		if err := database.Initialize(); err != nil {
			fmt.Printf("Failed to initialize database: %v\n", err)
			os.Exit(1)
		}

		// Create root node with model
		var model *string
		if llmModel != "" {
			model = &llmModel
		}
		node, err := database.CreateRootNode(content, model)
		if err != nil {
			fmt.Printf("Failed to create root node: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("🌱 \033[32mCreated seed node with ID:\033[0m \033[33m%s\033[0m\n", node.ID)
		if node.Model != nil {
			fmt.Printf("🧠 Model: \033[35m%s\033[0m\n", *node.Model)
		}
		fmt.Printf("💬 Message: \033[90m%s\033[0m\n", node.Content)

		// Generate LLM response if model is specified
		if llmModel != "" {
			fmt.Printf("Generating LLM response...\n")

			// Get API key from environment or config
			apiKey := config.GetAPIKey(llmModel)
			if apiKey == "" {
				fmt.Printf("Warning: No API key found for %s. Set %s environment variable.\n", llmModel, config.GetAPIKeyEnvVar(llmModel))
			} else {
				// Create LLM client
				llmConfig := llm.Config{
					APIKey:    apiKey,
					MaxTokens: 1000, // Reasonable default
				}

				client, err := llm.NewClient(llmModel, llmConfig)
				if err != nil {
					fmt.Printf("Warning: Failed to create LLM client: %v\n", err)
				} else {
					// Generate response with timeout
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()

					response, err := client.GenerateResponse(ctx, content, llmModel)
					if err != nil {
						fmt.Printf("Warning: Failed to get LLM response: %v\n", err)
					} else {
						// Create child node with LLM response
						llmNode, err := database.CreateLLMResponseNode(node.ID, response, llmModel)
						if err != nil {
							fmt.Printf("Warning: Failed to create LLM response node: %v\n", err)
						} else {
							fmt.Printf("Created LLM response node with ID: \033[33m%s\033[0m\n", llmNode.ID)
							fmt.Printf("🤖 LLM Response: %s\n", llmNode.Content)
						}
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(seedCmd)
	seedCmd.Flags().StringP("llm", "l", "", "LLM model to use for the conversation (e.g., gpt-4, claude-3-sonnet, gpt-3.5-turbo)")
}
