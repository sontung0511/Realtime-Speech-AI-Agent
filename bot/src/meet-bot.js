/**
 * Google Meet Bot - Joins meeting as participant and captures audio
 * Uses Puppeteer to control a headless Chrome browser
 */

const puppeteer = require("puppeteer");

class MeetBot {
  constructor(config) {
    this.meetUrl = config.meetUrl;
    this.botName = config.botName || "AI Assistant";
    this.headless = config.headless !== false;
    this.googleEmail = config.googleEmail;
    this.googlePassword = config.googlePassword;

    this.browser = null;
    this.page = null;
    this.isInMeeting = false;
  }

  /**
   * Launch browser and join the Google Meet.
   */
  async join() {
    console.log(`[bot] Launching browser (headless: ${this.headless})`);

    this.browser = await puppeteer.launch({
      headless: this.headless ? "new" : false,
      args: [
        "--use-fake-ui-for-media-stream", // Auto-allow mic/cam permissions
        "--use-fake-device-for-media-stream", // Use fake media device
        "--disable-web-security",
        "--allow-running-insecure-content",
        "--autoplay-policy=no-user-gesture-required",
        "--no-sandbox",
        "--disable-setuid-sandbox",
        "--disable-gpu",
        "--disable-dev-shm-usage",
        // Enable audio capture from tab
        "--enable-usermedia-screen-capturing",
        "--allow-http-screen-capture",
        "--auto-select-desktop-capture-source=Entire screen",
      ],
    });

    this.page = await this.browser.newPage();

    // Set permissions
    const context = this.browser.defaultBrowserContext();
    await context.overridePermissions("https://meet.google.com", [
      "microphone",
      "camera",
      "notifications",
    ]);

    // If Google credentials provided, sign in first
    if (this.googleEmail && this.googlePassword) {
      await this._signInGoogle();
    }

    // Navigate to meeting
    console.log(`[bot] Navigating to: ${this.meetUrl}`);
    await this.page.goto(this.meetUrl, { waitUntil: "networkidle2", timeout: 30000 });

    // Wait for page to load
    await this._sleep(3000);

    // Handle "Join now" flow
    await this._joinMeeting();

    this.isInMeeting = true;
    console.log("[bot] Successfully joined the meeting!");
  }

  /**
   * Sign in to Google account
   */
  async _signInGoogle() {
    console.log(`[bot] Signing in as: ${this.googleEmail}`);

    await this.page.goto("https://accounts.google.com/signin", {
      waitUntil: "networkidle2",
    });

    // Enter email
    await this.page.waitForSelector('input[type="email"]', { timeout: 10000 });
    await this.page.type('input[type="email"]', this.googleEmail, { delay: 50 });
    await this.page.click("#identifierNext");
    await this._sleep(3000);

    // Enter password
    await this.page.waitForSelector('input[type="password"]', { visible: true, timeout: 10000 });
    await this.page.type('input[type="password"]', this.googlePassword, { delay: 50 });
    await this.page.click("#passwordNext");
    await this._sleep(5000);

    console.log("[bot] Google sign-in completed");
  }

  /**
   * Handle the Google Meet join flow (disable cam/mic, set name, click Join)
   */
  async _joinMeeting() {
    try {
      // Try to dismiss any initial dialogs
      await this._dismissDialogs();

      // Turn off camera
      await this._clickButton('[aria-label*="camera" i]');
      await this._sleep(500);

      // Turn off microphone (bot only listens, doesn't speak)
      await this._clickButton('[aria-label*="microphone" i]');
      await this._sleep(500);

      // If anonymous join, set display name
      if (!this.googleEmail) {
        await this._setDisplayName();
      }

      // Click "Join now" or "Ask to join"
      await this._clickJoinButton();
      await this._sleep(5000);

      // If "Ask to join" was clicked, wait for host to admit
      await this._waitForAdmission();

    } catch (err) {
      console.error("[bot] Error during join flow:", err.message);
      // Take screenshot for debugging
      await this.page.screenshot({ path: "bot-join-error.png" });
      throw err;
    }
  }

  /**
   * Dismiss cookie banners and other dialogs
   */
  async _dismissDialogs() {
    try {
      // Dismiss "Got it" or cookie consent
      const dismissSelectors = [
        'button[aria-label="Got it"]',
        '[aria-label="Dismiss"]',
        'button:has-text("Got it")',
        'button:has-text("Accept all")',
      ];
      for (const sel of dismissSelectors) {
        try {
          const btn = await this.page.$(sel);
          if (btn) {
            await btn.click();
            await this._sleep(500);
          }
        } catch (_) {}
      }
    } catch (_) {}
  }

  /**
   * Set bot display name for anonymous join
   */
  async _setDisplayName() {
    try {
      const nameInput = await this.page.$('input[aria-label="Your name"]');
      if (nameInput) {
        await nameInput.click({ clickCount: 3 }); // Select all
        await nameInput.type(this.botName, { delay: 30 });
        console.log(`[bot] Set display name: ${this.botName}`);
      }
    } catch (_) {}
  }

  /**
   * Click the Join/Ask to join button
   */
  async _clickJoinButton() {
    const joinSelectors = [
      'button[data-idom-class*="join"]',
      '[jsname="Qx7uuf"]', // "Join now" button jsname
      'button:has(span:has-text("Join now"))',
      'button:has(span:has-text("Ask to join"))',
      'button:has(span:has-text("Tham gia ngay"))', // Vietnamese
      'button:has(span:has-text("Yêu cầu tham gia"))', // Vietnamese
    ];

    for (const sel of joinSelectors) {
      try {
        const btn = await this.page.$(sel);
        if (btn) {
          await btn.click();
          console.log("[bot] Clicked join button");
          return;
        }
      } catch (_) {}
    }

    // Fallback: find button by text content
    const buttons = await this.page.$$("button");
    for (const btn of buttons) {
      const text = await this.page.evaluate((el) => el.textContent, btn);
      if (
        text &&
        (text.includes("Join now") ||
          text.includes("Ask to join") ||
          text.includes("Tham gia") ||
          text.includes("Yêu cầu"))
      ) {
        await btn.click();
        console.log(`[bot] Clicked button: "${text.trim()}"`);
        return;
      }
    }

    throw new Error("Could not find Join button");
  }

  /**
   * Wait for host to admit the bot (for "Ask to join" scenario)
   */
  async _waitForAdmission() {
    // Check if we're actually in the meeting by looking for meeting controls
    const maxWait = 120000; // 2 minutes
    const start = Date.now();

    while (Date.now() - start < maxWait) {
      // Check for meeting indicators
      const inMeeting = await this.page.evaluate(() => {
        // If we can see participant list or chat button, we're in
        return !!(
          document.querySelector('[aria-label*="people" i]') ||
          document.querySelector('[aria-label*="chat" i]') ||
          document.querySelector('[data-meeting-title]') ||
          document.querySelector('[aria-label*="Leave" i]')
        );
      });

      if (inMeeting) {
        console.log("[bot] Confirmed in meeting");
        return;
      }

      await this._sleep(2000);
    }

    console.log("[bot] Warning: Could not confirm meeting admission after 2 min");
  }

  /**
   * Inject audio capture script into the page.
   * Captures all audio playing in the meeting tab.
   * Returns audio data via a callback.
   */
  async startAudioCapture(onAudioChunk) {
    if (!this.isInMeeting) {
      throw new Error("Bot is not in a meeting");
    }

    console.log("[bot] Starting audio capture from meeting...");

    // Expose callback to receive audio data from browser
    await this.page.exposeFunction("__onAudioChunk", (base64Data) => {
      onAudioChunk(base64Data);
    });

    // Inject audio capture script
    await this.page.evaluate((sampleRate, chunkMs) => {
      // Capture all audio output from the page
      const audioCtx = new AudioContext({ sampleRate: sampleRate });

      // Create a destination to capture audio
      const dest = audioCtx.createMediaStreamDestination();

      // Hook into all audio/video elements to capture their output
      function captureAllMedia() {
        const elements = document.querySelectorAll("audio, video");
        elements.forEach((el) => {
          if (!el.__captured) {
            try {
              const source = audioCtx.createMediaElementSource(el);
              source.connect(dest);
              source.connect(audioCtx.destination); // Also play normally
              el.__captured = true;
              console.log("[audio-capture] Captured media element");
            } catch (e) {
              // Already captured or CORS issue
            }
          }
        });
      }

      // Also try to capture via WebRTC
      const origAddTrack = RTCPeerConnection.prototype.addTrack;
      const origOnTrack = Object.getOwnPropertyDescriptor(
        RTCPeerConnection.prototype,
        "ontrack"
      );

      // Intercept incoming audio tracks from WebRTC (meeting audio)
      const capturedStreams = new Set();

      // Monitor for new media streams from WebRTC
      const origSetRemoteDesc =
        RTCPeerConnection.prototype.setRemoteDescription;
      RTCPeerConnection.prototype.setRemoteDescription = function (...args) {
        this.addEventListener("track", (event) => {
          if (event.track.kind === "audio") {
            const stream = event.streams[0] || new MediaStream([event.track]);
            const streamId = stream.id;
            if (!capturedStreams.has(streamId)) {
              capturedStreams.add(streamId);
              try {
                const source = audioCtx.createMediaStreamSource(stream);
                source.connect(dest);
                console.log("[audio-capture] Captured WebRTC audio stream");
              } catch (e) {
                console.error("[audio-capture] WebRTC capture error:", e);
              }
            }
          }
        });
        return origSetRemoteDesc.apply(this, args);
      };

      // Process captured audio into PCM16 chunks
      const stream = dest.stream;
      const recorder = audioCtx.createScriptProcessor(4096, 1, 1);
      const source = audioCtx.createMediaStreamSource(stream);

      let buffer = [];
      const samplesPerChunk = Math.floor((sampleRate * chunkMs) / 1000);

      recorder.onaudioprocess = (e) => {
        const input = e.inputBuffer.getChannelData(0);

        // Convert Float32 to Int16 (PCM16)
        for (let i = 0; i < input.length; i++) {
          const s = Math.max(-1, Math.min(1, input[i]));
          buffer.push(s < 0 ? s * 0x8000 : s * 0x7fff);
        }

        // When we have enough samples, send a chunk
        while (buffer.length >= samplesPerChunk) {
          const chunk = buffer.splice(0, samplesPerChunk);
          const pcm16 = new Int16Array(chunk);
          const bytes = new Uint8Array(pcm16.buffer);

          // Convert to base64
          let binary = "";
          for (let j = 0; j < bytes.length; j++) {
            binary += String.fromCharCode(bytes[j]);
          }
          const base64 = btoa(binary);

          // Send to Node.js
          window.__onAudioChunk(base64);
        }
      };

      source.connect(recorder);
      recorder.connect(audioCtx.destination);

      // Periodically check for new media elements
      setInterval(captureAllMedia, 2000);
      captureAllMedia();

      console.log("[audio-capture] Audio capture initialized");
    }, 16000, 1000);

    console.log("[bot] Audio capture active - streaming to server");
  }

  /**
   * Leave the meeting and close browser
   */
  async leave() {
    if (!this.page) return;

    try {
      // Click "Leave call" button
      const leaveBtn = await this.page.$('[aria-label*="Leave" i]');
      if (leaveBtn) {
        await leaveBtn.click();
        await this._sleep(1000);
      }
    } catch (_) {}

    this.isInMeeting = false;

    if (this.browser) {
      await this.browser.close();
      this.browser = null;
    }

    console.log("[bot] Left meeting and closed browser");
  }

  async _clickButton(selector) {
    try {
      const btn = await this.page.$(selector);
      if (btn) await btn.click();
    } catch (_) {}
  }

  _sleep(ms) {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }
}

module.exports = MeetBot;
