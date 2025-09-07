package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var seedsCmd = &cobra.Command{
	Use:   "seeds",
	Short: "List all root nodes (conversation tree seeds)",
	Long:  `List all root nodes (conversation tree seeds)`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize database
		database, err := initializeDatabase(false)
		if err != nil {
			fmt.Printf("\033[31m❌ %v\033[0m\n", err)
			os.Exit(1)
		}
		defer database.Close()

		// Get all root nodes
		rootNodes, err := database.GetRootNodes()
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to get root nodes: %v\033[0m\n", err)
			os.Exit(1)
		}

		// Get current working node
		currentNodeID, err := database.GetCurrentNode()
		if err != nil {
			fmt.Printf("\033[31m❌ Failed to get current node: %v\033[0m\n", err)
			os.Exit(1)
		}

		if len(rootNodes) == 0 {
			fmt.Println("\033[90mℹ️  No root nodes found.\033[0m")
			return
		}

		fmt.Printf("🌱 Found %d seed(s) in the Bonsai garden:\n\n", len(rootNodes))
		for _, node := range rootNodes {
			var statusMessage string

			if currentNodeID != nil {
				// Check if this root node is the current working node
				if *currentNodeID == node.ID {
					statusMessage = " \033[32m(current working node)\033[0m"
				} else {
					// Check if current working node is a descendant of this root node
					descendants, err := database.GetNodeAndAllChildren(node.ID)
					if err != nil {
						fmt.Printf("\033[33m⚠️  Warning: Failed to get descendants for node %s: %v\033[0m\n", node.ID, err)
					} else {
						// Check descendants (excluding the root node itself)
						for _, descendant := range descendants {
							if descendant.ID == *currentNodeID && descendant.ID != node.ID {
								statusMessage = " \033[36m(contains current working node)\033[0m"
								break
							}
						}
					}
				}
			}

			// Print the node with highlighting if applicable
			fmt.Printf("ID: \033[33m%s\033[0m%s\n", node.ID, statusMessage)

			if node.Model != nil {
				fmt.Printf("🧠 Model: \033[35m%s\033[0m\n", *node.Model)
			}
			fmt.Printf("💬 Message: \033[90m%s\033[0m\n", node.Content)
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(seedsCmd)
}
