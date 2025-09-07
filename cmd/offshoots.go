package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var offshootsCmd = &cobra.Command{
	Use:   "offshoots",
	Short: "List all conversation branches of the current working node",
	Long:  `List all conversation branches of the current working node. Shows the node ID, type, and a preview of their content.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize database
		database, err := initializeDatabase(false)
		if err != nil {
			fmt.Printf("\033[31m❌ %v\033[0m\n", err)
			os.Exit(1)
		}
		defer database.Close()

		// Get current working node
		currentNodeID, err := database.GetCurrentNode()
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to get current node: %v\033[0m\n", err)
			os.Exit(1)
		}

		if currentNodeID == nil {
			fmt.Println("\033[90mℹ️  No current working node set. Use 'bai seed' to create a root node or 'bai checkout' to move to an existing node.\033[0m")
			return
		}

		// Get direct children of the current node
		children, err := database.GetDirectChildren(*currentNodeID)
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to get child nodes: %v\033[0m\n", err)
			os.Exit(1)
		}

		if len(children) == 0 {
			fmt.Printf("🌿 No offshoots found for current node: \033[33m%s\033[0m\n", *currentNodeID)
			return
		}

		fmt.Printf("🌿 Child nodes of current working node (\033[33m%s\033[0m):\n\n", *currentNodeID)

		for i, child := range children {
			fmt.Printf("ID: \033[33m%s\033[0m\n",child.ID)
			if child.Model != nil {
				fmt.Printf("🧠 Model: \033[35m%s\033[0m\n", *child.Model)
			}

			// Show a preview of the content (first 100 characters)
			content := child.Content
			if len(content) > 100 {
				content = content[:100] + "..."
			}
			// Replace newlines with spaces for cleaner display
			content = strings.ReplaceAll(content, "\n", " ")
			fmt.Printf("💬 Message: \033[90m%s\033[0m\n", content)

			// Add spacing between nodes except for the last one
			if i < len(children)-1 {
				fmt.Println()
			}
		}

		fmt.Printf("\n\033[90mTotal: %d child node(s)\033[0m\n", len(children))
	},
}

func init() {
	rootCmd.AddCommand(offshootsCmd)
}
