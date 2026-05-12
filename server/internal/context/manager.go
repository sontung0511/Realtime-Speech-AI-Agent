package context

import (
	"sync"

	"github.com/sontungAI/realtime-speech-ai-agent/pkg/model"
)

// Manager manages session context in memory.
type Manager struct {
	mu       sync.RWMutex
	contexts map[string]*model.SessionContext
}

// NewManager creates a new context manager.
func NewManager() *Manager {
	return &Manager{
		contexts: make(map[string]*model.SessionContext),
	}
}

// Create initializes context for a session.
func (m *Manager) Create(sessionID, language string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.contexts[sessionID] = &model.SessionContext{
		SessionID: sessionID,
		Language:  language,
		Notes:     make([]string, 0),
	}
}

// Get retrieves context for a session.
func (m *Manager) Get(sessionID string) *model.SessionContext {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ctx, ok := m.contexts[sessionID]
	if !ok {
		return nil
	}
	// Return a copy to avoid data races
	cp := *ctx
	cp.Notes = make([]string, len(ctx.Notes))
	copy(cp.Notes, ctx.Notes)
	return &cp
}

// UpdateTranscript updates the last transcript and summary.
func (m *Manager) UpdateTranscript(sessionID, transcript, summary string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, ok := m.contexts[sessionID]
	if !ok {
		return
	}
	ctx.LastTranscript = transcript
	if summary != "" {
		ctx.Summary = summary
	}
}

// UpdateLanguage updates the session language.
func (m *Manager) UpdateLanguage(sessionID, language string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, ok := m.contexts[sessionID]
	if !ok {
		return
	}
	ctx.Language = language
}

// UpdateAnswer updates the last answer.
func (m *Manager) UpdateAnswer(sessionID, answer string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, ok := m.contexts[sessionID]
	if !ok {
		return
	}
	ctx.LastAnswer = answer
}

// AddNote adds a note to the session context. Keeps max 20 notes.
func (m *Manager) AddNote(sessionID, note string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, ok := m.contexts[sessionID]
	if !ok {
		return
	}
	ctx.Notes = append(ctx.Notes, note)
	if len(ctx.Notes) > 20 {
		ctx.Notes = ctx.Notes[len(ctx.Notes)-20:]
	}
}

// Delete removes context for a session.
func (m *Manager) Delete(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.contexts, sessionID)
}

// BuildRollingSummary creates a new summary from previous summary + latest transcript.
// For MVP, this is a simple concatenation. Can be replaced by LLM summarization later.
func BuildRollingSummary(previousSummary, latestTranscript string) string {
	if previousSummary == "" {
		return latestTranscript
	}
	summary := previousSummary + " " + latestTranscript
	// Keep summary under ~1500 chars
	if len(summary) > 1500 {
		summary = summary[len(summary)-1500:]
	}
	return summary
}
