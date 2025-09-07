package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aarose/bonsai/db"
	"github.com/aarose/bonsai/pkg/config"
	"github.com/aarose/bonsai/pkg/llm"
	"github.com/spf13/cobra"
)

const longDescription = `🌳 Bonsai is a CLI tool for managing LLM conversation trees.

It provides a structured way to explore multiple directions of a conversation,
experiment freely, and return to earlier points without friction.

Think of it as Git for your LLM sessions — a system that lets you branch, reseed,
and graft ideas without losing track of the bigger picture.

With Bonsai, you can:
• Maintain clean histories
• Compare alternate paths
• Keep your conversations organized as they grow

The result is a workflow where creativity and control coexist, making it easy to
cultivate ideas, revisit roots, and guide your conversations toward meaningful
outcomes.`

var rootCmd = &cobra.Command{
	Use:                "bai [message]",
	Short:              "🌳 Bonsai is a CLI tool for managing LLM conversation trees",
	Long:               longDescription,
	Args:               cobra.ArbitraryArgs,
	DisableFlagParsing: false,
	SilenceUsage:       true,
	Run: func(cmd *cobra.Command, args []string) {
		// If no arguments provided, show help message
		if len(args) == 0 {
			// Initialize database (close after init since we just want to ensure it exists)
			if _, err := initializeDatabase(true); err != nil {
				log.Fatalf("Failed to initialize database: %v", err)
			}
			fmt.Println("🌳 Hello from bai! Use --help to see available commands.")
			return
		}

		// Handle message input - create a child node
		message := args[0]

		// Get LLM flag value
		llmModel, err := cmd.Flags().GetString("llm")
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to get llm flag: %v\033[0m\n", err)
			os.Exit(1)
		}

		// Initialize database
		database, err := initializeDatabase(false)
		if err != nil {
			fmt.Printf("\033[31m❌ %v\033[0m\n", err)
			os.Exit(1)
		}
		defer database.Close()

		// Get current working node
		currentNodeID, err := database.GetCurrentNode()
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to get current node: %v\033[0m\n", err)
			os.Exit(1)
		}

		if currentNodeID == nil {
			fmt.Println("🌱 No current working node set. Use 'bai seed \"message\"' to create a root node first.")
			return
		}

		// Get the current working node to inherit model
		currentNode, err := database.GetNodeByID(*currentNodeID)
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to get current node details: %v\033[0m\n", err)
			os.Exit(1)
		}

		// Determine which model to use
		var model *string
		if llmModel != "" {
			model = &llmModel // Use flag if provided
		} else {
			model = currentNode.Model // Inherit from parent
		}

		// Create child node
		node, err := database.CreateChildNode(message, *currentNodeID, model)
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to create child node: %v\033[0m\n", err)
			os.Exit(1)
		}

		fmt.Printf("🔄 \033[32mCreated child node with ID:\033[0m \033[33m%s\033[0m\n", node.ID)
		fmt.Printf("⬆️  Parent: \033[33m%s\033[0m\n", *currentNodeID)
		if node.Model != nil {
			fmt.Printf("🧠 Model: \033[35m%s\033[0m\n", *node.Model)
		}
		fmt.Printf("💬 Message: \033[90m%s\033[0m\n", node.Content)

		// Generate LLM response if model is available
		if model != nil && *model != "" {
			fmt.Printf("Generating LLM response...\n")

			// Get API key from environment or config
			apiKey := config.GetAPIKey(*model)
			if apiKey == "" {
				fmt.Printf("Warning: No API key found for %s. Set %s environment variable.\n", *model, config.GetAPIKeyEnvVar(*model))
			} else {
				// Create LLM client
				llmConfig := llm.Config{
					APIKey:    apiKey,
					MaxTokens: 1000, // Reasonable default
				}

				client, err := llm.NewClient(*model, llmConfig)
				if err != nil {
					fmt.Printf("Warning: Failed to create LLM client: %v\n", err)
				} else {
					// Generate response with timeout
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()

					response, err := client.GenerateResponse(ctx, message, *model)
					if err != nil {
						fmt.Printf("Warning: Failed to get LLM response: %v\n", err)
					} else {
						// Create child node with LLM response
						llmNode, err := database.CreateLLMResponseNode(node.ID, response, *model)
						if err != nil {
							fmt.Printf("Warning: Failed to create LLM response node: %v\n", err)
						} else {
							fmt.Printf("Created LLM response node with ID: %s\n", llmNode.ID)
							fmt.Printf("LLM Response: %s\n", llmNode.Content)
						}
					}
				}
			}
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

// initializeDatabase creates and initializes the database, returning the connection
// If closeAfterInit is true, closes the connection and returns nil database
func initializeDatabase(closeAfterInit bool) (*db.Database, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create database path in user's home directory
	dbPath := filepath.Join(homeDir, ".bonsai", "bonsai.db")

	// Create and initialize database
	database, err := db.NewDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	if err := database.Initialize(); err != nil {
		database.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	if closeAfterInit {
		fmt.Printf("📊 Database created at: \033[90m%s\033[0m\n", dbPath)
		database.Close()
		return nil, nil
	}

	return database, nil
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringP("llm", "l", "", "LLM model to use for the conversation (e.g., gpt-4, claude-3-sonnet, gpt-3.5-turbo)")
}
