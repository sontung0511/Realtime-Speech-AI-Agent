package notes

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sontungAI/realtime-speech-ai-agent/pkg/model"
)

// fillerWords that should not produce notes.
var fillerWords = map[string]bool{
	"uh": true, "uhm": true, "hmm": true, "um": true,
	"hello": true, "hi": true, "hey": true,
	"wait": true, "ok": true, "okay": true,
}

var noteCounter atomic.Int64

// Extract decides if a transcript is worth noting and returns a Note if so.
func Extract(sessionID, text, summary string) *model.Note {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)

	// Skip very short or filler text
	if len(trimmed) < 10 {
		return nil
	}

	// Skip filler
	if fillerWords[lower] {
		return nil
	}

	// Skip questions (agent should answer, not note)
	if strings.HasSuffix(trimmed, "?") {
		return nil
	}

	// Check for noteworthy patterns
	if !isNoteworthy(lower) {
		return nil
	}

	id := noteCounter.Add(1)
	return &model.Note{
		ID:         fmt.Sprintf("%s_note_%d", sessionID, id),
		SessionID:  sessionID,
		Text:       trimmed,
		SourceText: text,
		CreatedAt:  time.Now(),
	}
}

// isNoteworthy checks if text contains patterns worth noting.
func isNoteworthy(lower string) bool {
	keywords := []string{
		"want", "need", "should", "must", "use", "prefer",
		"build", "create", "implement", "design", "plan",
		"decision", "requirement", "important", "goal",
		"deadline", "task", "action", "remember",
	}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
