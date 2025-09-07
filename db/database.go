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
	ID       string  `json:"id"`
	Content  string  `json:"content"`
	Type     string  `json:"type"`
	Parent   *string `json:"parent,omitempty"`
	Children string  `json:"children"`
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

	createConfigTable := `
	CREATE TABLE IF NOT EXISTS Config (
		key TEXT PRIMARY KEY,
		value TEXT
	);`

	if _, err := db.conn.Exec(createConfigTable); err != nil {
		return fmt.Errorf("failed to create Config table: %w", err)
	}

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

	if err := db.InsertNode(node); err != nil {
		return nil, err
	}

	// Set this as the current working node
	if err := db.SetCurrentNode(node.ID); err != nil {
		return node, fmt.Errorf("created node but failed to set as current: %w", err)
	}

	return node, nil
}

// CreateChildNode creates a new child node with the given content, parent, and optional model
func (db *Database) CreateChildNode(content, parentID string, model *string) (*Node, error) {
	return db.CreateChildNodeWithType(content, parentID, "user", model)
}

// CreateChildNodeWithType creates a new child node with specific type
func (db *Database) CreateChildNodeWithType(content, parentID, nodeType string, model *string) (*Node, error) {
	// Verify parent exists
	_, err := db.GetNodeByID(parentID)
	if err != nil {
		return nil, fmt.Errorf("parent node not found: %w", err)
	}

	// Validate node type
	if nodeType != "user" && nodeType != "llm" {
		return nil, fmt.Errorf("invalid node type: %s (must be 'user' or 'llm')", nodeType)
	}

	node := &Node{
		ID:       uuid.New().String(),
		Content:  content,
		Type:     nodeType,
		Parent:   &parentID,
		Children: "[]",
		Model:    model,
	}

	if err := db.InsertNode(node); err != nil {
		return nil, err
	}

	// Set this as the current working node
	if err := db.SetCurrentNode(node.ID); err != nil {
		return node, fmt.Errorf("created node but failed to set as current: %w", err)
	}

	return node, nil
}

// CreateLLMResponseNode creates a new LLM response node as a child of the specified parent
func (db *Database) CreateLLMResponseNode(parentID, content, model string) (*Node, error) {
	modelPtr := &model
	return db.CreateChildNodeWithType(content, parentID, "llm", modelPtr)
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

// GetRootNodes retrieves all nodes that have no parent (root nodes)
func (db *Database) GetRootNodes() ([]*Node, error) {
	query := `
		SELECT id, content, type, parent, children, model
		FROM Node
		WHERE parent IS NULL
		ORDER BY id
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query root nodes: %w", err)
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		node := &Node{}
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

// GetCurrentNode retrieves the current working node ID
func (db *Database) GetCurrentNode() (*string, error) {
	query := `SELECT value FROM Config WHERE key = 'current_node'`

	var nodeID string
	err := db.conn.QueryRow(query).Scan(&nodeID)
	if err == sql.ErrNoRows {
		return nil, nil // No current node set
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get current node: %w", err)
	}

	return &nodeID, nil
}

// SetCurrentNode sets the current working node
func (db *Database) SetCurrentNode(nodeID string) error {
	query := `
		INSERT INTO Config (key, value) VALUES ('current_node', ?)
		ON CONFLICT(key) DO UPDATE SET value = ?
	`

	_, err := db.conn.Exec(query, nodeID, nodeID)
	if err != nil {
		return fmt.Errorf("failed to set current node: %w", err)
	}

	return nil
}

// ClearCurrentNode removes the current working node setting
func (db *Database) ClearCurrentNode() error {
	query := `DELETE FROM Config WHERE key = 'current_node'`

	_, err := db.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to clear current node: %w", err)
	}

	return nil
}

// GetNodeAndAllChildren retrieves a node and all its descendants recursively
func (db *Database) GetNodeAndAllChildren(nodeID string) ([]*Node, error) {
	var allNodes []*Node
	visited := make(map[string]bool)

	if err := db.collectNodeAndChildren(nodeID, &allNodes, visited); err != nil {
		return nil, err
	}

	return allNodes, nil
}

// collectNodeAndChildren is a recursive helper function to collect all descendant nodes
func (db *Database) collectNodeAndChildren(nodeID string, allNodes *[]*Node, visited map[string]bool) error {
	// Avoid infinite loops
	if visited[nodeID] {
		return nil
	}
	visited[nodeID] = true

	// Get the current node
	query := `SELECT id, content, type, parent, children, model FROM Node WHERE id = ?`
	row := db.conn.QueryRow(query, nodeID)

	node := &Node{}
	err := row.Scan(&node.ID, &node.Content, &node.Type, &node.Parent, &node.Children, &node.Model)
	if err == sql.ErrNoRows {
		return nil // Node doesn't exist, skip
	}
	if err != nil {
		return fmt.Errorf("failed to scan node %s: %w", nodeID, err)
	}

	*allNodes = append(*allNodes, node)

	// Find all children of this node
	childQuery := `SELECT id FROM Node WHERE parent = ?`
	rows, err := db.conn.Query(childQuery, nodeID)
	if err != nil {
		return fmt.Errorf("failed to query children of node %s: %w", nodeID, err)
	}
	defer rows.Close()

	var childIDs []string
	for rows.Next() {
		var childID string
		if err := rows.Scan(&childID); err != nil {
			return fmt.Errorf("failed to scan child ID: %w", err)
		}
		childIDs = append(childIDs, childID)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating over child rows: %w", err)
	}

	// Recursively collect each child and their descendants
	for _, childID := range childIDs {
		if err := db.collectNodeAndChildren(childID, allNodes, visited); err != nil {
			return err
		}
	}

	return nil
}

// GetNodeByID retrieves a single node by its ID
func (db *Database) GetNodeByID(nodeID string) (*Node, error) {
	query := `SELECT id, content, type, parent, children, model FROM Node WHERE id = ?`
	row := db.conn.QueryRow(query, nodeID)

	node := &Node{}
	err := row.Scan(&node.ID, &node.Content, &node.Type, &node.Parent, &node.Children, &node.Model)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("node with ID %s not found", nodeID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan node %s: %w", nodeID, err)
	}

	return node, nil
}

// GetDirectChildren retrieves all direct children of a node (non-recursive)
func (db *Database) GetDirectChildren(parentID string) ([]*Node, error) {
	query := `
		SELECT id, content, type, parent, children, model
		FROM Node
		WHERE parent = ?
		ORDER BY id
	`

	rows, err := db.conn.Query(query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query children of node %s: %w", parentID, err)
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		node := &Node{}
		err := rows.Scan(&node.ID, &node.Content, &node.Type, &node.Parent, &node.Children, &node.Model)
		if err != nil {
			return nil, fmt.Errorf("failed to scan child node: %w", err)
		}
		nodes = append(nodes, node)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over child rows: %w", err)
	}

	return nodes, nil
}

// GetParentPath retrieves the parent chain from a given node up to the root
// If maxLevels is 0 or negative, returns the complete path to the root
// If maxLevels is positive, returns up to that many parent levels
func (db *Database) GetParentPath(nodeID string, maxLevels int) ([]*Node, error) {
	var parentChain []*Node
	currentNodeID := nodeID
	levelsTraversed := 0

	for {
		// Stop if we've reached the maximum levels (when maxLevels > 0)
		if maxLevels > 0 && levelsTraversed >= maxLevels {
			break
		}

		// Get the current node
		currentNode, err := db.GetNodeByID(currentNodeID)
		if err != nil {
			return parentChain, fmt.Errorf("failed to get node %s: %w", currentNodeID, err)
		}

		// If this node has no parent, we've reached a root node
		if currentNode.Parent == nil {
			break
		}

		// Get the parent node
		parentNode, err := db.GetNodeByID(*currentNode.Parent)
		if err != nil {
			return parentChain, fmt.Errorf("failed to get parent node %s: %w", *currentNode.Parent, err)
		}

		// Add parent to the chain
		parentChain = append(parentChain, parentNode)

		// Move to the parent for the next iteration
		currentNodeID = *currentNode.Parent
		levelsTraversed++
	}

	return parentChain, nil
}

// GetConversationHistory retrieves the conversation history from a given node up to the root
// Returns messages in chronological order (root to current), suitable for sending to LLMs
func (db *Database) GetConversationHistory(nodeID string) ([]*Node, error) {
	var conversationChain []*Node
	currentNodeID := nodeID

	// Traverse up the parent chain to collect all nodes
	for {
		// Get the current node
		currentNode, err := db.GetNodeByID(currentNodeID)
		if err != nil {
			return nil, fmt.Errorf("failed to get node %s: %w", currentNodeID, err)
		}

		// Add current node to the front of the chain (we'll reverse it later)
		conversationChain = append([]*Node{currentNode}, conversationChain...)

		// If this node has no parent, we've reached a root node
		if currentNode.Parent == nil {
			break
		}

		// Move to the parent for the next iteration
		currentNodeID = *currentNode.Parent
	}

	return conversationChain, nil
}

// DeleteNodeAndAllChildren deletes a node and all its descendants recursively
func (db *Database) DeleteNodeAndAllChildren(nodeID string) (int, error) {
	// First, get all nodes to be deleted
	nodesToDelete, err := db.GetNodeAndAllChildren(nodeID)
	if err != nil {
		return 0, fmt.Errorf("failed to get nodes to delete: %w", err)
	}

	if len(nodesToDelete) == 0 {
		return 0, nil
	}

	// Delete all nodes (children first, then parents)
	// We'll delete in reverse order to handle foreign key constraints
	deletedCount := 0
	for i := len(nodesToDelete) - 1; i >= 0; i-- {
		node := nodesToDelete[i]

		deleteQuery := `DELETE FROM Node WHERE id = ?`
		result, err := db.conn.Exec(deleteQuery, node.ID)
		if err != nil {
			return deletedCount, fmt.Errorf("failed to delete node %s: %w", node.ID, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return deletedCount, fmt.Errorf("failed to get rows affected for node %s: %w", node.ID, err)
		}

		deletedCount += int(rowsAffected)
	}

	return deletedCount, nil
}
