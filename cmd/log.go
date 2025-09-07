package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aarose/bonsai/db"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show parent nodes of the current working node",
	Long:  `Show the parent chain of the current working node. Use --up to specify levels or --all for complete path to root.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flag values
		upLevels, err := cmd.Flags().GetInt("up")
		if err != nil {
			fmt.Printf("Failed to get up flag: %v\n", err)
			os.Exit(1)
		}

		showAll, err := cmd.Flags().GetBool("all")
		if err != nil {
			fmt.Printf("Failed to get all flag: %v\n", err)
			os.Exit(1)
		}

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

		// Get the current node to show context
		currentNode, err := database.GetNodeByID(*currentNodeID)
		if err != nil {
			fmt.Printf("Failed to get current node details: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("🪵 Log from current working node: \033[33m%s\033[0m\n", *currentNodeID)
		fmt.Printf("Current node type: %s\n", currentNode.Type)
		if currentNode.Model != nil {
			fmt.Printf("Current node model: %s\n", *currentNode.Model)
		}
		fmt.Printf("Current node message: %s\n", currentNode.Content)
		fmt.Println()

		// Determine how many levels to traverse
		maxLevels := upLevels
		if showAll {
			maxLevels = 0 // 0 means traverse all the way to root
		}
		if maxLevels == 0 && !showAll {
			maxLevels = 1 // Default to showing immediate parent
		}

		// Get parent path
		parentPath, err := database.GetParentPath(*currentNodeID, maxLevels)
		if err != nil {
			fmt.Printf("Failed to get parent path: %v\n", err)
			os.Exit(1)
		}

		if len(parentPath) == 0 {
			fmt.Println("Current node has no parents - it's already at the root level.")
			return
		}

		// Display the parent chain
		for i, parent := range parentPath {
			level := i + 1
			var levelIndicator string
			if level == 1 {
				levelIndicator = "Parent"
			} else {
				levelIndicator = fmt.Sprintf("Level %d up", level)
			}

			fmt.Printf("%s:\n", levelIndicator)
			fmt.Printf("   ID: \033[33m%s\033[0m\n", parent.ID)
			fmt.Printf("   Type: %s\n", parent.Type)
			if parent.Model != nil {
				fmt.Printf("   Model: %s\n", *parent.Model)
			}

			// Show a preview of the content (first 150 characters)
			content := parent.Content
			if len(content) > 150 {
				content = content[:150] + "..."
			}
			// Replace newlines with spaces for cleaner display
			content = strings.ReplaceAll(content, "\n", " ")
			fmt.Printf("   Content: %s\n", content)

			// Add spacing between levels except for the last one
			if i < len(parentPath)-1 {
				fmt.Println()
			}
		}

		// Show summary
		fmt.Println()
		if showAll {
			if len(parentPath) == 1 {
				fmt.Println("Reached the root - this is the complete log.")
			} else {
				fmt.Printf("Complete log shown - %d level(s) to the root.\n", len(parentPath))
			}
		} else {
			fmt.Printf("Showing %d level(s) up. Use --all to see complete path to root.\n", len(parentPath))
		}
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
	logCmd.Flags().IntP("up", "u", 1, "Number of levels to climb up the parent chain")
	logCmd.Flags().BoolP("all", "a", false, "Show complete path to the root")
}
