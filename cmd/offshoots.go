package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aarose/bonsai/db"
	"github.com/spf13/cobra"
)

var offshootsCmd = &cobra.Command{
	Use:   "offshoots",
	Short: "List all child nodes of the current working node",
	Long:  `List all direct child nodes of the current working node. Shows their ID, type, and a preview of their content.`,
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

		// Get current working node
		currentNodeID, err := database.GetCurrentNode()
		if err != nil {
			fmt.Printf("Failed to get current node: %v\n", err)
			os.Exit(1)
		}

		if currentNodeID == nil {
			fmt.Println("No current working node set. Use 'bai seed' to create a root node or 'bai checkout' to move to an existing node.")
			return
		}

		// Get direct children of the current node
		children, err := database.GetDirectChildren(*currentNodeID)
		if err != nil {
			fmt.Printf("Failed to get child nodes: %v\n", err)
			os.Exit(1)
		}

		if len(children) == 0 {
			fmt.Printf("No offshoots found for current node: %s\n", *currentNodeID)
			return
		}

		fmt.Printf("Child nodes of current working node (%s):\n\n", *currentNodeID)

		for i, child := range children {
			// Show node number for easier reference
			fmt.Printf("%d. ID: %s\n", i+1, child.ID)
			fmt.Printf("   Type: %s\n", child.Type)
			if child.Model != nil {
				fmt.Printf("   Model: %s\n", *child.Model)
			}

			// Show a preview of the content (first 100 characters)
			content := child.Content
			if len(content) > 100 {
				content = content[:100] + "..."
			}
			// Replace newlines with spaces for cleaner display
			content = strings.ReplaceAll(content, "\n", " ")
			fmt.Printf("   Content: %s\n", content)

			// Add spacing between nodes except for the last one
			if i < len(children)-1 {
				fmt.Println()
			}
		}

		fmt.Printf("\nTotal: %d child node(s)\n", len(children))
	},
}

func init() {
	rootCmd.AddCommand(offshootsCmd)
}
