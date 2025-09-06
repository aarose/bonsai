package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Database struct {
	conn *sql.DB
	path string
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
