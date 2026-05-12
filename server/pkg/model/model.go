package model

import "time"

// Session statuses
const (
	StatusIdle         = "idle"
	StatusListening    = "listening"
	StatusTranscribing = "transcribing"
	StatusThinking     = "thinking"
	StatusAnswering    = "answering"
	StatusSpeaking     = "speaking"
	StatusError        = "error"
)

// Session represents a realtime speech session.
type Session struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Language  string    `json:"language"`
	EnableTTS bool      `json:"enable_tts"`
	Summary   string    `json:"summary"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TranscriptSegment represents a single speech-to-text segment.
type TranscriptSegment struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	StartMS   int64     `json:"start_ms"`
	EndMS     int64     `json:"end_ms"`
	Language  string    `json:"language"`
	Text      string    `json:"text"`
	IsFinal   bool      `json:"is_final"`
	Speaker   string    `json:"speaker"`
	CreatedAt time.Time `json:"created_at"`
}

// Note represents a note extracted from transcripts.
type Note struct {
	ID         string    `json:"id"`
	SessionID  string    `json:"session_id"`
	Text       string    `json:"text"`
	SourceText string    `json:"source_text"`
	CreatedAt  time.Time `json:"created_at"`
}

// AIAnswer represents the AI-generated answer.
type AIAnswer struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Language  string    `json:"language"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

// SessionContext holds the in-memory context for a session.
type SessionContext struct {
	SessionID      string   `json:"session_id"`
	Language       string   `json:"language"`
	Summary        string   `json:"summary"`
	LastTranscript string   `json:"last_transcript"`
	LastAnswer     string   `json:"last_answer"`
	LastIntent     string   `json:"intent"`
	Notes          []string `json:"notes"`
}
