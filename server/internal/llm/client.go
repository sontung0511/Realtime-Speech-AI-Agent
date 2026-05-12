package llm

// Provider defines the interface for LLM services.
type Provider interface {
	// StreamAnswer generates a streaming answer. onDelta is called for each token,
	// onComplete is called with the full answer text.
	StreamAnswer(sessionID, language, summary, latestText string, notes []string, onDelta func(text string), onComplete func(fullText string)) error

	// DetectLanguage uses AI to detect the language of the given text.
	// Returns a language code (e.g. "en", "vi", "ja") and confidence score.
	DetectLanguage(text string) (lang string, confidence float64, err error)
}
