package session

import (
	"sync"
	"time"

	"github.com/sontungAI/realtime-speech-ai-agent/pkg/model"
)

// Manager handles session lifecycle.
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*model.Session
}

// NewManager creates a new session manager.
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*model.Session),
	}
}

// Create creates a new session. Returns error if session already exists.
func (m *Manager) Create(id, language string, enableTTS bool) (*model.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[id]; exists {
		return nil, ErrSessionAlreadyExists
	}

	if language == "" {
		language = "auto"
	}

	now := time.Now()
	s := &model.Session{
		ID:        id,
		Status:    model.StatusListening,
		Language:  language,
		EnableTTS: enableTTS,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.sessions[id] = s
	return s, nil
}

// Get retrieves a session by ID.
func (m *Manager) Get(id string) (*model.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, ok := m.sessions[id]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return s, nil
}

// UpdateStatus updates the status of a session.
func (m *Manager) UpdateStatus(id, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[id]
	if !ok {
		return ErrSessionNotFound
	}
	s.Status = status
	s.UpdatedAt = time.Now()
	return nil
}

// UpdateLanguage updates the language of a session.
func (m *Manager) UpdateLanguage(id, lang string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[id]
	if !ok {
		return ErrSessionNotFound
	}
	s.Language = lang
	s.UpdatedAt = time.Now()
	return nil
}

// UpdateSummary updates the rolling summary of a session.
func (m *Manager) UpdateSummary(id, summary string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[id]
	if !ok {
		return ErrSessionNotFound
	}
	s.Summary = summary
	s.UpdatedAt = time.Now()
	return nil
}

// Delete removes a session.
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sessions[id]; !ok {
		return ErrSessionNotFound
	}
	delete(m.sessions, id)
	return nil
}

// List returns all active sessions.
func (m *Manager) List() []*model.Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*model.Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		result = append(result, s)
	}
	return result
}
