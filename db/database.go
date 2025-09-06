package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type Database struct {
	conn *sql.DB
	path string
}

type Node struct {
	ID       string `json:"id"`
	Content  string `json:"content"`
	Type     string `json:"type"`
	Parent   *string `json:"parent,omitempty"`
	Children string `json:"children"`
	Model    *string `json:"model,omitempty"`
}

// NewDatabase creates a new database connection
func NewDatabase(dbPath string) (*Database, error) {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &Database{
		conn: conn,
		path: dbPath,
	}

	return db, nil
}

// Initialize creates the necessary tables if they don't exist
func (db *Database) Initialize() error {
	createNodeTable := `
	CREATE TABLE IF NOT EXISTS Node (
		id TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		type TEXT NOT NULL CHECK (type IN ('user', 'llm')),
		parent TEXT,
		children TEXT DEFAULT '[]',
		model TEXT
	);`

	if _, err := db.conn.Exec(createNodeTable); err != nil {
		return fmt.Errorf("failed to create Node table: %w", err)
	}

	fmt.Println("Database initialized successfully")
	return nil
}

// Close closes the database connection
func (db *Database) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// GetConnection returns the underlying database connection
func (db *Database) GetConnection() *sql.DB {
	return db.conn
}

// GetPath returns the database file path
func (db *Database) GetPath() string {
	return db.path
}

// CreateRootNode creates a new root node with the given content and user type
func (db *Database) CreateRootNode(content string, model *string) (*Node, error) {
	node := &Node{
		ID:       uuid.New().String(),
		Content:  content,
		Type:     "user",
		Parent:   nil,
		Children: "[]",
		Model:    model,
	}

	return node, db.InsertNode(node)
}

// InsertNode inserts a node into the database
func (db *Database) InsertNode(node *Node) error {
	query := `
		INSERT INTO Node (id, content, type, parent, children, model)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := db.conn.Exec(query, node.ID, node.Content, node.Type, node.Parent, node.Children, node.Model)
	if err != nil {
		return fmt.Errorf("failed to insert node: %w", err)
	}

	return nil
}
