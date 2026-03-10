import http from "k6/http";
import { check, sleep, group } from "k6";
import { Rate, Trend } from "k6/metrics";

// Custom metrics
const errorRate = new Rate("errors");
const dailyDuration = new Trend("daily_leaderboard_duration", true);
const weeklyDuration = new Trend("weekly_leaderboard_duration", true);
const allTimeDuration = new Trend("all_time_leaderboard_duration", true);
const tournamentDuration = new Trend("tournament_leaderboard_duration", true);

export const options = {
  scenarios: {
    leaderboard_reads: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "20s", target: 100 },
        { duration: "1m", target: 500 },
        { duration: "2m", target: 1000 },
        { duration: "2m", target: 1000 }, // sustained peak
        { duration: "30s", target: 0 },
      ],
      gracefulRampDown: "10s",
    },
    // Simulate burst reads (e.g., after a match ends)
    leaderboard_burst: {
      executor: "constant-arrival-rate",
      rate: 500, // 500 requests per second
      timeUnit: "1s",
      duration: "30s",
      preAllocatedVUs: 200,
      maxVUs: 500,
      startTime: "3m30s", // after the ramp scenario
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<100"],
    errors: ["rate<0.05"],
    daily_leaderboard_duration: ["p(95)<100"],
    weekly_leaderboard_duration: ["p(95)<100"],
    all_time_leaderboard_duration: ["p(95)<100"],
    tournament_leaderboard_duration: ["p(95)<100"],
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const AUTH_TOKEN = __ENV.AUTH_TOKEN || "";
// Provide a valid tournament event UUID for tournament leaderboard tests
const TOURNAMENT_EVENT_ID =
  __ENV.TOURNAMENT_EVENT_ID || "00000000-0000-0000-0000-000000000001";

function authHeaders() {
  return {
    headers: {
      Authorization: `Bearer ${AUTH_TOKEN}`,
      "Content-Type": "application/json",
    },
  };
}

export default function () {
  // Randomly pick a leaderboard type to simulate realistic access patterns:
  // 40% daily, 20% weekly, 15% all-time, 15% tournament, 10% paginated
  let roll = Math.random();

  if (roll < 0.4) {
    group("Daily Leaderboard", function () {
      let res = http.get(
        `${BASE_URL}/v1/leaderboards/daily?limit=50`,
        authHeaders()
      );
      dailyDuration.add(res.timings.duration);
      let ok = check(res, {
        "daily: status 200": (r) => r.status === 200,
        "daily: returns data": (r) => {
          try {
            let body = JSON.parse(r.body);
            return body.data !== undefined;
          } catch (_) {
            return false;
          }
        },
      });
      errorRate.add(!ok);
    });
  } else if (roll < 0.6) {
    group("Weekly Leaderboard", function () {
      let res = http.get(
        `${BASE_URL}/v1/leaderboards/weekly?limit=50`,
        authHeaders()
      );
      weeklyDuration.add(res.timings.duration);
      let ok = check(res, {
        "weekly: status 200": (r) => r.status === 200,
      });
      errorRate.add(!ok);
    });
  } else if (roll < 0.75) {
    group("All-Time Leaderboard", function () {
      let res = http.get(
        `${BASE_URL}/v1/leaderboards/all-time?limit=50`,
        authHeaders()
      );
      allTimeDuration.add(res.timings.duration);
      let ok = check(res, {
        "all-time: status 200": (r) => r.status === 200,
      });
      errorRate.add(!ok);
    });
  } else if (roll < 0.9) {
    group("Tournament Leaderboard", function () {
      let res = http.get(
        `${BASE_URL}/v1/leaderboards/tournament/${TOURNAMENT_EVENT_ID}?limit=50`,
        authHeaders()
      );
      tournamentDuration.add(res.timings.duration);
      let ok = check(res, {
        "tournament: status 200 or 404": (r) =>
          r.status === 200 || r.status === 404,
      });
      errorRate.add(!ok);
    });
  } else {
    // Paginated reads - simulate scrolling through leaderboard pages
    group("Paginated Leaderboard", function () {
      let offsets = [0, 50, 100];
      for (let offset of offsets) {
        let res = http.get(
          `${BASE_URL}/v1/leaderboards/daily?limit=50&offset=${offset}`,
          authHeaders()
        );
        dailyDuration.add(res.timings.duration);
        let ok = check(res, {
          "paginated: status 200": (r) => r.status === 200,
        });
        errorRate.add(!ok);

        sleep(0.3); // brief pause between page loads
      }
    });
  }

  sleep(Math.random() * 1 + 0.5); // think time: 0.5-1.5s
}
