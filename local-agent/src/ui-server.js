const http = require("http");
const express = require("express");
const path = require("path");
const WebSocket = require("ws");

class UIServer {
    constructor(port) {
        this.port = port || 3000;
        this.clients = new Set();

        const app = express();
        app.use(express.static(path.join(__dirname, "..", "public")));

        // Redirect root to overlay
        app.get("/", (_req, res) => {
            res.redirect("/overlay.html");
        });

        this.server = http.createServer(app);
        this.wss = new WebSocket.Server({ server: this.server, path: "/ws" });

        this.wss.on("connection", (ws) => {
            this.clients.add(ws);
            console.log("[UIServer] Overlay client connected, total:", this.clients.size);

            ws.on("message", (data) => {
                try {
                    const msg = JSON.parse(data);
                    this.onCommand(msg);
                } catch (e) {}
            });

            ws.on("close", () => {
                this.clients.delete(ws);
                console.log("[UIServer] Overlay client disconnected, total:", this.clients.size);
            });
        });

        // Command handler set by index.js
        this.onCommand = () => {};
    }

    start() {
        return new Promise((resolve) => {
            this.server.listen(this.port, () => {
                console.log("[UIServer] Overlay UI at http://localhost:" + this.port);
                resolve();
            });
        });
    }

    broadcast(event) {
        const data = JSON.stringify(event);
        for (const ws of this.clients) {
            if (ws.readyState === WebSocket.OPEN) {
                ws.send(data);
            }
        }
    }

    stop() {
        this.wss.close();
        this.server.close();
    }
}

module.exports = UIServer;
