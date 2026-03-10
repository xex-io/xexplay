import http from "k6/http";
import { check, sleep, group } from "k6";
import { Rate, Trend } from "k6/metrics";

// Custom metrics
const errorRate = new Rate("errors");
const startSessionDuration = new Trend("start_session_duration", true);
const getCardDuration = new Trend("get_card_duration", true);
const submitAnswerDuration = new Trend("submit_answer_duration", true);
const skipCardDuration = new Trend("skip_card_duration", true);

export const options = {
  scenarios: {
    game_session_flow: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "30s", target: 200 },
        { duration: "1m", target: 500 },
        { duration: "2m", target: 1000 },
        { duration: "3m", target: 1000 }, // sustained peak
        { duration: "1m", target: 500 },
        { duration: "30s", target: 0 },
      ],
      gracefulRampDown: "10s",
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<200"],
    errors: ["rate<0.05"],
    start_session_duration: ["p(95)<200"],
    get_card_duration: ["p(95)<200"],
    submit_answer_duration: ["p(95)<200"],
    skip_card_duration: ["p(95)<200"],
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const AUTH_TOKEN = __ENV.AUTH_TOKEN || "";

function authHeaders() {
  return {
    headers: {
      Authorization: `Bearer ${AUTH_TOKEN}`,
      "Content-Type": "application/json",
    },
  };
}

export default function () {
  group("Game Session Flow", function () {
    // Step 1: Start a session
    let startRes = http.post(
      `${BASE_URL}/v1/sessions/start`,
      null,
      authHeaders()
    );
    startSessionDuration.add(startRes.timings.duration);
    let startOk = check(startRes, {
      "start session: status 200": (r) => r.status === 200,
      "start session: has session id": (r) => {
        try {
          let body = JSON.parse(r.body);
          return body.data && body.data.id;
        } catch (_) {
          return false;
        }
      },
    });
    errorRate.add(!startOk);

    if (startRes.status !== 200) {
      sleep(1);
      return;
    }

    sleep(Math.random() * 2 + 0.5); // think time: 0.5-2.5s

    // Step 2: Get current card
    let cardRes = http.get(
      `${BASE_URL}/v1/sessions/current/card`,
      authHeaders()
    );
    getCardDuration.add(cardRes.timings.duration);
    let cardOk = check(cardRes, {
      "get card: status 200": (r) => r.status === 200,
    });
    errorRate.add(!cardOk);

    if (cardRes.status !== 200) {
      sleep(1);
      return;
    }

    sleep(Math.random() * 3 + 1); // think time: 1-4s (user reads the card)

    // Step 3: Submit answer or skip (80% answer, 20% skip)
    if (Math.random() < 0.8) {
      let answer = Math.random() < 0.5; // random true/false prediction
      let answerRes = http.post(
        `${BASE_URL}/v1/sessions/current/answer`,
        JSON.stringify({ answer: answer }),
        authHeaders()
      );
      submitAnswerDuration.add(answerRes.timings.duration);
      let answerOk = check(answerRes, {
        "submit answer: status 200": (r) => r.status === 200,
      });
      errorRate.add(!answerOk);
    } else {
      let skipRes = http.post(
        `${BASE_URL}/v1/sessions/current/skip`,
        null,
        authHeaders()
      );
      skipCardDuration.add(skipRes.timings.duration);
      let skipOk = check(skipRes, {
        "skip card: status 200": (r) => r.status === 200,
      });
      errorRate.add(!skipOk);
    }

    sleep(Math.random() * 2 + 0.5); // think time between cards

    // Step 4: Get current session state
    let sessionRes = http.get(
      `${BASE_URL}/v1/sessions/current`,
      authHeaders()
    );
    check(sessionRes, {
      "get session: status 200": (r) => r.status === 200,
    });

    sleep(Math.random() * 2 + 1); // think time before next iteration
  });
}
