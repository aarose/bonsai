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
	Short: "Cut off a branch of the conversation tree, deleting it",
	Long:  `Cut off a branch of the conversation tree, deleting it. This action cannot be undone.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nodeID := args[0]

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

		// Get the node to be deleted and preview what will be affected
		nodesToDelete, err := database.GetNodeAndAllChildren(nodeID)
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to get node and children: %v\033[0m\n", err)
			os.Exit(1)
		}

		if len(nodesToDelete) == 0 {
			fmt.Printf("\033[31m❌ Node with ID '%s' not found.\033[0m\n", nodeID)
			os.Exit(1)
		}

		// Show what will be deleted
			fmt.Printf("🪚 \033[33mThis will delete the following %d node(s):\033[0m\n\n", len(nodesToDelete))
		for i, node := range nodesToDelete {
			indent := ""
			if i > 0 { // Child nodes get indented
				indent = "  └─ "
			} else { // Root node being deleted
				indent = "• "
			}
			var typeIcon string
			if node.Type == "user" {
				typeIcon = "👤"
			} else {
				typeIcon = "🤖"
			}
			fmt.Printf("%s%s \033[33m%s\033[0m: \033[90m%s\033[0m\n", indent, typeIcon, node.ID, truncateContent(node.Content, 50))
			if node.Model != nil {
				fmt.Printf("%s🧠 Model: \033[35m%s\033[0m\n", strings.Repeat(" ", len(indent)), *node.Model)
			}
		}

		// Check if current node will be affected
		currentNodeID, err := database.GetCurrentNode()
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to get current node: %v\033[0m\n", err)
			os.Exit(1)
		}

		willDeleteCurrent := false
		if currentNodeID != nil {
			for _, node := range nodesToDelete {
				if *currentNodeID == node.ID {
					willDeleteCurrent = true
					fmt.Printf("\n\033[33m⚠️  WARNING: This will delete your current working node!\033[0m\n")
					break
				}
			}
		}

		// Ask for confirmation
		fmt.Printf("\n\033[33mAre you sure you want to prune these nodes? This cannot be undone.\033[0m \033[1m(y/N):\033[0m ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to read input: %v\033[0m\n", err)
			os.Exit(1)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("\033[90mPruning cancelled.\033[0m")
			return
		}

		// Perform the deletion
		deletedCount, err := database.DeleteNodeAndAllChildren(nodeID)
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to delete nodes: %v\033[0m\n", err)
			os.Exit(1)
		}

		// Clear current node if it was deleted
		if willDeleteCurrent {
			if err := database.ClearCurrentNode(); err != nil {
				fmt.Printf("\033[33m⚠️  Deleted nodes but failed to clear current node: %v\033[0m\n", err)
			} else {
				fmt.Printf("\033[90mCurrent working node has been cleared.\033[0m\n")
			}
		}

		fmt.Printf("\033[32m✅ Successfully pruned %d node(s) from the Bonsai tree.\033[0m\n", deletedCount)
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
