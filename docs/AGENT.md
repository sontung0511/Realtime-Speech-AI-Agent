<!-- docs/AGENT.md -->

# AGENT.md

## Agent Name

Realtime Speech AI Agent

## Purpose

Realtime Speech AI Agent is an AI system that listens to real-time speech/audio, detects language, understands conversation context, takes useful notes, and generates real-time answers.

This agent is designed for live interaction, not offline audio file processing.

The main product behavior is:

```text
Listen -> Understand -> Answer
```

## Main Objectives

The agent must be able to:

- Listen to live microphone or audio stream
- Convert speech to text in real time
- Detect the spoken language automatically
- Understand the latest user message
- Maintain short-term context for the current session
- Decide when to answer and when to wait
- Generate short context-aware answers
- Save important information as notes
- Optionally convert answers into speech using TTS

## Agent Realtime Principle

The agent must prioritize:

```text
Low latency > Clean context > Advanced reasoning
```

The first version should focus on creating a stable realtime speech pipeline.

Advanced story understanding, emotion detection, and long-term memory can be added after the realtime core is stable.

---

## Agent Behavior

### 1. Listen Before Answering

The agent must not answer every partial word from speech-to-text.

Partial transcript should be used mainly for displaying subtitles.

The agent should answer only when the speech segment has enough meaning.

Good answer triggers:

- STT returns a final transcript segment
- User pauses for 700ms - 1500ms
- The sentence appears complete
- The intent is clear
- User asks a direct question

Bad answer triggers:

- Only one or two words are detected
- Transcript is still partial
- User is still speaking
- Context is ambiguous
- The message has no clear meaning yet

---

### 2. Same Language Response

The agent should answer in the same language as the user whenever possible.

Rules:

- If user speaks English, answer in English
- If user speaks Vietnamese, answer in Vietnamese
- If user mixes languages, use the dominant language of the latest final transcript
- If language confidence is low, use the previous session language
- If no previous language exists, default to English

---

### 3. Short Realtime Answers

Realtime answers should be short, useful, and direct.

The agent should avoid long explanations during live conversation unless the user explicitly asks for details.

Good realtime answer:

```text
Use Golang as the WebSocket orchestrator, then connect STT, LLM, and TTS as separate services.
```

Bad realtime answer:

```text
Here is a very long explanation with many unrelated technical details...
```

---

### 4. Do Not Over-Answer

The agent should not interrupt or dominate the conversation.

It should only answer when useful.

If the user is telling a story, the agent may only note the content instead of answering immediately.

If the user asks a question, the agent should answer.

---

### 5. Context Awareness

The agent must maintain context inside each session.

Context should include:

- Session ID
- Current language
- Latest final transcript
- Rolling summary
- Important notes
- Last user intent
- Last AI answer
- Current agent status

The agent must not send the full transcript to the LLM every time.

Use rolling summary instead:

```text
previous_summary + latest_final_transcript -> updated_summary
```

This keeps latency and token cost low.

---

## Agent Status

The agent can have these statuses:

```text
idle
listening
transcribing
thinking
answering
speaking
error
```

### Status Meaning

| Status       | Meaning                                       |
| ------------ | --------------------------------------------- |
| idle         | Session exists but no active speech           |
| listening    | Agent is listening to microphone/audio stream |
| transcribing | STT is processing audio                       |
| thinking     | LLM is generating an answer                   |
| answering    | Text answer is streaming to client            |
| speaking     | TTS audio is playing                          |
| error        | Something failed                              |

---

## Realtime Speech Rules

### Audio Chunk

Recommended audio chunk size:

```text
500ms - 2000ms
```

Recommended audio format for MVP:

```text
PCM16
16000 Hz
Mono channel
```

### Partial Transcript

Partial transcript is temporary.

Use partial transcript for:

- Subtitle preview
- Live UI display
- Debugging

Do not use partial transcript for:

- Final notes
- Final answer generation
- Long-term context update

### Final Transcript

Final transcript is stable.

Use final transcript for:

- Updating context
- Creating notes
- Calling LLM
- Generating answers
- Saving transcript history

---

## Answer Decision Logic

The agent should call LLM only when the latest transcript is meaningful.

Pseudo logic:

```text
if transcript.type == "final_text" and isMeaningful(transcript.text):
    updateContext(transcript.text)
    extractNotes(transcript.text)
    generateAnswerIfNeeded(transcript.text)
```

The agent should not call LLM for every partial transcript.

---

## Context Rules

### Rolling Summary

The rolling summary should be short.

Recommended summary size:

```text
500 - 1500 characters
```

The summary should contain only important context:

- User goal
- Current topic
- Important decisions
- Technical requirements
- Unresolved questions
- Story state if applicable

### Example Context

```json
{
  "session_id": "session_001",
  "language": "en",
  "status": "listening",
  "summary": "User is designing a realtime speech AI agent. They want Golang backend, WebSocket streaming, realtime STT, context-aware answers, and optional TTS.",
  "latest_transcript": "How should I build the backend?",
  "last_intent": "technical_question",
  "last_answer": "Use Golang as the realtime WebSocket orchestrator.",
  "notes": [
    "Backend should be written in Golang.",
    "Audio should be processed in realtime.",
    "Agent should answer from speech immediately."
  ]
}
```

---

## Note Taking Rules

The agent should save a note only when the information is useful.

Good notes:

- User goals
- Requirements
- Decisions
- Names
- Dates
- Tasks
- Action items
- Important story events
- Technical constraints
- User preferences

Bad notes:

- Filler words
- Repeated sentences
- Incomplete partial speech
- Low-value conversation
- Noise from STT

### Note Example

```json
{
  "type": "note_created",
  "session_id": "session_001",
  "text": "User wants to build a realtime speech AI agent with Golang backend.",
  "source_text": "I want to build this realtime speech AI agent using Go."
}
```

---

## Story Understanding Rules

For story or conversation sense, the agent should track:

- Main topic
- Timeline of events
- Important entities
- User goal
- Current unresolved question
- Speaker intent
- Emotional tone if available

MVP does not require advanced story intelligence.

MVP story context can be represented by a rolling summary.

---

## Interrupt Rules

The user has higher priority than the AI.

If the user starts speaking while AI is answering or speaking, the system should stop the current answer/TTS.

Expected behavior:

```text
User starts speaking
  -> stop current LLM stream
  -> stop current TTS playback
  -> change status to listening
  -> process new speech
```

This prevents the AI from talking over the user.

---

## Agent Input Events

The agent can receive these input events:

```text
start_session
audio_chunk
stop_session
interrupt
```

---

## Agent Output Events

The agent can emit these output events:

```text
session_started
partial_text
final_text
answer_delta
answer_completed
note_created
speech_started
speech_ended
tts_audio
error
```

---

## MVP Agent Scope

MVP must support:

- Realtime audio input
- WebSocket session
- Streaming speech-to-text
- Partial transcript event
- Final transcript event
- Basic language detection
- Short context memory
- LLM answer generation
- Basic note taking
- Transcript storage

MVP can skip:

- Voice cloning
- Full local offline AI
- Multi-agent orchestration
- Advanced emotion detection
- Advanced speaker diarization
- Long-term memory across many sessions
- Complex knowledge graph

---

## Final Rule

The agent should behave like a realtime assistant.

It should listen carefully, understand enough context, answer briefly, and avoid unnecessary reasoning or long responses during live speech.
