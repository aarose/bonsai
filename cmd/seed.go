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
	Short: "Create a conversation starting with this message",
	Long:  `Create a conversation starting with this message and optional LLM model`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		content := args[0]

		// Get LLM flag value
		llm, err := cmd.Flags().GetString("llm")
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to get llm flag: %v\033[0m\n", err)
			os.Exit(1)
		}

		// Get user's home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to get home directory: %v\033[0m\n", err)
			os.Exit(1)
		}

		// Create database path in user's home directory
		dbPath := filepath.Join(homeDir, ".bonsai", "bonsai.db")

		// Create and initialize database
		database, err := db.NewDatabase(dbPath)
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to create database: %v\033[0m\n", err)
			os.Exit(1)
		}
		defer database.Close()

		if err := database.Initialize(); err != nil {
			fmt.Printf("\033[31m❌ Failed to initialize database: %v\033[0m\n", err)
			os.Exit(1)
		}

		// Create root node with model
		var model *string
		if llm != "" {
			model = &llm
		}
		node, err := database.CreateRootNode(content, model)
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to create seed node: %v\033[0m\n", err)
			os.Exit(1)
		}

		fmt.Printf("🌱 \033[32mCreated seed node with ID:\033[0m \033[33m%s\033[0m\n", node.ID)
		if node.Model != nil {
			fmt.Printf("🧠 Model: \033[35m%s\033[0m\n", *node.Model)
		}
		fmt.Printf("💬 Message: \033[90m%s\033[0m\n", node.Content)
	},
}

func init() {
	rootCmd.AddCommand(seedCmd)
	seedCmd.Flags().StringP("llm", "l", "", "LLM model to use for the conversation (e.g., gpt-4, claude-3-sonnet, gpt-3.5-turbo)")
}
