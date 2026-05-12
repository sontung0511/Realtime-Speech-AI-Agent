package language

import (
	"log"
	"strings"

	"github.com/sontungAI/realtime-speech-ai-agent/internal/llm"
)

// DetectResult holds the language detection result.
type DetectResult struct {
	Language   string  `json:"language"`
	Confidence float64 `json:"confidence"`
}

// Detector uses an LLM provider for AI-based language detection.
type Detector struct {
	llmProvider llm.Provider
}

// NewDetector creates a new AI-powered language detector.
func NewDetector(provider llm.Provider) *Detector {
	return &Detector{llmProvider: provider}
}

// Detect detects the language of the given text using the LLM provider.
func (d *Detector) Detect(text string) DetectResult {
	if strings.TrimSpace(text) == "" {
		return DetectResult{Language: "en", Confidence: 0.0}
	}

	lang, confidence, err := d.llmProvider.DetectLanguage(text)
	if err != nil {
		log.Printf("[language] AI detection failed, falling back to 'en': %v", err)
		return DetectResult{Language: "en", Confidence: 0.5}
	}

	// Normalize the response
	lang = strings.TrimSpace(strings.ToLower(lang))
	if len(lang) > 5 {
		// LLM returned something unexpected, try to extract the code
		lang = lang[:2]
	}
	if lang == "" {
		lang = "en"
		confidence = 0.5
	}

	return DetectResult{Language: lang, Confidence: confidence}
}

// ResolveLanguage decides session language using detection result and fallback.
func ResolveLanguage(detected DetectResult, sessionLang string) string {
	if detected.Confidence >= 0.75 {
		return detected.Language
	}
	if sessionLang != "" && sessionLang != "auto" {
		return sessionLang
	}
	return "en"
}
