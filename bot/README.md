# Google Meet Bot

Bot tham gia Google Meet như một participant, capture audio từ meeting và stream về AI agent server để xử lý realtime.

## Architecture

```
Google Meet
  → Bot (Puppeteer/Chrome) joins as participant
  → Captures WebRTC audio from meeting
  → Converts to PCM16 base64 chunks
  → Sends via WebSocket to AI Agent Server
  → Server processes: STT → Language Detect → LLM → Answer
```

## Prerequisites

- Node.js 18+
- Google Chrome or Chromium
- AI Agent server running (`cd server && go run cmd/server/main.go`)

## Setup

```bash
cd bot
npm install
cp .env.example .env
# Edit .env with your meeting URL
```

## Usage

### Basic (anonymous join - meeting must allow)

```bash
MEET_URL=https://meet.google.com/xxx-yyy-zzz npm start
```

### With Google account

```bash
MEET_URL=https://meet.google.com/xxx-yyy-zzz \
GOOGLE_EMAIL=your-email@gmail.com \
GOOGLE_PASSWORD=your-password \
npm start
```

### Non-headless (see the browser)

```bash
MEET_URL=https://meet.google.com/xxx-yyy-zzz \
HEADLESS=false \
npm start
```

## How It Works

1. **Launch** - Puppeteer opens Chrome with media permissions pre-granted
2. **Sign in** - (Optional) Signs into Google account
3. **Join** - Navigates to Meet URL, disables cam/mic, clicks "Join now"
4. **Wait** - If "Ask to join", waits for host to admit
5. **Capture** - Hooks into WebRTC audio tracks from other participants
6. **Stream** - Converts audio to PCM16 → base64 → sends to AI server
7. **Listen** - Server processes audio and returns transcripts + AI answers

## Configuration

| Variable          | Default                         | Description             |
| ----------------- | ------------------------------- | ----------------------- |
| MEET_URL          | (required)                      | Google Meet link        |
| BOT_NAME          | AI Assistant                    | Display name in meeting |
| WS_URL            | ws://localhost:8080/ws/realtime | AI server WebSocket     |
| GOOGLE_EMAIL      |                                 | Google account email    |
| GOOGLE_PASSWORD   |                                 | Google account password |
| HEADLESS          | true                            | Run browser headless    |
| SAMPLE_RATE       | 16000                           | Audio sample rate       |
| CHUNK_DURATION_MS | 1000                            | Audio chunk duration    |

## Notes

- Bot joins with camera and microphone **OFF** (listen-only mode)
- If meeting requires authentication, provide Google credentials
- For meetings with "Ask to join", host must admit the bot
- Audio capture uses WebRTC track interception
- Set `HEADLESS=false` for debugging the join flow
- Screenshots saved on error: `bot-join-error.png`

## Troubleshooting

**Bot can't join meeting:**

- Check if meeting allows anonymous participants
- Use Google credentials if required
- Run with `HEADLESS=false` to see what's happening
- Check `bot-join-error.png` for visual state

**No audio captured:**

- Ensure other participants are speaking
- WebRTC track interception requires participants to be active
- Check browser console logs (run non-headless)

**Connection refused to AI server:**

- Make sure server is running: `cd server && go run cmd/server/main.go`
- Check WS_URL is correct
