package stt

import (
	"strings"
	"time"
)

// MockProvider is a mock STT provider for local development and testing.
type MockProvider struct{}

// NewMockProvider creates a new MockProvider.
func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

// ProcessAudio simulates processing audio data and returns mock transcript results.
func (m *MockProvider) ProcessAudio(sessionID string, audioData []byte, onPartial func(lang, text string, startMS, endMS int64), onFinal func(lang, text string, startMS, endMS int64)) error {
	// Simulate processing delay
	time.Sleep(200 * time.Millisecond)

	mockText := "Hello, this is a test transcript from the mock STT provider."
	words := strings.Fields(mockText)

	// Simulate partial transcripts
	partial := ""
	for i, w := range words {
		if i > 0 {
			partial += " "
		}
		partial += w
		endMS := int64((i + 1) * 300)
		onPartial("en", partial, 0, endMS)
		time.Sleep(50 * time.Millisecond)
	}

	// Send final transcript
	onFinal("en", mockText, 0, int64(len(words)*300))
	return nil
}
