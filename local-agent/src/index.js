require("dotenv").config();

const AudioCapture = require("./audio-capture");
const AgentClient = require("./agent-client");
const UIServer = require("./ui-server");

// Config
const AGENT_WS_URL = process.env.AGENT_WS_URL || "ws://localhost:8080/ws/realtime";
const UI_PORT = parseInt(process.env.UI_PORT || "3000", 10);
const AUDIO_SOURCE = process.env.AUDIO_SOURCE || "auto";
const SAMPLE_RATE = parseInt(process.env.SAMPLE_RATE || "16000", 10);
const CHUNK_MS = parseInt(process.env.CHUNK_MS || "1000", 10);

async function main() {
    console.log("=== Realtime Speech AI - Local Agent ===");
    console.log("Agent server:", AGENT_WS_URL);
    console.log("Overlay UI port:", UI_PORT);
    console.log("Audio source:", AUDIO_SOURCE);
    console.log("");

    // 1. Start overlay UI server
    const ui = new UIServer(UI_PORT);
    await ui.start();

    // 2. Connect to AI agent server
    const agent = new AgentClient({
        wsUrl: AGENT_WS_URL,
        sessionId: "local_" + Date.now(),
    });

    // 3. Setup audio capture
    const audio = new AudioCapture({
        source: AUDIO_SOURCE,
        sampleRate: SAMPLE_RATE,
        chunkDurationMs: CHUNK_MS,
    });

    let chunkCount = 0;

    // Wire: audio -> agent server
    audio.on("audio", (base64Data) => {
        chunkCount++;
        agent.sendAudioChunk(base64Data);
        // Periodically notify UI of capture status
        if (chunkCount % 5 === 0) {
            ui.broadcast({ type: "capture_status", active: true, chunks: chunkCount });
        }
    });

    audio.on("error", (err) => {
        ui.broadcast({ type: "error", code: "AUDIO_ERROR", message: err.message });
    });

    audio.on("stop", () => {
        ui.broadcast({ type: "capture_status", active: false, chunks: chunkCount });
    });

    // Wire: agent events -> overlay UI
    agent.on("event", (evt) => {
        ui.broadcast(evt);
    });

    // Handle overlay commands (start/stop capture)
    ui.onCommand = (msg) => {
        switch (msg.type) {
            case "start_capture":
                if (!audio.isCapturing) {
                    chunkCount = 0;
                    audio.start();
                    ui.broadcast({ type: "capture_status", active: true, chunks: 0 });
                }
                break;
            case "stop_capture":
                audio.stop();
                ui.broadcast({ type: "capture_status", active: false, chunks: chunkCount });
                break;
        }
    };

    // 4. Connect to agent and start capture
    try {
        await agent.connect();
        console.log("[main] Connected to AI agent server");
    } catch (err) {
        console.error("[main] Failed to connect to AI agent server:", err.message);
        console.error("[main] Make sure the server is running at", AGENT_WS_URL);
        console.log("[main] UI is still available - will auto-reconnect when server is up");
    }

    // Auto-start capture
    try {
        audio.start();
        console.log("[main] Audio capture started");
        ui.broadcast({ type: "capture_status", active: true, chunks: 0 });
    } catch (err) {
        console.error("[main] Failed to start audio capture:", err.message);
        ui.broadcast({ type: "error", code: "AUDIO_ERROR", message: err.message });
    }

    console.log("");
    console.log("Open overlay UI: http://localhost:" + UI_PORT);
    console.log("Press Ctrl+C to stop");

    // Graceful shutdown
    process.on("SIGINT", () => {
        console.log("\n[main] Shutting down...");
        audio.stop();
        agent.disconnect();
        ui.stop();
        process.exit(0);
    });

    process.on("SIGTERM", () => {
        audio.stop();
        agent.disconnect();
        ui.stop();
        process.exit(0);
    });
}

main().catch((err) => {
    console.error("Fatal error:", err);
    process.exit(1);
});
