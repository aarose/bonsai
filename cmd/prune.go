package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aarose/bonsai/db"
	"github.com/spf13/cobra"
)

var pruneCmd = &cobra.Command{
	Use:   "prune <node-id>",
	Short: "Delete a node and all of its children",
	Long:  `Delete the specified node and all of its children recursively. This action cannot be undone.`,
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

		// Get the node to be deleted and preview what will be affected
		nodesToDelete, err := database.GetNodeAndAllChildren(nodeID)
		if err != nil {
			fmt.Printf("Failed to get node and children: %v\n", err)
			os.Exit(1)
		}

		if len(nodesToDelete) == 0 {
			fmt.Printf("Node with ID '%s' not found.\n", nodeID)
			os.Exit(1)
		}

		// Show what will be deleted
		fmt.Printf("🗑️  This will delete the following %d node(s):\n\n", len(nodesToDelete))
		for i, node := range nodesToDelete {
			indent := ""
			if i > 0 { // Child nodes get indented
				indent = "  └─ "
			} else { // Root node being deleted
				indent = "• "
			}
			fmt.Printf("%s%s: %s\n", indent, node.ID, truncateContent(node.Content, 50))
			if node.Model != nil {
				fmt.Printf("%s   Model: %s\n", strings.Repeat(" ", len(indent)), *node.Model)
			}
		}

		// Check if current node will be affected
		currentNodeID, err := database.GetCurrentNode()
		if err != nil {
			fmt.Printf("Failed to get current node: %v\n", err)
			os.Exit(1)
		}

		willDeleteCurrent := false
		if currentNodeID != nil {
			for _, node := range nodesToDelete {
				if *currentNodeID == node.ID {
					willDeleteCurrent = true
					fmt.Printf("\n⚠️  WARNING: This will delete your current working node!\n")
					break
				}
			}
		}

		// Ask for confirmation
		fmt.Printf("\nAre you sure you want to prune these nodes? This cannot be undone. (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Failed to read input: %v\n", err)
			os.Exit(1)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Pruning cancelled.")
			return
		}

		// Perform the deletion
		deletedCount, err := database.DeleteNodeAndAllChildren(nodeID)
		if err != nil {
			fmt.Printf("Failed to delete nodes: %v\n", err)
			os.Exit(1)
		}

		// Clear current node if it was deleted
		if willDeleteCurrent {
			if err := database.ClearCurrentNode(); err != nil {
				fmt.Printf("Deleted nodes but failed to clear current node: %v\n", err)
			} else {
				fmt.Printf("Current working node has been cleared.\n")
			}
		}

		fmt.Printf("✅ Successfully pruned %d node(s) from the Bonsai tree.\n", deletedCount)
	},
}

// truncateContent truncates content to a specified length with ellipsis
func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen-3] + "..."
}

func init() {
	rootCmd.AddCommand(pruneCmd)
}
