package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var checkoutCmd = &cobra.Command{
	Use:   "checkout <node-id>",
	Short: "Jump to a different node in the conversation tree",
	Long:  `Jumps to a different node in the conversation tree.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nodeID := args[0]

		// Initialize database
		database, err := initializeDatabase(false)
		if err != nil {
			fmt.Printf("\033[31m❌ %v\033[0m\n", err)
			os.Exit(1)
		}
		defer database.Close()

		// Check if the node exists
		node, err := database.GetNodeByID(nodeID)
		if err != nil {
			fmt.Printf("\033[31m❌ Error: %v\033[0m\n", err)
			os.Exit(1)
		}

		// Get current working node for comparison
		currentNodeID, err := database.GetCurrentNode()
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to get current node: %v\033[0m\n", err)
			os.Exit(1)
		}

		// Check if we're already on this node
		if currentNodeID != nil && *currentNodeID == nodeID {
			fmt.Printf("📍 Already on node \033[33m%s\033[0m\n", nodeID)
			return
		}

		// Set the new current node
		if err := database.SetCurrentNode(nodeID); err != nil {
			fmt.Printf("\033[31m❌ Failed to set current node: %v\033[0m\n", err)
			os.Exit(1)
		}

		fmt.Printf("📍 \033[32mMoved to node:\033[0m \033[33m%s\033[0m\n", nodeID)
		var typeIcon string
		if node.Type == "user" {
			typeIcon = "👤"
		} else {
			typeIcon = "🤖"
		}
		fmt.Printf("%s Type: \033[90m%s\033[0m\n", typeIcon, node.Type)
		if node.Model != nil {
			fmt.Printf("🧠 Model: \033[35m%s\033[0m\n", *node.Model)
		}
		if node.Parent != nil {
			fmt.Printf("⬆️  Parent: \033[33m%s\033[0m\n", *node.Parent)
		}
		fmt.Printf("💬 Message: \033[90m%s\033[0m\n", node.Content)
	},
}

func init() {
	rootCmd.AddCommand(checkoutCmd)
}
