package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aarose/bonsai/db"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bai [message]",
	Short: "Bonsai is a CLI tool for managing LLM conversation trees",
	Long: `Bonsai is a CLI tool for managing LLM conversation trees.`,
	Args:                  cobra.ArbitraryArgs,
	DisableFlagParsing:    false,
	SilenceUsage:         true,
	Run: func(cmd *cobra.Command, args []string) {
		// If no arguments provided, show help message
		if len(args) == 0 {
			// Initialize database
			if err := initializeDatabase(); err != nil {
				log.Fatalf("Failed to initialize database: %v", err)
			}
			fmt.Println("Hello from bai! Use --help to see available commands.")
			return
		}

		// Handle message input - create a child node
		message := args[0]

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

		// Get current working node
		currentNodeID, err := database.GetCurrentNode()
		if err != nil {
			fmt.Printf("Failed to get current node: %v\n", err)
			os.Exit(1)
		}

		if currentNodeID == nil {
			fmt.Println("No current working node set. Use 'bai seed \"message\"' to create a root node first.")
			return
		}

		// Get the current working node to inherit model
		currentNode, err := database.GetNodeByID(*currentNodeID)
		if err != nil {
			fmt.Printf("Failed to get current node details: %v\n", err)
			os.Exit(1)
		}

		// Determine which model to use
		var model *string
		if llm != "" {
			model = &llm // Use flag if provided
		} else {
			model = currentNode.Model // Inherit from parent
		}

		// Create child node
		node, err := database.CreateChildNode(message, *currentNodeID, model)
		if err != nil {
			fmt.Printf("Failed to create child node: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Created child node with ID: %s\n", node.ID)
		fmt.Printf("Parent: %s\n", *currentNodeID)
		if node.Model != nil {
			fmt.Printf("Model: %s\n", *node.Model)
		}
		fmt.Printf("Message: %s\n", node.Content)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func initializeDatabase() error {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create database path in user's home directory
	dbPath := filepath.Join(homeDir, ".bonsai", "bonsai.db")

	// Create and initialize database
	database, err := db.NewDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	defer database.Close()

	if err := database.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	fmt.Printf("Database created at: %s\n", dbPath)
	return nil
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringP("llm", "l", "", "LLM model to use for the conversation (e.g., gpt-4, claude-3-sonnet, gpt-3.5-turbo)")
}
