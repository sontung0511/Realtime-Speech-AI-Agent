package event

// Client -> Server event types
const (
	TypeStartSession = "start_session"
	TypeAudioChunk   = "audio_chunk"
	TypeStopSession  = "stop_session"
	TypeInterrupt    = "interrupt"
)

// Server -> Client event types
const (
	TypeSessionStarted  = "session_started"
	TypePartialText     = "partial_text"
	TypeFinalText       = "final_text"
	TypeAnswerDelta     = "answer_delta"
	TypeAnswerCompleted = "answer_completed"
	TypeNoteCreated     = "note_created"
	TypeSpeechStarted   = "speech_started"
	TypeSpeechEnded     = "speech_ended"
	TypeTTSAudio        = "tts_audio"
	TypeError           = "error"
)

// RealtimeEvent is the unified message envelope for all WebSocket messages.
type RealtimeEvent struct {
	Type       string `json:"type"`
	SessionID  string `json:"session_id"`
	Lang       string `json:"lang,omitempty"`
	Text       string `json:"text,omitempty"`
	StartMS    int64  `json:"start_ms,omitempty"`
	EndMS      int64  `json:"end_ms,omitempty"`
	Format     string `json:"format,omitempty"`
	SampleRate int    `json:"sample_rate,omitempty"`
	Data       string `json:"data,omitempty"`
	Status     string `json:"status,omitempty"`
	Code       string `json:"code,omitempty"`
	Message    string `json:"message,omitempty"`
	Language   string `json:"language,omitempty"`
	EnableTTS  bool   `json:"enable_tts,omitempty"`
	SourceText string `json:"source_text,omitempty"`
}

// Error codes
const (
	ErrInvalidMessage       = "invalid_message"
	ErrInvalidJSON          = "invalid_json"
	ErrInvalidSessionID     = "invalid_session_id"
	ErrSessionNotFound      = "session_not_found"
	ErrSessionAlreadyExists = "session_already_exists"
	ErrInvalidAudioFormat   = "invalid_audio_format"
	ErrInvalidSampleRate    = "invalid_sample_rate"
	ErrAudioPayloadTooLarge = "audio_payload_too_large"
	ErrSTTFailed            = "stt_failed"
	ErrLLMFailed            = "llm_failed"
	ErrTTSFailed            = "tts_failed"
	ErrStorageFailed        = "storage_failed"
	ErrTimeout              = "timeout"
	ErrInternalError        = "internal_error"
)
