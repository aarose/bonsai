package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aarose/bonsai/db"
	"github.com/spf13/cobra"
)

var cherryPickCmd = &cobra.Command{
	Use:   "cherry-pick <node-id>",
	Short: "Duplicate a node's content as a child of the current working node (graft a tree bud)",
	Long:  `Duplicates the content of the specified node and creates it as a child of the current working node with a new ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sourceNodeID := args[0]

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

		// Get the source node to cherry-pick
		sourceNode, err := database.GetNodeByID(sourceNodeID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Check if we're trying to cherry-pick the current working node
		if sourceNodeID == *currentNodeID {
			fmt.Printf("Cannot cherry-pick the current working node onto itself.\n")
			os.Exit(1)
		}

		// Create a duplicate of the source node as a child of current working node
		// We preserve the source node's type and model
		duplicateNode, err := database.CreateChildNodeWithType(sourceNode.Content, *currentNodeID, sourceNode.Type, sourceNode.Model)
		if err != nil {
			fmt.Printf("Failed to create cherry-picked node: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("🍒 Cherry-picked node: \033[33m%s\033[0m\n", sourceNodeID)
		fmt.Printf("Created new node with ID: \033[33m%s\033[0m\n", duplicateNode.ID)
		fmt.Printf("Type: %s\n", duplicateNode.Type)

		fmt.Printf("Parent: \033[33m%s\033[0m\n", *currentNodeID)
		if duplicateNode.Model != nil {
			fmt.Printf("Model: %s\n", *duplicateNode.Model)
		}

		// Show a preview of the content
		content := duplicateNode.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		content = strings.ReplaceAll(content, "\n", " ")
		fmt.Printf("Content: %s\n", content)

		fmt.Printf("\nSuccessfully cherry-picked content from \033[33m%s\033[0m to new node \033[33m%s\033[0m\n", sourceNodeID, duplicateNode.ID)
	},
}

func init() {
	rootCmd.AddCommand(cherryPickCmd)
}
