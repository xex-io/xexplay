import ws from "k6/ws";
import { check, sleep } from "k6";
import { Rate, Counter, Trend } from "k6/metrics";

// Custom metrics
const wsConnectErrors = new Rate("ws_connect_errors");
const wsMessages = new Counter("ws_messages_received");
const wsConnectDuration = new Trend("ws_connect_duration", true);

export const options = {
  scenarios: {
    websocket_connections: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "30s", target: 500 },
        { duration: "1m", target: 2000 },
        { duration: "1m", target: 5000 },
        { duration: "3m", target: 5000 }, // sustained peak: 5000 concurrent
        { duration: "1m", target: 2000 },
        { duration: "30s", target: 0 },
      ],
      gracefulRampDown: "30s",
    },
  },
  thresholds: {
    ws_connect_errors: ["rate<0.05"],
    ws_connect_duration: ["p(95)<1000"], // connection setup under 1s
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const AUTH_TOKEN = __ENV.AUTH_TOKEN || "";

// Derive WebSocket URL from BASE_URL
function getWsUrl() {
  let url = BASE_URL.replace("https://", "wss://").replace(
    "http://",
    "ws://"
  );
  return `${url}/ws?token=${AUTH_TOKEN}`;
}

export default function () {
  let connectStart = Date.now();

  let res = ws.connect(getWsUrl(), {}, function (socket) {
    let connectDuration = Date.now() - connectStart;
    wsConnectDuration.add(connectDuration);

    socket.on("open", function () {
      // Connection established - just stay connected and listen for broadcasts
    });

    socket.on("message", function (msg) {
      wsMessages.add(1);

      // Validate broadcast message structure
      try {
        let data = JSON.parse(msg);
        check(data, {
          "ws message: has type": (d) => d.type !== undefined,
        });
      } catch (_) {
        // Some messages might not be JSON (ping/pong)
      }
    });

    socket.on("error", function (e) {
      wsConnectErrors.add(1);
    });

    socket.on("close", function () {
      // Connection closed by server
    });

    // Keep connection alive for 30-60 seconds to simulate a real user session
    // Send periodic pings to keep connection alive
    let duration = Math.random() * 30 + 30; // 30-60s
    let elapsed = 0;
    let pingInterval = 10; // seconds

    socket.setInterval(function () {
      elapsed += pingInterval;

      // Send a ping to keep connection alive
      socket.ping();

      if (elapsed >= duration) {
        socket.close();
      }
    }, pingInterval * 1000);

    // Set a hard timeout
    socket.setTimeout(function () {
      socket.close();
    }, (duration + 5) * 1000);
  });

  let connected = check(res, {
    "ws: connection established": (r) => r && r.status === 101,
  });
  wsConnectErrors.add(!connected);

  sleep(Math.random() * 2 + 1); // pause before reconnecting
}
