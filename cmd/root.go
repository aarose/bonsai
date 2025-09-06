package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aarose/bonsai/db"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bai",
	Short: "Bai is a CLI tool",
	Long: `Bai is a command-line interface tool that provides various utilities.
This application is built with love using Cobra in Go.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize database
		if err := initializeDatabase(); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		fmt.Println("Hello from bai! Use --help to see available commands.")
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func initializeDatabase() error {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create database path in user's home directory
	dbPath := filepath.Join(homeDir, ".bonsai", "bonsai.db")

	// Create and initialize database
	database, err := db.NewDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	defer database.Close()

	if err := database.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	fmt.Printf("Database created at: %s\n", dbPath)
	return nil
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
