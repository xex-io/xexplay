# XEX Play Load Tests

k6 load test scripts for the XEX Play backend API.

## Prerequisites

Install [k6](https://k6.io/docs/get-started/installation/):

```bash
# macOS
brew install k6

# Docker
docker pull grafana/k6
```

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `BASE_URL` | Yes | API base URL (e.g., `http://localhost:8080`) |
| `AUTH_TOKEN` | Yes | Valid JWT token for an authenticated user |
| `ADMIN_TOKEN` | For card_resolution | JWT token with admin role |
| `TOURNAMENT_EVENT_ID` | For leaderboard | UUID of an existing tournament event |
| `CARD_IDS` | For card_resolution | Comma-separated card UUIDs to resolve |

## Test Scripts

### 1. Game Session Flow (`game_session.js`)

Simulates 1000 concurrent users playing the game: starting sessions, viewing cards, submitting answers, and skipping cards.

**Target:** p95 < 200ms

```bash
k6 run -e BASE_URL=http://localhost:8080 -e AUTH_TOKEN=<token> game_session.js
```

### 2. Leaderboard Queries (`leaderboard.js`)

Concurrent reads of daily, weekly, all-time, and tournament leaderboards with realistic access patterns (40% daily, 20% weekly, 15% all-time, 15% tournament, 10% paginated). Includes a burst scenario simulating post-match traffic.

**Target:** p95 < 100ms

```bash
k6 run -e BASE_URL=http://localhost:8080 -e AUTH_TOKEN=<token> \
  -e TOURNAMENT_EVENT_ID=<uuid> leaderboard.js
```

### 3. WebSocket Connections (`websocket.js`)

Ramps up to 5000 concurrent WebSocket connections, each staying connected for 30-60 seconds and receiving broadcast events.

**Target:** connection setup p95 < 1s, error rate < 5%

```bash
k6 run -e BASE_URL=http://localhost:8080 -e AUTH_TOKEN=<token> websocket.js
```

### 4. Card Resolution (`card_resolution.js`)

Two parallel scenarios: admins resolving cards while users are actively reading data. Tests the impact of batch answer updates on concurrent read performance.

**Target:** resolution p95 < 2s, user reads p95 < 200ms

```bash
k6 run -e BASE_URL=http://localhost:8080 -e AUTH_TOKEN=<token> \
  -e ADMIN_TOKEN=<admin_token> -e CARD_IDS=<uuid1>,<uuid2>,<uuid3> \
  card_resolution.js
```

## Running All Tests

Run them sequentially:

```bash
export BASE_URL=http://localhost:8080
export AUTH_TOKEN=<token>
export ADMIN_TOKEN=<admin_token>

k6 run -e BASE_URL=$BASE_URL -e AUTH_TOKEN=$AUTH_TOKEN game_session.js
k6 run -e BASE_URL=$BASE_URL -e AUTH_TOKEN=$AUTH_TOKEN leaderboard.js
k6 run -e BASE_URL=$BASE_URL -e AUTH_TOKEN=$AUTH_TOKEN websocket.js
k6 run -e BASE_URL=$BASE_URL -e AUTH_TOKEN=$AUTH_TOKEN \
  -e ADMIN_TOKEN=$ADMIN_TOKEN -e CARD_IDS=<ids> card_resolution.js
```

## Running with Docker

```bash
docker run --rm -i --network host \
  -e BASE_URL=http://localhost:8080 \
  -e AUTH_TOKEN=<token> \
  -v $(pwd):/scripts \
  grafana/k6 run /scripts/game_session.js
```

## Output to Grafana / InfluxDB

```bash
k6 run --out influxdb=http://localhost:8086/k6 \
  -e BASE_URL=http://localhost:8080 \
  -e AUTH_TOKEN=<token> \
  game_session.js
```

## Notes

- All scripts use `AUTH_TOKEN` for authentication. In a real load test, you would need multiple distinct tokens representing different users. For multi-user testing, consider generating tokens in a setup phase or using a shared data file.
- The WebSocket test requires k6 with WebSocket support (included by default).
- Card resolution test requires pre-existing card UUIDs in the database. Create test cards before running.
- Rate limiter on the API (30 req/min per user) will affect results if using a single token. Use multiple tokens or disable the rate limiter in the test environment.
