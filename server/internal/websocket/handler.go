package websocket

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sontungAI/realtime-speech-ai-agent/internal/audio"
	ctxmgr "github.com/sontungAI/realtime-speech-ai-agent/internal/context"
	"github.com/sontungAI/realtime-speech-ai-agent/internal/language"
	"github.com/sontungAI/realtime-speech-ai-agent/internal/llm"
	"github.com/sontungAI/realtime-speech-ai-agent/internal/notes"
	"github.com/sontungAI/realtime-speech-ai-agent/internal/session"
	"github.com/sontungAI/realtime-speech-ai-agent/internal/storage"
	"github.com/sontungAI/realtime-speech-ai-agent/internal/stt"
	"github.com/sontungAI/realtime-speech-ai-agent/pkg/event"
	"github.com/sontungAI/realtime-speech-ai-agent/pkg/model"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Handler handles WebSocket connections and orchestrates the realtime pipeline.
type Handler struct {
	sessionMgr   *session.Manager
	contextMgr   *ctxmgr.Manager
	sttProvider  stt.Provider
	llmProvider  llm.Provider
	langDetector *language.Detector
	repo         *storage.Repository
	maxPayload   int

	// Track active answer streams so they can be cancelled on interrupt
	cancelMu sync.Mutex
	cancels  map[string]*atomic.Bool

	transcriptCounter atomic.Int64
	answerCounter     atomic.Int64
}

// NewHandler creates a new WebSocket handler.
func NewHandler(
	sessionMgr *session.Manager,
	contextMgr *ctxmgr.Manager,
	sttProvider stt.Provider,
	llmProvider llm.Provider,
	langDetector *language.Detector,
	repo *storage.Repository,
	maxPayload int,
) *Handler {
	return &Handler{
		sessionMgr:   sessionMgr,
		contextMgr:   contextMgr,
		sttProvider:  sttProvider,
		llmProvider:  llmProvider,
		langDetector: langDetector,
		repo:         repo,
		maxPayload:   maxPayload,
		cancels:      make(map[string]*atomic.Bool),
	}
}

// ServeHTTP upgrades the HTTP connection to WebSocket and starts the message loop.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ws] upgrade error: %v", err)
		return
	}

	client := NewClient(conn)
	defer func() {
		sid := client.SessionID()
		if sid != "" {
			h.cleanupSession(sid)
		}
		client.Close()
	}()

	log.Println("[ws] client connected")

	for {
		data, err := client.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[ws] read error: %v", err)
			}
			break
		}

		evt, err := ParseEvent(data)
		if err != nil {
			h.sendError(client, "", event.ErrInvalidJSON, "Invalid JSON message")
			continue
		}

		eventType := GetString(evt, "type")
		sessionID := GetString(evt, "session_id")

		switch eventType {
		case event.TypeStartSession:
			h.handleStartSession(client, evt)
		case event.TypeAudioChunk:
			h.handleAudioChunk(client, sessionID, evt)
		case event.TypeStopSession:
			h.handleStopSession(client, sessionID)
		case event.TypeInterrupt:
			h.handleInterrupt(client, sessionID)
		default:
			h.sendError(client, sessionID, event.ErrInvalidMessage, fmt.Sprintf("Unknown event type: %s", eventType))
		}
	}
}

func (h *Handler) handleStartSession(client *Client, evt map[string]interface{}) {
	sessionID := GetString(evt, "session_id")
	lang := GetString(evt, "language")
	enableTTS := GetBool(evt, "enable_tts")

	if sessionID == "" {
		h.sendError(client, "", event.ErrInvalidSessionID, "session_id is required")
		return
	}

	LogEvent("recv", sessionID, event.TypeStartSession)

	sess, err := h.sessionMgr.Create(sessionID, lang, enableTTS)
	if err != nil {
		h.sendError(client, sessionID, event.ErrSessionAlreadyExists, err.Error())
		return
	}

	client.SetSessionID(sessionID)
	h.contextMgr.Create(sessionID, sess.Language)

	// Save session to DB
	if h.repo != nil {
		if err := h.repo.SaveSession(sess); err != nil {
			log.Printf("[ws] save session error: %v", err)
		}
	}

	// Send session_started
	client.SendJSON(event.RealtimeEvent{
		Type:      event.TypeSessionStarted,
		SessionID: sessionID,
		Status:    model.StatusListening,
	})

	LogEvent("send", sessionID, event.TypeSessionStarted)
}

func (h *Handler) handleAudioChunk(client *Client, sessionID string, evt map[string]interface{}) {
	if sessionID == "" {
		h.sendError(client, "", event.ErrInvalidSessionID, "session_id is required")
		return
	}

	// Check session exists
	if _, err := h.sessionMgr.Get(sessionID); err != nil {
		h.sendError(client, sessionID, event.ErrSessionNotFound, err.Error())
		return
	}

	format := GetString(evt, "format")
	sampleRate := GetInt(evt, "sample_rate")
	data := GetString(evt, "data")

	// Validate audio
	audioData, err := audio.ValidateChunk(format, sampleRate, data, h.maxPayload)
	if err != nil {
		errCode := event.ErrInvalidAudioFormat
		errMsg := err.Error()
		if strings.Contains(errMsg, "sample_rate") {
			errCode = event.ErrInvalidSampleRate
		} else if strings.Contains(errMsg, "too_large") {
			errCode = event.ErrAudioPayloadTooLarge
		}
		h.sendError(client, sessionID, errCode, errMsg)
		return
	}

	h.sessionMgr.UpdateStatus(sessionID, model.StatusTranscribing)

	// Process audio through STT
	go h.processSTT(client, sessionID, audioData)
}

func (h *Handler) processSTT(client *Client, sessionID string, audioData []byte) {
	err := h.sttProvider.ProcessAudio(sessionID, audioData,
		// onPartial
		func(lang, text string, startMS, endMS int64) {
			client.SendJSON(event.RealtimeEvent{
				Type:      event.TypePartialText,
				SessionID: sessionID,
				Lang:      lang,
				Text:      text,
				StartMS:   startMS,
				EndMS:     endMS,
			})
		},
		// onFinal
		func(lang, text string, startMS, endMS int64) {
			h.handleFinalTranscript(client, sessionID, lang, text, startMS, endMS)
		},
	)
	if err != nil {
		log.Printf("[stt] error for session %s: %v", sessionID, err)
		h.sendError(client, sessionID, event.ErrSTTFailed, err.Error())
		h.sessionMgr.UpdateStatus(sessionID, model.StatusListening)
	}
}

func (h *Handler) handleFinalTranscript(client *Client, sessionID, lang, text string, startMS, endMS int64) {
	// Use AI to detect language
	detected := h.langDetector.Detect(text)
	sess, _ := h.sessionMgr.Get(sessionID)
	resolvedLang := language.ResolveLanguage(detected, "")
	if sess != nil {
		resolvedLang = language.ResolveLanguage(detected, sess.Language)
	}

	// Send final_text to client
	client.SendJSON(event.RealtimeEvent{
		Type:      event.TypeFinalText,
		SessionID: sessionID,
		Lang:      resolvedLang,
		Text:      text,
		StartMS:   startMS,
		EndMS:     endMS,
	})
	LogEvent("send", sessionID, event.TypeFinalText)

	// Save transcript
	tID := h.transcriptCounter.Add(1)
	segment := &model.TranscriptSegment{
		ID:        fmt.Sprintf("%s_t_%d", sessionID, tID),
		SessionID: sessionID,
		StartMS:   startMS,
		EndMS:     endMS,
		Language:  resolvedLang,
		Text:      text,
		IsFinal:   true,
		CreatedAt: time.Now(),
	}
	if h.repo != nil {
		if err := h.repo.SaveTranscript(segment); err != nil {
			log.Printf("[storage] save transcript error: %v", err)
		}
	}

	// Update context
	ctx := h.contextMgr.Get(sessionID)
	prevSummary := ""
	if ctx != nil {
		prevSummary = ctx.Summary
	}
	newSummary := ctxmgr.BuildRollingSummary(prevSummary, text)
	h.contextMgr.UpdateTranscript(sessionID, text, newSummary)
	h.contextMgr.UpdateLanguage(sessionID, resolvedLang)
	h.sessionMgr.UpdateLanguage(sessionID, resolvedLang)
	h.sessionMgr.UpdateSummary(sessionID, newSummary)

	// Extract notes
	note := notes.Extract(sessionID, text, newSummary)
	if note != nil {
		h.contextMgr.AddNote(sessionID, note.Text)
		if h.repo != nil {
			h.repo.SaveNote(note)
		}
		client.SendJSON(event.RealtimeEvent{
			Type:       event.TypeNoteCreated,
			SessionID:  sessionID,
			Text:       note.Text,
			SourceText: note.SourceText,
		})
		LogEvent("send", sessionID, event.TypeNoteCreated)
	}

	// Check if we should answer
	if isMeaningful(text) {
		h.sessionMgr.UpdateStatus(sessionID, model.StatusThinking)
		go h.generateAnswer(client, sessionID, resolvedLang, text)
	} else {
		h.sessionMgr.UpdateStatus(sessionID, model.StatusListening)
	}
}

func (h *Handler) generateAnswer(client *Client, sessionID, lang, text string) {
	ctx := h.contextMgr.Get(sessionID)
	summary := ""
	var ctxNotes []string
	if ctx != nil {
		summary = ctx.Summary
		ctxNotes = ctx.Notes
	}

	// Set up cancellation
	cancelled := &atomic.Bool{}
	h.cancelMu.Lock()
	h.cancels[sessionID] = cancelled
	h.cancelMu.Unlock()
	defer func() {
		h.cancelMu.Lock()
		delete(h.cancels, sessionID)
		h.cancelMu.Unlock()
	}()

	h.sessionMgr.UpdateStatus(sessionID, model.StatusAnswering)

	err := h.llmProvider.StreamAnswer(sessionID, lang, summary, text, ctxNotes,
		// onDelta
		func(delta string) {
			if cancelled.Load() {
				return
			}
			client.SendJSON(event.RealtimeEvent{
				Type:      event.TypeAnswerDelta,
				SessionID: sessionID,
				Lang:      lang,
				Text:      delta,
			})
		},
		// onComplete
		func(fullText string) {
			if cancelled.Load() {
				return
			}
			client.SendJSON(event.RealtimeEvent{
				Type:      event.TypeAnswerCompleted,
				SessionID: sessionID,
				Lang:      lang,
				Text:      fullText,
			})
			LogEvent("send", sessionID, event.TypeAnswerCompleted)

			// Save answer
			aID := h.answerCounter.Add(1)
			answer := &model.AIAnswer{
				ID:        fmt.Sprintf("%s_a_%d", sessionID, aID),
				SessionID: sessionID,
				Language:  lang,
				Text:      fullText,
				CreatedAt: time.Now(),
			}
			if h.repo != nil {
				h.repo.SaveAnswer(answer)
			}
			h.contextMgr.UpdateAnswer(sessionID, fullText)
		},
	)
	if err != nil {
		if !cancelled.Load() {
			log.Printf("[llm] error for session %s: %v", sessionID, err)
			h.sendError(client, sessionID, event.ErrLLMFailed, err.Error())
		}
	}

	if !cancelled.Load() {
		h.sessionMgr.UpdateStatus(sessionID, model.StatusListening)
	}
}

func (h *Handler) handleStopSession(client *Client, sessionID string) {
	if sessionID == "" {
		h.sendError(client, "", event.ErrInvalidSessionID, "session_id is required")
		return
	}
	LogEvent("recv", sessionID, event.TypeStopSession)
	h.cleanupSession(sessionID)
}

func (h *Handler) handleInterrupt(client *Client, sessionID string) {
	if sessionID == "" {
		h.sendError(client, "", event.ErrInvalidSessionID, "session_id is required")
		return
	}
	LogEvent("recv", sessionID, event.TypeInterrupt)

	// Cancel any active LLM stream
	h.cancelMu.Lock()
	if cancel, ok := h.cancels[sessionID]; ok {
		cancel.Store(true)
	}
	h.cancelMu.Unlock()

	h.sessionMgr.UpdateStatus(sessionID, model.StatusListening)
}

func (h *Handler) cleanupSession(sessionID string) {
	// Cancel active streams
	h.cancelMu.Lock()
	if cancel, ok := h.cancels[sessionID]; ok {
		cancel.Store(true)
		delete(h.cancels, sessionID)
	}
	h.cancelMu.Unlock()

	h.sessionMgr.Delete(sessionID)
	h.contextMgr.Delete(sessionID)
	log.Printf("[ws] session %s cleaned up", sessionID)
}

func (h *Handler) sendError(client *Client, sessionID, code, message string) {
	client.SendJSON(event.RealtimeEvent{
		Type:      event.TypeError,
		SessionID: sessionID,
		Code:      code,
		Message:   message,
	})
	log.Printf("[ws] error session=%s code=%s msg=%s", sessionID, code, message)
}

// isMeaningful checks if a transcript is worth answering.
func isMeaningful(text string) bool {
	trimmed := strings.TrimSpace(text)
	if len(trimmed) < 5 {
		return false
	}

	lower := strings.ToLower(trimmed)
	fillers := []string{"uh", "uhm", "hmm", "um", "hello", "hi", "hey", "wait", "ok", "okay", "i mean"}
	for _, f := range fillers {
		if lower == f {
			return false
		}
	}

	return true
}
