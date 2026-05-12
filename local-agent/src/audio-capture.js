/**
 * System Audio Capture using PulseAudio/PipeWire on Linux
 *
 * Captures audio from monitor source (system output) using `parec` command.
 * This captures whatever audio is playing through your speakers/headphones,
 * including Google Meet audio.
 */

const { spawn, execSync } = require("child_process");
const { EventEmitter } = require("events");

class AudioCapture extends EventEmitter {
  constructor(config = {}) {
    super();
    this.sampleRate = config.sampleRate || 16000;
    this.chunkDurationMs = config.chunkDurationMs || 1000;
    this.source = config.source || "auto";
    this.process = null;
    this.isCapturing = false;
  }

  /**
   * Find the monitor source for system audio output.
   * Monitor sources capture what's being played (speaker/headphone output).
   */
  findMonitorSource() {
    try {
      const output = execSync("pactl list short sources", { encoding: "utf-8" });
      const lines = output.trim().split("\n");

      // Look for monitor sources (they capture output audio)
      const monitors = lines.filter((l) => l.includes(".monitor"));

      if (monitors.length === 0) {
        throw new Error("No monitor sources found. Is PulseAudio/PipeWire running?");
      }

      // Prefer the default sink's monitor
      const defaultSink = this._getDefaultSink();
      if (defaultSink) {
        const match = monitors.find((m) => m.includes(defaultSink));
        if (match) {
          const name = match.split("\t")[1];
          console.log(`[audio] Using monitor of default sink: ${name}`);
          return name;
        }
      }

      // Fallback to first monitor
      const name = monitors[0].split("\t")[1];
      console.log(`[audio] Using monitor source: ${name}`);
      return name;
    } catch (err) {
      throw new Error(`Failed to find audio source: ${err.message}`);
    }
  }

  /**
   * Get the default PulseAudio sink name
   */
  _getDefaultSink() {
    try {
      const output = execSync("pactl get-default-sink", { encoding: "utf-8" });
      return output.trim();
    } catch (_) {
      return null;
    }
  }

  /**
   * List all available audio sources
   */
  static listSources() {
    try {
      const output = execSync("pactl list short sources", { encoding: "utf-8" });
      console.log("Available audio sources:");
      console.log(output);
      return output.trim().split("\n").map((line) => {
        const parts = line.split("\t");
        return { id: parts[0], name: parts[1], driver: parts[2] };
      });
    } catch (err) {
      console.error("Failed to list sources:", err.message);
      return [];
    }
  }

  /**
   * Start capturing audio. Emits 'audio' events with base64 PCM16 data.
   */
  start() {
    if (this.isCapturing) return;

    // Resolve source
    let sourceName;
    if (this.source === "auto") {
      sourceName = this.findMonitorSource();
    } else {
      sourceName = this.source;
    }

    console.log(`[audio] Starting capture: source=${sourceName} rate=${this.sampleRate}`);

    // Use parec (PulseAudio record) to capture audio
    // Output: raw PCM 16-bit signed little-endian mono
    this.process = spawn("parec", [
      "--device", sourceName,
      "--format", "s16le",
      "--rate", String(this.sampleRate),
      "--channels", "1",
      "--raw",
    ]);

    this.isCapturing = true;

    // Calculate bytes per chunk
    // PCM16 = 2 bytes per sample
    const bytesPerChunk = Math.floor((this.sampleRate * this.chunkDurationMs * 2) / 1000);
    let buffer = Buffer.alloc(0);

    this.process.stdout.on("data", (data) => {
      buffer = Buffer.concat([buffer, data]);

      // Emit chunks of the configured duration
      while (buffer.length >= bytesPerChunk) {
        const chunk = buffer.subarray(0, bytesPerChunk);
        buffer = buffer.subarray(bytesPerChunk);

        // Convert to base64
        const base64 = chunk.toString("base64");
        this.emit("audio", base64);
      }
    });

    this.process.stderr.on("data", (data) => {
      const msg = data.toString().trim();
      if (msg && !msg.includes("RUNNING")) {
        console.error(`[audio] parec stderr: ${msg}`);
      }
    });

    this.process.on("error", (err) => {
      console.error(`[audio] Process error: ${err.message}`);
      if (err.code === "ENOENT") {
        console.error("[audio] 'parec' not found. Install: sudo apt install pulseaudio-utils");
      }
      this.emit("error", err);
    });

    this.process.on("close", (code) => {
      this.isCapturing = false;
      if (code !== 0 && code !== null) {
        console.error(`[audio] parec exited with code ${code}`);
      }
      this.emit("stop");
    });

    console.log("[audio] Capture started - listening to system audio");
  }

  /**
   * Stop capturing
   */
  stop() {
    if (this.process) {
      this.process.kill("SIGTERM");
      this.process = null;
    }
    this.isCapturing = false;
    console.log("[audio] Capture stopped");
  }
}

module.exports = AudioCapture;
