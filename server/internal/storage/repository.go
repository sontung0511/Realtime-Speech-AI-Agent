package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sontungAI/realtime-speech-ai-agent/pkg/model"
)

// Repository handles persistent storage via SQLite.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new Repository and initializes the DB schema.
func NewRepository(dbPath string) (*Repository, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, err
	}

	repo := &Repository{db: db}
	if err := repo.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return repo, nil
}

func (r *Repository) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		status TEXT NOT NULL,
		language TEXT,
		enable_tts BOOLEAN DEFAULT FALSE,
		summary TEXT,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS transcript_segments (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		start_ms BIGINT,
		end_ms BIGINT,
		language TEXT,
		text TEXT NOT NULL,
		is_final BOOLEAN NOT NULL,
		speaker TEXT,
		created_at TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS notes (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		text TEXT NOT NULL,
		source_text TEXT,
		created_at TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS ai_answers (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		language TEXT,
		text TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL
	);
	`
	_, err := r.db.Exec(schema)
	return err
}

// SaveSession saves or updates a session.
func (r *Repository) SaveSession(s *model.Session) error {
	_, err := r.db.Exec(
		`INSERT INTO sessions (id, status, language, enable_tts, summary, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET status=?, language=?, enable_tts=?, summary=?, updated_at=?`,
		s.ID, s.Status, s.Language, s.EnableTTS, s.Summary, s.CreatedAt, s.UpdatedAt,
		s.Status, s.Language, s.EnableTTS, s.Summary, s.UpdatedAt,
	)
	return err
}

// SaveTranscript saves a transcript segment.
func (r *Repository) SaveTranscript(t *model.TranscriptSegment) error {
	_, err := r.db.Exec(
		`INSERT INTO transcript_segments (id, session_id, start_ms, end_ms, language, text, is_final, speaker, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.SessionID, t.StartMS, t.EndMS, t.Language, t.Text, t.IsFinal, t.Speaker, t.CreatedAt,
	)
	return err
}

// SaveNote saves a note.
func (r *Repository) SaveNote(n *model.Note) error {
	_, err := r.db.Exec(
		`INSERT INTO notes (id, session_id, text, source_text, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		n.ID, n.SessionID, n.Text, n.SourceText, n.CreatedAt,
	)
	return err
}

// SaveAnswer saves an AI answer.
func (r *Repository) SaveAnswer(a *model.AIAnswer) error {
	_, err := r.db.Exec(
		`INSERT INTO ai_answers (id, session_id, language, text, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		a.ID, a.SessionID, a.Language, a.Text, a.CreatedAt,
	)
	return err
}

// GetTranscripts returns all transcripts for a session.
func (r *Repository) GetTranscripts(sessionID string) ([]model.TranscriptSegment, error) {
	rows, err := r.db.Query(
		`SELECT id, session_id, start_ms, end_ms, language, text, is_final, speaker, created_at
		 FROM transcript_segments WHERE session_id = ? ORDER BY created_at`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var segments []model.TranscriptSegment
	for rows.Next() {
		var s model.TranscriptSegment
		if err := rows.Scan(&s.ID, &s.SessionID, &s.StartMS, &s.EndMS, &s.Language, &s.Text, &s.IsFinal, &s.Speaker, &s.CreatedAt); err != nil {
			return nil, err
		}
		segments = append(segments, s)
	}
	return segments, rows.Err()
}

// GetNotes returns all notes for a session.
func (r *Repository) GetNotes(sessionID string) ([]model.Note, error) {
	rows, err := r.db.Query(
		`SELECT id, session_id, text, source_text, created_at
		 FROM notes WHERE session_id = ? ORDER BY created_at`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var noteList []model.Note
	for rows.Next() {
		var n model.Note
		var createdAt time.Time
		if err := rows.Scan(&n.ID, &n.SessionID, &n.Text, &n.SourceText, &createdAt); err != nil {
			return nil, err
		}
		n.CreatedAt = createdAt
		noteList = append(noteList, n)
	}
	return noteList, rows.Err()
}

// Close closes the database connection.
func (r *Repository) Close() error {
	return r.db.Close()
}
