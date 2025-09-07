package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aarose/bonsai/db"
	"github.com/spf13/cobra"
)

var checkoutCmd = &cobra.Command{
	Use:   "checkout <node-id>",
	Short: "Move the current working node to the specified node",
	Long:  `Move the current working node to the node with the given ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nodeID := args[0]

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

		// Check if the node exists
		node, err := database.GetNodeByID(nodeID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Get current working node for comparison
		currentNodeID, err := database.GetCurrentNode()
		if err != nil {
			fmt.Printf("Failed to get current node: %v\n", err)
			os.Exit(1)
		}

		// Check if we're already on this node
		if currentNodeID != nil && *currentNodeID == nodeID {
			fmt.Printf("Already on node %s\n", nodeID)
			return
		}

		// Set the new current node
		if err := database.SetCurrentNode(nodeID); err != nil {
			fmt.Printf("Failed to set current node: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Moved to node: %s\n", nodeID)
		fmt.Printf("Type: %s\n", node.Type)
		if node.Model != nil {
			fmt.Printf("Model: %s\n", *node.Model)
		}
		if node.Parent != nil {
			fmt.Printf("Parent: %s\n", *node.Parent)
		}
		fmt.Printf("Content: %s\n", node.Content)
	},
}

func init() {
	rootCmd.AddCommand(checkoutCmd)
}
