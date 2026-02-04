package memory

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Message represents a chat message
type Message struct {
	ID        int64     `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	SessionID string    `json:"session_id"`
}

// Store manages message storage
type Store struct {
	db        *sql.DB
	sessionID string
}

// New creates a new memory store
func New(dbPath, sessionID string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create table if not exists
	if err := createSchema(db); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	store := &Store{
		db:        db,
		sessionID: sessionID,
	}

	return store, nil
}

func createSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_session_timestamp ON messages(session_id, timestamp);
	`
	_, err := db.Exec(query)
	return err
}

// Add adds a message to the store
func (s *Store) Add(role, content string) error {
	query := `
		INSERT INTO messages (session_id, role, content, timestamp)
		VALUES (?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, s.sessionID, role, content, time.Now())
	return err
}

// GetHistory retrieves message history
func (s *Store) GetHistory(limit int) ([]Message, error) {
	query := `
		SELECT id, role, content, timestamp, session_id
		FROM messages
		WHERE session_id = ?
		ORDER BY timestamp ASC
		LIMIT ?
	`
	rows, err := s.db.Query(query, s.sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.Role, &msg.Content, &msg.Timestamp, &msg.SessionID); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

// GetRecent retrieves recent messages
func (s *Store) GetRecent(count int) ([]Message, error) {
	return s.GetHistory(count)
}

// Clear clears all messages for the current session
func (s *Store) Clear() error {
	query := `DELETE FROM messages WHERE session_id = ?`
	_, err := s.db.Exec(query, s.sessionID)
	return err
}

// Count returns the number of messages in the current session
func (s *Store) Count() (int, error) {
	query := `SELECT COUNT(*) FROM messages WHERE session_id = ?`
	var count int
	err := s.db.QueryRow(query, s.sessionID).Scan(&count)
	return count, err
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// ToProviderFormat converts messages to provider format
func (s *Store) ToProviderFormat(limit int) ([]map[string]string, error) {
	messages, err := s.GetHistory(limit)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]string, len(messages))
	for i, msg := range messages {
		result[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	return result, nil
}

// ImportJSON imports messages from JSON format
func (s *Store) ImportJSON(jsonData []byte) error {
	var messages []Message
	if err := json.Unmarshal(jsonData, &messages); err != nil {
		return err
	}

	for _, msg := range messages {
		if err := s.Add(msg.Role, msg.Content); err != nil {
			return err
		}
	}

	return nil
}

// ExportJSON exports messages to JSON format
func (s *Store) ExportJSON() ([]byte, error) {
	messages, err := s.GetHistory(-1) // Get all messages
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(messages, "", "  ")
}
