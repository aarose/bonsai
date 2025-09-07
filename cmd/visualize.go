package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aarose/bonsai/web"
	"github.com/spf13/cobra"
)

var visualizeCmd = &cobra.Command{
	Use:   "visualize",
	Short: "Launch web visualization for conversation trees",
	Long: `Launch a web server that provides an interactive D3.js visualization 
of your conversation trees. The server will serve an HTML page with 
a tree visualization that shows the branching structure of your conversations.

The visualization includes:
- Interactive tree layout with expand/collapse functionality
- Hover tooltips showing full message content
- Different colors for user vs LLM messages
- Zoom and pan controls
- Real-time refresh capability`,
	Example: `  # Launch with default settings (port 8080)
  bai visualize

  # Launch on specific port
  bai visualize --port 3000

  # Use custom database file
  bai visualize --database ./custom.db`,
	Run: runVisualize,
}

var (
	visualizePort int
	visualizeDB   string
)

func runVisualize(cmd *cobra.Command, args []string) {
	// Determine database path
	dbPath := visualizeDB
	if dbPath == "" {
		defaultPath, err := web.GetDefaultDatabasePath()
		if err != nil {
			log.Fatalf("Failed to get default database path: %v", err)
		}
		dbPath = defaultPath
	}

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("❌ Database not found at: %s\n\n", dbPath)
		fmt.Println("💡 Try generating some fake data first:")
		fmt.Println("   ./scripts/generate_fake_data.sh")
		fmt.Println("   or")
		fmt.Println("   go run scripts/generate_fake_data.go")
		os.Exit(1)
	}

	// Find available port if the specified one is in use
	actualPort := web.FindAvailablePort(visualizePort)
	if actualPort != visualizePort {
		fmt.Printf("⚠️  Port %d is in use, using port %d instead\n", visualizePort, actualPort)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		err := web.StartVisualizationServer(dbPath, actualPort)
		serverErr <- err
	}()

	// Wait for either an error or interrupt signal
	select {
	case err := <-serverErr:
		if err != nil {
			log.Fatalf("Server error: %v", err)
		}
	case sig := <-sigChan:
		fmt.Printf("\n🛑 Received signal: %v\n", sig)
		fmt.Println("👋 Shutting down visualization server...")
		os.Exit(0)
	}
}

func init() {
	rootCmd.AddCommand(visualizeCmd)

	// Add flags
	visualizeCmd.Flags().IntVarP(&visualizePort, "port", "p", 8080, 
		"Port to run the web server on")
	visualizeCmd.Flags().StringVarP(&visualizeDB, "database", "d", "", 
		"Path to database file (defaults to ~/.bonsai/bonsai.db)")
}