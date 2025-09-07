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
			var statusMessage string

			if currentNodeID != nil {
				// Check if this root node is the current working node
				if *currentNodeID == node.ID {
					statusMessage = " \033[32m(current working node)\033[0m"
				} else {
					// Check if current working node is a descendant of this root node
					descendants, err := database.GetNodeAndAllChildren(node.ID)
					if err != nil {
						fmt.Printf("Warning: Failed to get descendants for node %s: %v\n", node.ID, err)
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
