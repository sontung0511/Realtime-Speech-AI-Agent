package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sontungAI/realtime-speech-ai-agent/internal/config"
	ctxmgr "github.com/sontungAI/realtime-speech-ai-agent/internal/context"
	"github.com/sontungAI/realtime-speech-ai-agent/internal/language"
	"github.com/sontungAI/realtime-speech-ai-agent/internal/llm"
	"github.com/sontungAI/realtime-speech-ai-agent/internal/session"
	"github.com/sontungAI/realtime-speech-ai-agent/internal/storage"
	"github.com/sontungAI/realtime-speech-ai-agent/internal/stt"
	ws "github.com/sontungAI/realtime-speech-ai-agent/internal/websocket"
)

func main() {
	cfg := config.Load()

	log.Println("=== Realtime Speech AI Agent ===")
	log.Printf("Port: %d", cfg.Port)
	log.Printf("STT Provider: %s", cfg.STTProvider)
	log.Printf("LLM Provider: %s", cfg.LLMProvider)

	repo, err := storage.NewRepository(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer repo.Close()

	var sttProvider stt.Provider
	sttProvider = stt.NewMockProvider()
	log.Println("Using Mock STT provider")

	var llmProvider llm.Provider
	llmProvider = llm.NewMockProvider()
	log.Println("Using Mock LLM provider")

	sessionMgr := session.NewManager()
	contextMgr := ctxmgr.NewManager()

	langDetector := language.NewDetector(llmProvider)
	log.Println("Using AI-based language detection")

	wsHandler := ws.NewHandler(
		sessionMgr,
		contextMgr,
		sttProvider,
		llmProvider,
		langDetector,
		repo,
		cfg.MaxAudioPayload,
	)

	http.Handle("/ws/realtime", wsHandler)
	http.Handle("/", http.FileServer(http.Dir("../client")))

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Server starting on %s", addr)
	log.Printf("WebSocket endpoint: ws://localhost:%d/ws/realtime", cfg.Port)
	log.Printf("Client UI: http://localhost:%d", cfg.Port)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
