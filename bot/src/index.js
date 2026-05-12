/**
 * Realtime Speech AI Agent - Google Meet Bot
 *
 * This bot joins a Google Meet as a participant,
 * captures audio from the meeting, and streams it
 * to the AI agent server for processing.
 *
 * Usage:
 *   node src/index.js https://meet.google.com/xxx-yyy-zzz
 *   npm start -- https://meet.google.com/xxx-yyy-zzz
 *   MEET_URL=https://meet.google.com/xxx-yyy-zzz npm start
 */

require("dotenv").config();

const MeetBot = require("./meet-bot");
const AgentClient = require("./agent-client");

// --- Configuration ---
// Priority: CLI arg > env var > .env file
const meetUrlArg = process.argv[2];
const config = {
  meetUrl: meetUrlArg || process.env.MEET_URL,
  botName: process.env.BOT_NAME || "AI Assistant",
  wsUrl: process.env.WS_URL || "ws://localhost:8080/ws/realtime",
  googleEmail: process.env.GOOGLE_EMAIL || "",
  googlePassword: process.env.GOOGLE_PASSWORD || "",
  headless: process.env.HEADLESS !== "false",
  sampleRate: parseInt(process.env.SAMPLE_RATE || "16000", 10),
  chunkDurationMs: parseInt(process.env.CHUNK_DURATION_MS || "1000", 10),
};

// Validate
if (!config.meetUrl) {
  console.error("ERROR: Meet URL is required");
  console.error("");
  console.error("Usage:");
  console.error("  node src/index.js https://meet.google.com/xxx-yyy-zzz");
  console.error("  npm start -- https://meet.google.com/xxx-yyy-zzz");
  console.error("  MEET_URL=https://meet.google.com/xxx-yyy-zzz npm start");
  process.exit(1);
}

// Validate URL format
if (!config.meetUrl.match(/^https:\/\/meet\.google\.com\/[a-z]{3}-[a-z]{4}-[a-z]{3}/)) {
  console.error("ERROR: Invalid Google Meet URL format");
  console.error("Expected: https://meet.google.com/xxx-yyyy-zzz");
  process.exit(1);
}

// --- Stats ---
let audioChunksReceived = 0;
let audioChunksSent = 0;
let transcriptCount = 0;
let answerCount = 0;

// --- Main ---
async function main() {
  console.log("╔══════════════════════════════════════════╗");
  console.log("║   Realtime Speech AI - Google Meet Bot   ║");
  console.log("╚══════════════════════════════════════════╝");
  console.log("");
  console.log(`Meeting:  ${config.meetUrl}`);
  console.log(`Bot Name: ${config.botName}`);
  console.log(`Server:   ${config.wsUrl}`);
  console.log(`Headless: ${config.headless}`);
  console.log("");

  // 1. Connect to AI agent server
  const agent = new AgentClient({
    wsUrl: config.wsUrl,
    sessionId: `meet_${Date.now()}`,
    onTranscript: (evt) => {
      transcriptCount++;
    },
    onAnswer: (evt) => {
      answerCount++;
      console.log(`\n💬 AI: ${evt.text}\n`);
    },
    onNote: (evt) => {
      console.log(`📝 Note: ${evt.text}`);
    },
  });

  try {
    await agent.connect();
  } catch (err) {
    console.error("Failed to connect to AI agent server:", err.message);
    console.error("Make sure the server is running: cd server && go run cmd/server/main.go");
    process.exit(1);
  }

  // 2. Launch bot and join meeting
  const bot = new MeetBot({
    meetUrl: config.meetUrl,
    botName: config.botName,
    headless: config.headless,
    googleEmail: config.googleEmail,
    googlePassword: config.googlePassword,
  });

  try {
    await bot.join();
  } catch (err) {
    console.error("Failed to join meeting:", err.message);
    agent.disconnect();
    process.exit(1);
  }

  // 3. Start capturing audio and streaming to server
  await bot.startAudioCapture((base64Data) => {
    audioChunksReceived++;
    agent.sendAudioChunk(base64Data);
    audioChunksSent++;
  });

  // Print stats periodically
  const statsInterval = setInterval(() => {
    console.log(
      `[stats] chunks: ${audioChunksSent} | transcripts: ${transcriptCount} | answers: ${answerCount}`
    );
  }, 30000);

  // Handle graceful shutdown
  async function shutdown(signal) {
    console.log(`\n[${signal}] Shutting down...`);
    clearInterval(statsInterval);

    try {
      agent.disconnect();
      await bot.leave();
    } catch (_) {}

    console.log("[bot] Goodbye!");
    process.exit(0);
  }

  process.on("SIGINT", () => shutdown("SIGINT"));
  process.on("SIGTERM", () => shutdown("SIGTERM"));

  console.log("");
  console.log("✅ Bot is active! Listening to meeting audio...");
  console.log("   Press Ctrl+C to stop");
  console.log("");
}

main().catch((err) => {
  console.error("Fatal error:", err);
  process.exit(1);
});
