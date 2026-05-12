// Realtime Speech AI Agent - Client
(function () {
    "use strict";

    // --- State ---
    let ws = null;
    let mediaStream = null;
    let audioContext = null;
    let processor = null;
    let isRecording = false;
    let sessionId = null;
    let currentAnswer = "";
    let hasNotes = false;

    // --- DOM ---
    const statusBadge = document.getElementById("statusBadge");
    const transcriptArea = document.getElementById("transcriptArea");
    const answerArea = document.getElementById("answerArea");
    const notesArea = document.getElementById("notesArea");
    const errorArea = document.getElementById("errorArea");
    const btnRecord = document.getElementById("btnRecord");
    const btnInterrupt = document.getElementById("btnInterrupt");
    const btnStop = document.getElementById("btnStop");
    const sessionInfo = document.getElementById("sessionInfo");

    // --- WebSocket ---
    function connectWebSocket() {
        const protocol = location.protocol === "https:" ? "wss:" : "ws:";
        const url = protocol + "//" + location.host + "/ws/realtime";

        ws = new WebSocket(url);

        ws.onopen = function () {
            console.log("[ws] connected");
            sessionInfo.textContent = "Connected to server";
        };

        ws.onmessage = function (e) {
            try {
                const evt = JSON.parse(e.data);
                handleServerEvent(evt);
            } catch (err) {
                console.error("[ws] parse error:", err);
            }
        };

        ws.onclose = function () {
            console.log("[ws] disconnected");
            sessionInfo.textContent = "Disconnected";
            setStatus("idle");
            if (isRecording) {
                stopRecording();
            }
        };

        ws.onerror = function (err) {
            console.error("[ws] error:", err);
            showError("WebSocket connection error");
        };
    }

    function sendEvent(evt) {
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify(evt));
        }
    }

    // --- Event Handlers ---
    function handleServerEvent(evt) {
        switch (evt.type) {
            case "session_started":
                setStatus("listening");
                sessionInfo.textContent = "Session: " + evt.session_id;
                btnStop.disabled = false;
                break;

            case "partial_text":
                showPartialTranscript(evt.lang, evt.text);
                setStatus("transcribing");
                break;

            case "final_text":
                showFinalTranscript(evt.lang, evt.text);
                setStatus("listening");
                break;

            case "answer_delta":
                appendAnswerDelta(evt.text);
                setStatus("answering");
                btnInterrupt.disabled = false;
                break;

            case "answer_completed":
                completeAnswer(evt.text);
                setStatus("listening");
                btnInterrupt.disabled = true;
                break;

            case "note_created":
                addNote(evt.text);
                break;

            case "error":
                showError(evt.code + ": " + evt.message);
                break;

            default:
                console.log("[ws] unknown event:", evt.type);
        }
    }

    // --- UI Updates ---
    function setStatus(status) {
        statusBadge.textContent = status.toUpperCase();
        statusBadge.className = "status-badge status-" + status;
    }

    function showPartialTranscript(lang, text) {
        // Remove previous partial line
        const existing = transcriptArea.querySelector(".partial");
        if (existing) existing.remove();

        const div = document.createElement("div");
        div.className = "transcript-line partial";
        div.innerHTML = '<span class="lang-tag">[' + (lang || "?") + ']</span>' + escapeHtml(text);
        transcriptArea.appendChild(div);
        transcriptArea.scrollTop = transcriptArea.scrollHeight;
    }

    function showFinalTranscript(lang, text) {
        // Remove partial line
        const existing = transcriptArea.querySelector(".partial");
        if (existing) existing.remove();

        // Clear placeholder
        if (transcriptArea.querySelector("[style]")) {
            transcriptArea.innerHTML = "";
        }

        const div = document.createElement("div");
        div.className = "transcript-line final";
        div.innerHTML = '<span class="lang-tag">[' + (lang || "?") + ']</span>' + escapeHtml(text);
        transcriptArea.appendChild(div);
        transcriptArea.scrollTop = transcriptArea.scrollHeight;
    }

    function appendAnswerDelta(text) {
        if (answerArea.querySelector("[style]")) {
            answerArea.innerHTML = "";
            currentAnswer = "";
        }
        currentAnswer += text;
        answerArea.textContent = currentAnswer;
        answerArea.scrollTop = answerArea.scrollHeight;
    }

    function completeAnswer(text) {
        currentAnswer = text;
        answerArea.textContent = text;
    }

    function addNote(text) {
        if (!hasNotes) {
            notesArea.innerHTML = "";
            hasNotes = true;
        }
        const div = document.createElement("div");
        div.className = "note-item";
        div.textContent = text;
        notesArea.appendChild(div);
        notesArea.scrollTop = notesArea.scrollHeight;
    }

    function showError(msg) {
        errorArea.innerHTML = '<div class="error-msg">' + escapeHtml(msg) + "</div>";
        setTimeout(function () {
            errorArea.innerHTML = "";
        }, 5000);
    }

    function escapeHtml(text) {
        const div = document.createElement("div");
        div.appendChild(document.createTextNode(text));
        return div.innerHTML;
    }

    // --- Audio Recording ---
    async function startRecording() {
        try {
            mediaStream = await navigator.mediaDevices.getUserMedia({
                audio: {
                    sampleRate: 16000,
                    channelCount: 1,
                    echoCancellation: true,
                    noiseSuppression: true,
                },
            });

            audioContext = new AudioContext({ sampleRate: 16000 });
            const source = audioContext.createMediaStreamSource(mediaStream);

            // Use ScriptProcessorNode for simplicity (MVP)
            processor = audioContext.createScriptProcessor(4096, 1, 1);
            processor.onaudioprocess = function (e) {
                if (!isRecording) return;

                const inputData = e.inputBuffer.getChannelData(0);
                // Convert float32 to PCM16
                const pcm16 = new Int16Array(inputData.length);
                for (let i = 0; i < inputData.length; i++) {
                    const s = Math.max(-1, Math.min(1, inputData[i]));
                    pcm16[i] = s < 0 ? s * 0x8000 : s * 0x7fff;
                }

                // Convert to base64
                const bytes = new Uint8Array(pcm16.buffer);
                let binary = "";
                for (let i = 0; i < bytes.length; i++) {
                    binary += String.fromCharCode(bytes[i]);
                }
                const base64 = btoa(binary);

                sendEvent({
                    type: "audio_chunk",
                    session_id: sessionId,
                    format: "pcm16",
                    sample_rate: 16000,
                    data: base64,
                });
            };

            source.connect(processor);
            processor.connect(audioContext.destination);

            isRecording = true;
            btnRecord.textContent = "🔴 Recording...";
            btnRecord.classList.add("recording");
        } catch (err) {
            console.error("[audio] error:", err);
            showError("Microphone access denied or not available");
        }
    }

    function stopRecording() {
        isRecording = false;

        if (processor) {
            processor.disconnect();
            processor = null;
        }
        if (audioContext) {
            audioContext.close();
            audioContext = null;
        }
        if (mediaStream) {
            mediaStream.getTracks().forEach(function (t) { t.stop(); });
            mediaStream = null;
        }

        btnRecord.textContent = "🎤 Start Recording";
        btnRecord.classList.remove("recording");
    }

    // --- Public Actions ---
    window.toggleRecording = function () {
        if (!isRecording) {
            // Connect WS and start session
            if (!ws || ws.readyState !== WebSocket.OPEN) {
                connectWebSocket();
                // Wait for connection
                setTimeout(function () {
                    sessionId = "session_" + Date.now();
                    sendEvent({
                        type: "start_session",
                        session_id: sessionId,
                        language: "auto",
                        enable_tts: false,
                    });
                    startRecording();
                }, 500);
            } else {
                sessionId = "session_" + Date.now();
                sendEvent({
                    type: "start_session",
                    session_id: sessionId,
                    language: "auto",
                    enable_tts: false,
                });
                startRecording();
            }
        } else {
            stopRecording();
        }
    };

    window.sendInterrupt = function () {
        if (sessionId) {
            sendEvent({
                type: "interrupt",
                session_id: sessionId,
            });
            setStatus("listening");
            btnInterrupt.disabled = true;
            currentAnswer = "";
        }
    };

    window.stopSession = function () {
        if (sessionId) {
            sendEvent({
                type: "stop_session",
                session_id: sessionId,
            });
        }
        stopRecording();
        setStatus("idle");
        btnStop.disabled = true;
        btnInterrupt.disabled = true;
        sessionId = null;
        sessionInfo.textContent = "Session ended";
    };
})();
