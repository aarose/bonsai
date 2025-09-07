package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aarose/bonsai/db"
	"github.com/spf13/cobra"
)

var seedsCmd = &cobra.Command{
	Use:   "seeds",
	Short: "List all root nodes",
	Long:  `List all root nodes (nodes without a parent) in the database.`,
	Run: func(cmd *cobra.Command, args []string) {
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

		// Get all root nodes
		rootNodes, err := database.GetRootNodes()
		if err != nil {
			fmt.Printf("Failed to get root nodes: %v\n", err)
			os.Exit(1)
		}

		// Get current working node
		currentNodeID, err := database.GetCurrentNode()
		if err != nil {
			fmt.Printf("Failed to get current node: %v\n", err)
			os.Exit(1)
		}

		if len(rootNodes) == 0 {
			fmt.Println("No root nodes found.")
			return
		}

		fmt.Printf("🌱 Found %d seed(s) in the Bonsai garden:\n\n", len(rootNodes))
		for _, node := range rootNodes {
			if currentNodeID != nil && *currentNodeID == node.ID {
				fmt.Printf("ID: \033[33m%s\033[0m \033[32m(current working node)\033[0m\n", node.ID)
			} else {
				fmt.Printf("ID: \033[33m%s\033[0m\n", node.ID)
			}
			if node.Model != nil {
				fmt.Printf("Model: %s\n", *node.Model)
			}
			fmt.Printf("Message: %s\n", node.Content)
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(seedsCmd)
}
