<!-- docs/SKILL.md -->

# SKILL.md

## Skill Overview

This document defines the internal skills/modules required for the Realtime Speech AI Agent.

Each skill should be implemented as a separate module, package, or service so the system can be tested, replaced, and scaled independently.

The Golang backend should orchestrate these skills.

Heavy AI processing should be handled by external services or separate workers.

---

## Core Skills

The agent requires these skills:

```text
1. Audio Stream Skill
2. Voice Activity Detection Skill
3. Speech To Text Skill
4. Language Detection Skill
5. Session Management Skill
6. Context Memory Skill
7. Answer Generation Skill
8. Note Taking Skill
9. Text To Speech Skill
10. Interruption Skill
```

---

## 1. Audio Stream Skill

### Purpose

Receive audio chunks from the client in real time and forward them to the speech processing pipeline.

### Responsibilities

- Accept WebSocket audio chunks
- Validate message type
- Validate session ID
- Validate audio format
- Validate sample rate
- Decode base64 audio data
- Forward audio to STT provider
- Handle reconnect
- Handle broken stream
- Prevent oversized messages

### Input

```json
{
  "type": "audio_chunk",
  "session_id": "session_001",
  "format": "pcm16",
  "sample_rate": 16000,
  "data": "base64_audio_data"
}
```

### Output

Internal audio frame:

```json
{
  "session_id": "session_001",
  "format": "pcm16",
  "sample_rate": 16000,
  "duration_ms": 500,
  "payload": "raw_audio_bytes"
}
```

### MVP Audio Format

```text
format: pcm16
sample_rate: 16000
channel: mono
chunk_duration: 500ms - 2000ms
```

### Error Cases

```text
invalid_audio_format
invalid_sample_rate
empty_audio_payload
audio_payload_too_large
session_not_found
```

---

## 2. Voice Activity Detection Skill

### Purpose

Detect whether the user is speaking or silent.

This skill helps decide when the AI should answer.

### Responsibilities

- Detect speech start
- Detect speech end
- Detect silence duration
- Prevent answering too early
- Trigger interruption when user speaks while AI is speaking

### Output Events

Speech started:

```json
{
  "type": "speech_started",
  "session_id": "session_001"
}
```

Speech ended:

```json
{
  "type": "speech_ended",
  "session_id": "session_001",
  "silence_ms": 900
}
```

### Answer Trigger Rule

The system may consider answering when silence duration is:

```text
700ms - 1500ms
```

### MVP Note

VAD can be skipped in the earliest MVP if the STT provider already returns final transcript segments.

---

## 3. Speech To Text Skill

### Purpose

Convert speech audio into text in real time.

### Responsibilities

- Stream audio to STT provider
- Receive partial transcripts
- Receive final transcripts
- Return language if provider supports it
- Return timestamps if provider supports them
- Handle STT provider failure
- Support mock STT for local testing

### Possible Providers

```text
OpenAI Realtime API
Whisper
Faster Whisper
Deepgram
Google Speech To Text
Local STT model
```

### Partial Transcript Output

```json
{
  "type": "partial_text",
  "session_id": "session_001",
  "lang": "en",
  "text": "I want to build",
  "start_ms": 0,
  "end_ms": 1200
}
```

### Final Transcript Output

```json
{
  "type": "final_text",
  "session_id": "session_001",
  "lang": "en",
  "text": "I want to build a realtime speech AI agent.",
  "start_ms": 0,
  "end_ms": 2800
}
```

### Rules

- Partial transcript should be sent to UI as subtitle preview
- Final transcript should update context
- Final transcript should be saved
- Only final transcript should trigger LLM answer generation

---

## 4. Language Detection Skill

### Purpose

Detect the language of the latest speech/text.

### Responsibilities

- Detect language from STT result
- Detect language from text fallback
- Update session language
- Help LLM choose response language
- Fallback to previous session language when confidence is low

### Input

```json
{
  "text": "Tôi muốn xây dựng AI nghe và trả lời realtime"
}
```

### Output

```json
{
  "language": "vi",
  "confidence": 0.98
}
```

### Language Fallback Rule

```text
if confidence >= 0.75:
    use detected language
else if session.language exists:
    use session.language
else:
    use "en"
```

---

## 5. Session Management Skill

### Purpose

Manage realtime conversation sessions.

### Responsibilities

- Create session
- Stop session
- Track session status
- Store session metadata
- Track WebSocket connection
- Restore session after reconnect if possible
- Clean up expired sessions
- Cancel running tasks when session stops

### Session Status

```text
idle
listening
transcribing
thinking
answering
speaking
error
```

### Session Example

```json
{
  "session_id": "session_001",
  "status": "listening",
  "language": "en",
  "enable_tts": false,
  "created_at": "2026-01-01T10:00:00Z",
  "updated_at": "2026-01-01T10:01:00Z"
}
```

### Rules

- Every event must belong to a valid session
- Session ID must be unique
- Expired sessions must be cleaned up
- Disconnected sessions should release resources

---

## 6. Context Memory Skill

### Purpose

Maintain short-term memory for each realtime session.

### Stored Data

- Session ID
- Current language
- Latest transcript
- Rolling summary
- Important notes
- Last answer
- Last intent
- Agent status

### Example State

```json
{
  "session_id": "session_001",
  "language": "en",
  "summary": "User is building a realtime speech AI agent using Golang backend.",
  "last_transcript": "I need it to answer immediately from speech.",
  "last_answer": "Use a streaming pipeline with STT, context, LLM, and TTS.",
  "intent": "technical_design"
}
```

### Responsibilities

- Store context per session
- Update rolling summary
- Keep context short
- Provide compact context to LLM
- Avoid sending full transcript every time

### Rolling Summary Rule

```text
previous_summary + latest_final_transcript -> new_summary
```

### Context Size Rule

Keep the LLM prompt compact.

Recommended:

```text
summary: 500 - 1500 characters
notes: latest 5 - 20 useful notes
latest_text: latest final transcript
```

---

## 7. Answer Generation Skill

### Purpose

Generate a realtime AI answer using the latest final transcript and session context.

### Responsibilities

- Build compact LLM prompt
- Use latest final transcript
- Use rolling summary
- Use important notes
- Match response language
- Stream answer tokens
- Keep response short
- Save final answer

### Input

```json
{
  "session_id": "session_001",
  "language": "en",
  "summary": "User is building a realtime speech AI agent.",
  "latest_text": "How should I build the backend?",
  "notes": [
    "User prefers Golang backend.",
    "System should use WebSocket streaming."
  ]
}
```

### Output Delta

```json
{
  "type": "answer_delta",
  "session_id": "session_001",
  "lang": "en",
  "text": "Use Golang"
}
```

### Output Completed

```json
{
  "type": "answer_completed",
  "session_id": "session_001",
  "lang": "en",
  "text": "Use Golang as the WebSocket orchestrator, and keep STT, LLM, and TTS as separate services."
}
```

### Rules

- Do not generate answer from partial transcript
- Generate answer from final transcript only
- Keep realtime answer concise
- Use same language as user
- Do not hallucinate missing context
- If not enough information, ask a short clarification

---

## 8. Note Taking Skill

### Purpose

Extract important notes from final transcripts.

### Responsibilities

- Detect useful facts
- Save user goals
- Save technical requirements
- Save decisions
- Save action items
- Save story events
- Avoid duplicate notes
- Avoid saving low-value filler

### Input

```json
{
  "session_id": "session_001",
  "text": "I want the backend to be written in Golang.",
  "summary": "User is designing a realtime speech AI agent."
}
```

### Output

```json
{
  "type": "note_created",
  "session_id": "session_001",
  "text": "User wants the backend implemented in Golang.",
  "source_text": "I want the backend to be written in Golang."
}
```

### Good Notes

```text
requirements
decisions
preferences
tasks
deadlines
names
technical constraints
story events
meeting action items
```

### Bad Notes

```text
filler words
partial incomplete text
repeated content
low-value phrases
STT noise
```

---

## 9. Text To Speech Skill

### Purpose

Convert AI answer text into audio.

### Responsibilities

- Convert final answer text to speech
- Stream TTS audio back to client
- Match language if possible
- Support cancellation
- Stop speaking when user interrupts

### Input

```json
{
  "session_id": "session_001",
  "text": "Use Golang as the realtime WebSocket server.",
  "language": "en"
}
```

### Output

```json
{
  "type": "tts_audio",
  "session_id": "session_001",
  "format": "pcm16",
  "sample_rate": 24000,
  "data": "base64_audio_data"
}
```

### Rule

TTS is optional for MVP.

Text answer must work first before voice output.

---

## 10. Interruption Skill

### Purpose

Stop current AI answer or TTS when the user starts speaking again.

### Trigger

```json
{
  "type": "speech_started",
  "session_id": "session_001"
}
```

### Responsibilities

- Cancel current LLM stream
- Stop current TTS stream
- Change agent status to listening
- Prevent AI from talking over the user

### Rule

User speech has higher priority than AI speech.

---

## Skill Execution Flow

```text
Audio Stream Skill
  -> Voice Activity Detection Skill
  -> Speech To Text Skill
  -> Language Detection Skill
  -> Session Management Skill
  -> Context Memory Skill
  -> Note Taking Skill
  -> Answer Generation Skill
  -> Text To Speech Skill
  -> Interruption Skill
```

---

## MVP Skill Priority

Build in this order:

```text
1. Session Management Skill
2. WebSocket Audio Stream Skill
3. Mock STT Skill
4. Mock LLM Answer Skill
5. Realtime Event Flow
6. Real STT Skill
7. Context Memory Skill
8. Real LLM Answer Skill
9. Language Detection Skill
10. Note Taking Skill
11. Text To Speech Skill
12. Voice Activity Detection Skill
13. Interruption Skill
```

---

## Testing Strategy

Each skill should have unit tests.

Required test cases:

- Start session success
- Stop session success
- Invalid session ID
- Audio chunk validation
- Invalid audio format
- Partial transcript handling
- Final transcript handling
- Language detection fallback
- Context summary update
- Answer trigger logic
- Note extraction
- Duplicate note prevention
- LLM timeout
- STT failure
- TTS cancellation
- Session cleanup
- Interrupt while answering
- Interrupt while speaking
