/**
 * WebSocket client that connects to the AI agent server
 * and forwards audio + receives transcripts/answers.
 * Also broadcasts events to the overlay UI via a local EventEmitter.
 */

const WebSocket = require("ws");
const { EventEmitter } = require("events");

class AgentClient extends EventEmitter {
  constructor(config) {
    super();
    this.wsUrl = config.wsUrl || "ws://localhost:8080/ws/realtime";
    this.sessionId = config.sessionId || `local_${Date.now()}`;

    this.ws = null;
    this.isConnected = false;
    this.reconnectAttempts = 0;
    this.maxReconnects = 10;
  }

  /**
   * Connect to the AI agent WebSocket server
   */
  connect() {
    return new Promise((resolve, reject) => {
      console.log(`[ws] Connecting to: ${this.wsUrl}`);

      this.ws = new WebSocket(this.wsUrl);

      this.ws.on("open", () => {
        console.log("[ws] Connected to AI agent server");
        this.isConnected = true;
        this.reconnectAttempts = 0;

        // Start session
        this._send({
          type: "start_session",
          session_id: this.sessionId,
          language: "auto",
          enable_tts: false,
        });

        resolve();
      });

      this.ws.on("message", (data) => {
        try {
          const evt = JSON.parse(data.toString());
          this._handleEvent(evt);
        } catch (err) {
          console.error("[ws] Parse error:", err);
        }
      });

      this.ws.on("close", () => {
        console.log("[ws] Disconnected");
        this.isConnected = false;
        this.emit("disconnected");
        this._tryReconnect();
      });

      this.ws.on("error", (err) => {
        console.error("[ws] Error:", err.message);
        if (!this.isConnected) {
          reject(err);
        }
      });
    });
  }

  sendAudioChunk(base64Data) {
    if (!this.isConnected) return;
    this._send({
      type: "audio_chunk",
      session_id: this.sessionId,
      format: "pcm16",
      sample_rate: 16000,
      data: base64Data,
    });
  }

  interrupt() {
    this._send({ type: "interrupt", session_id: this.sessionId });
  }

  disconnect() {
    if (this.ws) {
      this._send({ type: "stop_session", session_id: this.sessionId });
      setTimeout(() => {
        if (this.ws) { this.ws.close(); this.ws = null; }
      }, 500);
    }
    this.isConnected = false;
  }

  _send(evt) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(evt));
    }
  }

  _handleEvent(evt) {
    // Emit all events to UI
    this.emit("event", evt);

    switch (evt.type) {
      case "session_started":
        console.log(`[agent] Session started: ${evt.session_id}`);
        break;
      case "partial_text":
        this.emit("partial_text", evt);
        break;
      case "final_text":
        console.log(`[transcript] [${evt.lang}] ${evt.text}`);
        this.emit("final_text", evt);
        break;
      case "answer_delta":
        this.emit("answer_delta", evt);
        break;
      case "answer_completed":
        console.log(`[suggest] 💡 ${evt.text}`);
        this.emit("answer_completed", evt);
        break;
      case "note_created":
        console.log(`[note] 📝 ${evt.text}`);
        this.emit("note_created", evt);
        break;
      case "error":
        console.error(`[error] ${evt.code}: ${evt.message}`);
        this.emit("agent_error", evt);
        break;
    }
  }

  _tryReconnect() {
    if (this.reconnectAttempts >= this.maxReconnects) {
      console.error("[ws] Max reconnect attempts reached");
      return;
    }
    this.reconnectAttempts++;
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
    console.log(`[ws] Reconnecting in ${delay}ms...`);
    setTimeout(() => {
      this.connect().catch(() => {});
    }, delay);
  }
}

module.exports = AgentClient;
