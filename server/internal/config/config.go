package config

import (
	"os"
	"strconv"
)

// Config holds all application configuration.
type Config struct {
	Port               int
	STTProvider        string
	LLMProvider        string
	LLMAPIKey          string
	LLMModel           string
	DBPath             string
	MaxAudioPayload    int
	MaxSessionDuration int // seconds
}

// Load reads configuration from environment variables with defaults.
func Load() *Config {
	return &Config{
		Port:               getEnvInt("PORT", 8080),
		STTProvider:        getEnv("STT_PROVIDER", "mock"),
		LLMProvider:        getEnv("LLM_PROVIDER", "mock"),
		LLMAPIKey:          getEnv("LLM_API_KEY", ""),
		LLMModel:           getEnv("LLM_MODEL", "gpt-4o-mini"),
		DBPath:             getEnv("DB_PATH", "data/speech_agent.db"),
		MaxAudioPayload:    getEnvInt("MAX_AUDIO_PAYLOAD", 1048576), // 1MB
		MaxSessionDuration: getEnvInt("MAX_SESSION_DURATION", 3600),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
