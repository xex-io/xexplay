import http from "k6/http";
import { check, sleep, group } from "k6";
import { Rate, Trend } from "k6/metrics";

// Custom metrics
const errorRate = new Rate("errors");
const resolveDuration = new Trend("resolve_card_duration", true);
const getCardsDuration = new Trend("get_cards_duration", true);

export const options = {
  scenarios: {
    // Scenario 1: Admin resolves cards (low concurrency, high impact)
    card_resolution: {
      executor: "per-vu-iterations",
      vus: 5, // small number of admins resolving cards
      iterations: 50, // each admin resolves multiple cards
      maxDuration: "5m",
    },
    // Scenario 2: Concurrent user reads during resolution
    // Simulates users hitting the API while cards are being resolved
    user_reads_during_resolution: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "15s", target: 100 },
        { duration: "2m", target: 500 },
        { duration: "2m", target: 500 }, // sustained
        { duration: "30s", target: 0 },
      ],
      gracefulRampDown: "10s",
    },
  },
  thresholds: {
    "resolve_card_duration": ["p(95)<2000"], // resolution can be heavier (batch updates)
    "get_cards_duration": ["p(95)<200"],
    "errors": ["rate<0.10"],
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const AUTH_TOKEN = __ENV.AUTH_TOKEN || "";
// Admin token with admin role for resolving cards
const ADMIN_TOKEN = __ENV.ADMIN_TOKEN || AUTH_TOKEN;
// Comma-separated list of card UUIDs to resolve (pre-created in the DB)
const CARD_IDS = (__ENV.CARD_IDS || "").split(",").filter((id) => id.length > 0);

function authHeaders(token) {
  return {
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
  };
}

// Scenario handler: admin resolves cards
export function card_resolution() {
  if (CARD_IDS.length === 0) {
    console.warn(
      "No CARD_IDS provided. Set CARD_IDS env var with comma-separated UUIDs."
    );
    sleep(1);
    return;
  }

  group("Admin Resolves Card", function () {
    // Pick a card to resolve (cycle through available cards)
    let idx = (__VU * __ITER) % CARD_IDS.length;
    let cardId = CARD_IDS[idx].trim();

    // Step 1: Get card list (to verify card exists)
    let listRes = http.get(
      `${BASE_URL}/v1/admin/cards`,
      authHeaders(ADMIN_TOKEN)
    );
    getCardsDuration.add(listRes.timings.duration);
    check(listRes, {
      "list cards: status 200": (r) => r.status === 200,
    });

    sleep(Math.random() * 1 + 0.5); // admin reviews

    // Step 2: Resolve the card
    let correctAnswer = Math.random() < 0.5;
    let resolveRes = http.post(
      `${BASE_URL}/v1/admin/cards/${cardId}/resolve`,
      JSON.stringify({ correct_answer: correctAnswer }),
      authHeaders(ADMIN_TOKEN)
    );
    resolveDuration.add(resolveRes.timings.duration);
    let ok = check(resolveRes, {
      "resolve: status 200 or 400": (r) =>
        r.status === 200 || r.status === 400,
      "resolve: has message": (r) => {
        try {
          let body = JSON.parse(r.body);
          return body.data && body.data.message;
        } catch (_) {
          return false;
        }
      },
    });
    errorRate.add(!ok);

    sleep(Math.random() * 3 + 2); // admin waits between resolutions
  });
}

// Scenario handler: user reads during card resolution
export function user_reads_during_resolution() {
  group("User Activity During Resolution", function () {
    // Simulate users checking their session, cards, and leaderboard
    // while admin is resolving cards in parallel

    // Get current card (may have just been resolved)
    let cardRes = http.get(
      `${BASE_URL}/v1/sessions/current/card`,
      authHeaders(AUTH_TOKEN)
    );
    check(cardRes, {
      "get card during resolution: status ok": (r) =>
        r.status === 200 || r.status === 400 || r.status === 404,
    });

    sleep(Math.random() * 0.5 + 0.2);

    // Get session state
    let sessionRes = http.get(
      `${BASE_URL}/v1/sessions/current`,
      authHeaders(AUTH_TOKEN)
    );
    check(sessionRes, {
      "get session during resolution: status ok": (r) =>
        r.status === 200 || r.status === 400 || r.status === 404,
    });

    sleep(Math.random() * 0.5 + 0.2);

    // Check leaderboard (scores update after resolution)
    let lbRes = http.get(
      `${BASE_URL}/v1/leaderboards/daily?limit=20`,
      authHeaders(AUTH_TOKEN)
    );
    check(lbRes, {
      "leaderboard during resolution: status 200": (r) => r.status === 200,
    });

    sleep(Math.random() * 2 + 1);
  });
}

// Route scenarios to their handlers
export default function () {
  // The default function is used by the user_reads_during_resolution scenario
  user_reads_during_resolution();
}

// Named export for the card_resolution scenario
export { card_resolution };
