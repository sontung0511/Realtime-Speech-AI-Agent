package llm

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// MockProvider is a mock LLM provider for local development and testing.
type MockProvider struct{}

// NewMockProvider creates a new MockProvider.
func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

// StreamAnswer simulates streaming LLM response.
func (m *MockProvider) StreamAnswer(sessionID, language, summary, latestText string, notes []string, onDelta func(text string), onComplete func(fullText string)) error {
	answer := generateMockAnswer(language, latestText)
	words := strings.Fields(answer)

	var full strings.Builder
	for i, w := range words {
		if i > 0 {
			full.WriteString(" ")
			onDelta(" ")
		}
		full.WriteString(w)
		onDelta(w)
		time.Sleep(30 * time.Millisecond)
	}

	onComplete(full.String())
	return nil
}

// DetectLanguage uses heuristics for mock language detection.
func (m *MockProvider) DetectLanguage(text string) (string, float64, error) {
	// Simple heuristic: check for non-ASCII characters common in Vietnamese
	for _, r := range text {
		if r > 127 && unicode.IsLetter(r) {
			return "vi", 0.85, nil
		}
	}
	return "en", 0.90, nil
}

func generateMockAnswer(language, text string) string {
	if language == "vi" {
		return fmt.Sprintf("Đây là câu trả lời mẫu cho: %s", truncate(text, 50))
	}
	return fmt.Sprintf("This is a mock AI answer for: %s", truncate(text, 50))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
