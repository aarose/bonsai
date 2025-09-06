package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aarose/bonsai/db"
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
		llm, err := cmd.Flags().GetString("llm")
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
		if llm != "" {
			model = &llm
		}
		node, err := database.CreateRootNode(content, model)
		if err != nil {
			fmt.Printf("Failed to create root node: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Created root node with ID: %s\n", node.ID)
		fmt.Printf("Content: %s\n", node.Content)
		fmt.Printf("Type: %s\n", node.Type)
		if node.Model != nil {
			fmt.Printf("Model: %s\n", *node.Model)
		}
	},
}

func init() {
	rootCmd.AddCommand(seedCmd)
	seedCmd.Flags().StringP("llm", "l", "", "LLM model to use for the conversation (e.g., gpt-4, claude-3-sonnet, gpt-3.5-turbo)")
}
