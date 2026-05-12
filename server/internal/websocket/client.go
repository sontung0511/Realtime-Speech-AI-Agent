package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a connected WebSocket client.
type Client struct {
	conn      *websocket.Conn
	sessionID string
	mu        sync.Mutex
}

// NewClient creates a new Client.
func NewClient(conn *websocket.Conn) *Client {
	return &Client{conn: conn}
}

// SetSessionID sets the session ID for this client.
func (c *Client) SetSessionID(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sessionID = id
}

// SessionID returns the current session ID.
func (c *Client) SessionID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sessionID
}

// SendJSON sends a JSON message to the client.
func (c *Client) SendJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(v)
}

// ReadMessage reads one raw JSON message from the WebSocket.
func (c *Client) ReadMessage() ([]byte, error) {
	_, msg, err := c.conn.ReadMessage()
	return msg, err
}

// Close closes the WebSocket connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// ParseEvent parses a raw JSON message into a map.
func ParseEvent(data []byte) (map[string]interface{}, error) {
	var evt map[string]interface{}
	if err := json.Unmarshal(data, &evt); err != nil {
		return nil, err
	}
	return evt, nil
}

// GetString safely extracts a string field from a parsed event.
func GetString(evt map[string]interface{}, key string) string {
	if v, ok := evt[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetInt safely extracts an int field from a parsed event.
func GetInt(evt map[string]interface{}, key string) int {
	if v, ok := evt[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return 0
}

// GetBool safely extracts a bool field from a parsed event.
func GetBool(evt map[string]interface{}, key string) bool {
	if v, ok := evt[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// LogEvent logs an event for debugging.
func LogEvent(direction, sessionID, eventType string) {
	log.Printf("[ws] %s session=%s type=%s", direction, sessionID, eventType)
}
