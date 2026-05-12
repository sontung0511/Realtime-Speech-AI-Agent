package stt

// Provider defines the interface for Speech-To-Text services.
type Provider interface {
	// ProcessAudio sends audio data and returns transcript events via callback.
	ProcessAudio(sessionID string, audioData []byte, onPartial func(lang, text string, startMS, endMS int64), onFinal func(lang, text string, startMS, endMS int64)) error
}
