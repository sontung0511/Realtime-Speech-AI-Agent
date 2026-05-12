<!-- docs/SPEC.md -->

# SPEC.md

## Project Name

Realtime Speech AI Agent

## Objective

Build a realtime AI agent that listens to speech/audio, converts speech to text, detects language, understands session context, creates useful notes, and generates low-latency answers.

The first version must focus on a stable realtime pipeline.

The system should not start with complex long-term memory or advanced story intelligence.

---

## High Level Architecture

```text
Client
  -> WebSocket
  -> Golang Backend
  -> STT Service
  -> Context Engine
  -> LLM Service
  -> Optional TTS Service
  -> Client
```

---

## Component Responsibilities

## 1. Client

The client is responsible for:

- Capturing microphone audio
- Encoding audio into supported format
- Sending audio chunks through WebSocket
- Sending session control events
- Displaying partial transcript
- Displaying final transcript
- Displaying AI answer stream
- Playing TTS audio if enabled
- Sending interrupt event if user speaks while AI is answering

---

## 2. Golang Backend

The Golang backend is the realtime orchestrator.

It is responsible for:

- WebSocket connection management
- Message parsing
- Message validation
- Session lifecycle management
- Audio chunk validation
- Forwarding audio to STT service
- Receiving partial/final transcript
- Sending transcript events to client
- Updating session context
- Extracting notes
- Calling LLM service
- Streaming answer back to client
- Calling TTS service if enabled
- Handling interruption
- Saving transcripts, notes, and answers
- Handling errors and cleanup

The backend should not run heavy AI models directly in the main process for MVP.

---

## 3. STT Service

The STT service is responsible for:

- Receiving realtime audio stream
- Converting speech to text
- Returning partial transcript
- Returning final transcript
- Detecting language if supported
- Returning timestamps if supported

Possible providers:

```text
OpenAI Realtime API
Whisper
Faster Whisper
Deepgram
Google Speech To Text
Local STT model
```

---

## 4. LLM Service

The LLM service is responsible for:

- Receiving compact session context
- Receiving latest final transcript
- Generating short context-aware answer
- Streaming answer tokens
- Matching response language
- Avoiding long answers in realtime mode

---

## 5. TTS Service

The TTS service is responsible for:

- Converting text answer into audio
- Streaming audio back to client
- Supporting cancellation
- Matching language voice if possible

TTS is optional for MVP.

---

## Realtime Requirements

### Latency Target

| Stage                      |         Target |
| -------------------------- | -------------: |
| Client audio capture       |        < 100ms |
| Audio chunk transfer       |        < 100ms |
| STT partial result         |        < 500ms |
| Final transcript detection | 700ms - 1500ms |
| LLM first token            |      < 1s - 2s |
| TTS first audio            |           < 1s |

### Audio Format

MVP recommended format:

```text
format: pcm16
sample_rate: 16000
channel: mono
chunk_duration: 500ms - 2000ms
```

---

## WebSocket API

### Endpoint

```text
/ws/realtime
```

### Protocol

MVP uses JSON messages with base64 audio data.

Future optimization can support binary audio frames.

---

## Client To Server Events

## 1. start_session

Start a realtime speech session.

```json
{
  "type": "start_session",
  "session_id": "session_001",
  "language": "auto",
  "enable_tts": false
}
```

### Fields

| Field      | Type   | Required | Description              |
| ---------- | ------ | -------: | ------------------------ |
| type       | string |      yes | Must be `start_session`  |
| session_id | string |      yes | Unique session ID        |
| language   | string |       no | `auto`, `en`, `vi`, etc. |
| enable_tts | bool   |       no | Enable voice response    |

---

## 2. audio_chunk

Send one audio chunk.

```json
{
  "type": "audio_chunk",
  "session_id": "session_001",
  "format": "pcm16",
  "sample_rate": 16000,
  "data": "base64_audio_data"
}
```

### Fields

| Field       | Type   | Required | Description           |
| ----------- | ------ | -------: | --------------------- |
| type        | string |      yes | Must be `audio_chunk` |
| session_id  | string |      yes | Session ID            |
| format      | string |      yes | Audio format          |
| sample_rate | number |      yes | Audio sample rate     |
| data        | string |      yes | Base64 audio payload  |

---

## 3. stop_session

Stop the current realtime session.

```json
{
  "type": "stop_session",
  "session_id": "session_001"
}
```

---

## 4. interrupt

Stop current answer/TTS and return to listening state.

```json
{
  "type": "interrupt",
  "session_id": "session_001"
}
```

---

## Server To Client Events

## 1. session_started

```json
{
  "type": "session_started",
  "session_id": "session_001",
  "status": "listening"
}
```

---

## 2. partial_text

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

---

## 3. final_text

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

---

## 4. answer_delta

```json
{
  "type": "answer_delta",
  "session_id": "session_001",
  "lang": "en",
  "text": "Use Golang"
}
```

---

## 5. answer_completed

```json
{
  "type": "answer_completed",
  "session_id": "session_001",
  "lang": "en",
  "text": "Use Golang as the WebSocket orchestrator, and keep STT, LLM, and TTS as separate services."
}
```

---

## 6. note_created

```json
{
  "type": "note_created",
  "session_id": "session_001",
  "text": "User wants Golang backend for realtime WebSocket orchestration.",
  "source_text": "I want to use Golang for the backend."
}
```

---

## 7. tts_audio

```json
{
  "type": "tts_audio",
  "session_id": "session_001",
  "format": "pcm16",
  "sample_rate": 24000,
  "data": "base64_audio_data"
}
```

---

## 8. error

```json
{
  "type": "error",
  "session_id": "session_001",
  "code": "stt_failed",
  "message": "Speech to text service failed"
}
```

---

## Golang Data Models

### Session

```go
type Session struct {
    ID        string
    Status    string
    Language  string
    EnableTTS bool
    Summary   string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

---

### TranscriptSegment

```go
type TranscriptSegment struct {
    ID        string
    SessionID string
    StartMS   int64
    EndMS     int64
    Language  string
    Text      string
    IsFinal   bool
    Speaker   string
    CreatedAt time.Time
}
```

---

### Note

```go
type Note struct {
    ID         string
    SessionID  string
    Text       string
    SourceText string
    CreatedAt  time.Time
}
```

---

### AIAnswer

```go
type AIAnswer struct {
    ID        string
    SessionID string
    Language  string
    Text      string
    CreatedAt time.Time
}
```

---

### RealtimeEvent

```go
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
}
```

---

## Database Schema

### sessions

```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    status TEXT NOT NULL,
    language TEXT,
    enable_tts BOOLEAN DEFAULT FALSE,
    summary TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

---

### transcript_segments

```sql
CREATE TABLE transcript_segments (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    start_ms BIGINT,
    end_ms BIGINT,
    language TEXT,
    text TEXT NOT NULL,
    is_final BOOLEAN NOT NULL,
    speaker TEXT,
    created_at TIMESTAMP NOT NULL
);
```

---

### notes

```sql
CREATE TABLE notes (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    text TEXT NOT NULL,
    source_text TEXT,
    created_at TIMESTAMP NOT NULL
);
```

---

### ai_answers

```sql
CREATE TABLE ai_answers (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    language TEXT,
    text TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);
```

---

## Suggested Backend Structure

```text
cmd/
  server/
    main.go

internal/
  config/
    config.go

  websocket/
    handler.go
    client.go
    message.go

  session/
    manager.go
    model.go

  audio/
    validator.go
    stream.go

  stt/
    client.go
    provider.go
    mock.go

  language/
    detector.go

  context/
    manager.go
    summary.go

  llm/
    client.go
    prompt.go
    mock.go

  notes/
    extractor.go

  tts/
    client.go

  storage/
    postgres.go
    repository.go

pkg/
  event/
    event.go

  model/
    model.go

web/
  index.html
  app.js

docs/
  AGENT.md
  SKILL.md
  SPEC.md
```

---

## Realtime Flow

### Normal Flow

```text
1. Client connects to /ws/realtime
2. Client sends start_session
3. Backend creates session
4. Backend returns session_started
5. Client sends audio_chunk continuously
6. Backend validates audio_chunk
7. Backend forwards audio to STT service
8. STT returns partial_text
9. Backend sends partial_text to client
10. STT returns final_text
11. Backend saves transcript
12. Backend updates session context
13. Backend extracts useful notes
14. Backend calls LLM
15. LLM streams answer_delta
16. Backend sends answer_delta to client
17. LLM completes
18. Backend sends answer_completed
19. Backend saves AI answer
20. If TTS enabled, backend calls TTS
21. Backend streams tts_audio to client
```

---

## Interrupt Flow

```text
1. AI is answering or speaking
2. User starts speaking again
3. Client sends interrupt or VAD detects speech_started
4. Backend cancels current LLM/TTS stream
5. Backend changes session status to listening
6. Backend processes new audio
```

---

## Answer Trigger Logic

The backend should call LLM only when:

```text
- final_text event is received
- text is meaningful
- session is active
- user intent requires an answer
```

Pseudo logic:

```text
if event.type == "final_text":
    saveTranscript(event)
    updateLanguage(event.lang)
    updateContext(event.text)
    extractNotes(event.text)

    if shouldAnswer(event.text, context):
        streamAnswer(context, event.text)
```

Do not call LLM for every partial transcript.

---

## Meaningful Text Rules

A text is meaningful when:

```text
- length > minimum threshold
- not only filler words
- not only background noise
- contains a question, command, decision, requirement, or useful statement
```

Examples that should not trigger answer:

```text
uh
hmm
hello...
wait...
I mean...
```

Examples that should trigger answer:

```text
How should I build the backend?
I want the system to use Golang.
Can you explain the architecture?
What is the next step?
```

---

## LLM Prompt Format

Use compact prompt.

Example:

```text
You are a realtime speech AI assistant.

Language:
en

Conversation summary:
User is building a realtime speech AI agent using Golang backend.

Important notes:
- Backend should use WebSocket.
- STT, LLM, and TTS should be separate services.

Latest user message:
How should I build the backend?

Instruction:
Answer shortly, clearly, and in the same language as the user.
Avoid long explanations unless the user asks for details.
```

---

## Error Codes

```text
invalid_message
invalid_json
invalid_session_id
session_not_found
session_already_exists
invalid_audio_format
invalid_sample_rate
audio_payload_too_large
stt_failed
llm_failed
tts_failed
storage_failed
timeout
internal_error
```

### Error Example

```json
{
  "type": "error",
  "session_id": "session_001",
  "code": "invalid_audio_format",
  "message": "Only pcm16 16000Hz mono is supported in MVP"
}
```

---

## MVP Scope

MVP must include:

- Golang WebSocket server
- Start session
- Stop session
- Receive audio chunks
- Validate audio chunks
- Mock STT provider
- Mock LLM provider
- Real STT integration
- Real LLM integration
- Partial transcript event
- Final transcript event
- Basic language detection
- Basic context memory
- Basic note extraction
- Stream AI answer to client
- Save transcript history

MVP can skip:

- TTS
- Speaker diarization
- Long-term memory
- Emotion detection
- Voice cloning
- Full offline mode
- Mobile app

---

## Non Functional Requirements

### Performance

- Must handle multiple concurrent WebSocket sessions
- Must not block audio receiving while waiting for LLM
- Must use goroutines safely
- Must use context cancellation
- Must use timeout for external services
- Must avoid memory leak when WebSocket disconnects

### Reliability

- Handle invalid JSON
- Handle invalid audio format
- Handle WebSocket disconnect
- Handle STT provider failure
- Handle LLM provider timeout
- Handle storage failure
- Clean up expired sessions

### Security

Production should support:

- Authentication
- Message size limit
- Rate limiting
- Session duration limit
- Secure transcript storage
- No sensitive audio logging
- Input validation

---

## Testing Requirements

### Unit Tests

Required unit tests:

```text
start session success
start duplicate session
stop session success
invalid session ID
parse valid message
parse invalid JSON
audio chunk validation
invalid audio format
invalid sample rate
partial transcript handling
final transcript handling
language detection fallback
context update
answer trigger logic
note extraction
duplicate note prevention
LLM timeout
STT failure
session cleanup
interrupt while answering
```

### Integration Tests

Required integration tests:

```text
client connects websocket
client sends start_session
backend returns session_started
client sends audio_chunk
mock STT returns partial_text
mock STT returns final_text
mock LLM streams answer_delta
backend returns answer_completed
backend saves transcript
backend handles interrupt
backend handles stop_session
```

### Load Test

Basic load test:

```text
100 concurrent sessions
audio chunk every 500ms
duration: 5 minutes
```

Metrics to observe:

```text
WebSocket memory usage
goroutine count
STT latency
LLM latency
event processing latency
error rate
CPU usage
RAM usage
```

---

## Implementation Priority

Recommended order:

```text
1. Project structure
2. Config loader
3. WebSocket server
4. Session manager
5. Message schema
6. Audio validator
7. Mock STT service
8. Mock LLM service
9. End-to-end realtime event flow
10. Storage repository
11. Real STT integration
12. Real LLM integration
13. Context memory
14. Note extraction
15. Language detection
16. TTS integration
17. Interrupt handling
18. Load test
```

---

## Final Design Rule

Keep Golang as the realtime orchestrator.

Do not put heavy AI model execution directly inside the main Golang backend process for MVP.

Use external AI services first, then optimize later only after the realtime pipeline is stable.
