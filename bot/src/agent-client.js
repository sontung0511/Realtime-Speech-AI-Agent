/**
 * WebSocket client that connects bot audio to the AI agent server
 */

const WebSocket = require("ws");

class AgentClient {
  constructor(config) {
    this.wsUrl = config.wsUrl || "ws://localhost:8080/ws/realtime";
    this.sessionId = config.sessionId || `meet_bot_${Date.now()}`;

    this.ws = null;
    this.isConnected = false;
    this.reconnectAttempts = 0;
    this.maxReconnects = 5;

    // Callbacks
    this.onTranscript = config.onTranscript || null;
    this.onAnswer = config.onAnswer || null;
    this.onNote = config.onNote || null;
    this.onError = config.onError || null;
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

  /**
   * Send audio chunk to server
   */
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

  /**
   * Send interrupt signal
   */
  interrupt() {
    this._send({
      type: "interrupt",
      session_id: this.sessionId,
    });
  }

  /**
   * Stop session and disconnect
   */
  disconnect() {
    if (this.ws) {
      this._send({
        type: "stop_session",
        session_id: this.sessionId,
      });

      setTimeout(() => {
        if (this.ws) {
          this.ws.close();
          this.ws = null;
        }
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
    switch (evt.type) {
      case "session_started":
        console.log(`[ws] Session started: ${evt.session_id}`);
        break;

      case "partial_text":
        // Don't log partials to avoid noise
        break;

      case "final_text":
        console.log(`[transcript] [${evt.lang}] ${evt.text}`);
        if (this.onTranscript) this.onTranscript(evt);
        break;

      case "answer_delta":
        // Streaming answer, don't log each delta
        break;

      case "answer_completed":
        console.log(`[ai-answer] ${evt.text}`);
        if (this.onAnswer) this.onAnswer(evt);
        break;

      case "note_created":
        console.log(`[note] 📝 ${evt.text}`);
        if (this.onNote) this.onNote(evt);
        break;

      case "error":
        console.error(`[ws-error] ${evt.code}: ${evt.message}`);
        if (this.onError) this.onError(evt);
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
    console.log(`[ws] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);

    setTimeout(() => {
      this.connect().catch((err) => {
        console.error("[ws] Reconnect failed:", err.message);
      });
    }, delay);
  }
}

module.exports = AgentClient;
