package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var cherryPickCmd = &cobra.Command{
	Use:   "cherry-pick <node-id>",
	Short: "Duplicate a node's content as a child of the current working node (graft a tree bud)",
	Long:  `Duplicates the content of the specified node and creates it as a child of the current working node with a new ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sourceNodeID := args[0]

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

		// Get the source node to cherry-pick
		sourceNode, err := database.GetNodeByID(sourceNodeID)
		if err != nil {
			fmt.Printf("\033[31m❌ Error: %v\033[0m\n", err)
			os.Exit(1)
		}

		// Check if we're trying to cherry-pick the current working node
		if sourceNodeID == *currentNodeID {
			fmt.Printf("\033[31m❌ Cannot cherry-pick the current working node onto itself.\033[0m\n")
			os.Exit(1)
		}

		// Create a duplicate of the source node as a child of current working node
		// We preserve the source node's type and model
		duplicateNode, err := database.CreateChildNodeWithType(sourceNode.Content, *currentNodeID, sourceNode.Type, sourceNode.Model)
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to create cherry-picked node: %v\033[0m\n", err)
			os.Exit(1)
		}

			fmt.Printf("🍒 \033[32mCherry-picked node:\033[0m \033[33m%s\033[0m\n", sourceNodeID)
			fmt.Printf("✨ \033[32mCreated new node with ID:\033[0m \033[33m%s\033[0m\n", duplicateNode.ID)
			var typeIcon string
			if duplicateNode.Type == "user" {
				typeIcon = "👤"
			} else {
				typeIcon = "🤖"
			}
			fmt.Printf("%s Type: \033[90m%s\033[0m\n", typeIcon, duplicateNode.Type)

			fmt.Printf("⬆️  Parent: \033[33m%s\033[0m\n", *currentNodeID)
		if duplicateNode.Model != nil {
			fmt.Printf("🧠 Model: \033[35m%s\033[0m\n", *duplicateNode.Model)
		}

		// Show a preview of the content
		content := duplicateNode.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		content = strings.ReplaceAll(content, "\n", " ")
			fmt.Printf("💬 Message: \033[90m%s\033[0m\n", content)

			fmt.Printf("\n\033[32m✓ Successfully cherry-picked content from \033[33m%s\033[32m to new node \033[33m%s\033[0m\n", sourceNodeID, duplicateNode.ID)
	},
}

func init() {
	rootCmd.AddCommand(cherryPickCmd)
}
