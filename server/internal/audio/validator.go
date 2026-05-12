package audio

import (
	"encoding/base64"
	"fmt"
)

const (
	FormatPCM16     = "pcm16"
	SampleRate16000 = 16000
	MaxPayloadSize  = 1048576 // 1MB
)

// ValidateChunk validates an audio chunk message fields.
func ValidateChunk(format string, sampleRate int, data string, maxPayload int) ([]byte, error) {
	if format != FormatPCM16 {
		return nil, fmt.Errorf("invalid_audio_format: only %s is supported, got %s", FormatPCM16, format)
	}

	if sampleRate != SampleRate16000 {
		return nil, fmt.Errorf("invalid_sample_rate: only %d is supported, got %d", SampleRate16000, sampleRate)
	}

	if data == "" {
		return nil, fmt.Errorf("empty_audio_payload")
	}

	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("invalid_audio_format: base64 decode failed: %w", err)
	}

	if maxPayload > 0 && len(raw) > maxPayload {
		return nil, fmt.Errorf("audio_payload_too_large: %d bytes exceeds max %d", len(raw), maxPayload)
	}

	return raw, nil
}
