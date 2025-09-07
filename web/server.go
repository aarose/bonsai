package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aarose/bonsai/db"
)

//go:embed index.html
var content embed.FS

// Server represents the web visualization server
type Server struct {
	db   *db.Database
	port int
}

// NewServer creates a new web server instance
func NewServer(database *db.Database, port int) *Server {
	return &Server{
		db:   database,
		port: port,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Serve the main HTML page
	mux.HandleFunc("/", s.handleIndex)

	// API endpoint to get tree data
	mux.HandleFunc("/api/tree", s.handleTreeData)

	// Health check endpoint
	mux.HandleFunc("/api/health", s.handleHealth)

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", s.port),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Printf("🌳 Bonsai visualization server starting...\n")
	fmt.Printf("📱 Open your browser to: http://localhost:%d\n", s.port)
	fmt.Printf("💡 Press Ctrl+C to stop the server\n\n")

	return server.ListenAndServe()
}

// handleIndex serves the main HTML page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the embedded HTML file
	htmlContent, err := content.ReadFile("index.html")
	if err != nil {
		http.Error(w, "Failed to read HTML template", http.StatusInternalServerError)
		log.Printf("Error reading HTML template: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Write(htmlContent)
}

// handleTreeData serves the conversation tree data as JSON
func (s *Server) handleTreeData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all nodes from database
	nodes, err := s.getAllNodes()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch tree data: %v", err), http.StatusInternalServerError)
		log.Printf("Error fetching tree data: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	if err := json.NewEncoder(w).Encode(nodes); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		log.Printf("Error encoding JSON: %v", err)
		return
	}
}

// handleHealth provides a simple health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"database":  s.db.GetPath(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}
}

// getAllNodes retrieves all nodes from the database
func (s *Server) getAllNodes() ([]*TreeNode, error) {
	query := `
		SELECT id, content, type, parent, children, model
		FROM Node
		ORDER BY id
	`

	rows, err := s.db.GetConnection().Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query nodes: %w", err)
	}
	defer rows.Close()

	var nodes []*TreeNode
	for rows.Next() {
		node := &TreeNode{}
		err := rows.Scan(&node.ID, &node.Content, &node.Type, &node.Parent, &node.Children, &node.Model)
		if err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}
		nodes = append(nodes, node)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return nodes, nil
}

// TreeNode represents a node in the conversation tree for JSON serialization
type TreeNode struct {
	ID       string  `json:"id"`
	Content  string  `json:"content"`
	Type     string  `json:"type"`
	Parent   *string `json:"parent,omitempty"`
	Children string  `json:"children"`
	Model    *string `json:"model,omitempty"`
}

// StartVisualizationServer is a convenience function to start the server
func StartVisualizationServer(dbPath string, port int) error {
	// Create database connection
	database, err := db.NewDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	// Initialize database (create tables if they don't exist)
	if err := database.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create and start server
	server := NewServer(database, port)
	return server.Start()
}

// GetDefaultDatabasePath returns the default database path used by the CLI
func GetDefaultDatabasePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".bonsai", "bonsai.db"), nil
}

// FindAvailablePort finds an available port starting from the given port
func FindAvailablePort(startPort int) int {
	for port := startPort; port < startPort+100; port++ {
		if isPortAvailable(port) {
			return port
		}
	}
	return startPort // fallback to original port if none found
}

// isPortAvailable checks if a port is available
func isPortAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false // Port is not available
	}
	listener.Close()
	return true // Port is available
}