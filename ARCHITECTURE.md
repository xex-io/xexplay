# XEX Play — System Architecture & Technical Design

> For product vision, game mechanics, and design — see [README.md](./README.md).

## Table of Contents

1. [High-Level Architecture](#1-high-level-architecture)
2. [Tech Stack Overview](#2-tech-stack-overview)
3. [Authentication & Authorization](#3-authentication--authorization)
4. [Database Design](#4-database-design)
5. [API Design](#5-api-design)
6. [Flutter App Architecture](#6-flutter-app-architecture)
7. [Go Backend Architecture](#7-go-backend-architecture)
8. [Next.js Admin Panel](#8-nextjs-admin-panel)
9. [Real-Time System](#9-real-time-system)
10. [Push Notifications](#10-push-notifications)
11. [Caching Strategy](#11-caching-strategy)
12. [Session Persistence](#12-session-persistence)
13. [Smart Shuffle Algorithm](#13-smart-shuffle-algorithm)
14. [Security Considerations](#14-security-considerations)
15. [Deployment Strategy](#15-deployment-strategy)
16. [Development Phases / Roadmap](#16-development-phases--roadmap)

---

## 1. High-Level Architecture

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                              CLIENTS                                         │
│                                                                              │
│   ┌─────────────────┐            ┌─────────────────────┐                     │
│   │  Flutter App     │            │  Next.js Admin Panel │                    │
│   │  (iOS/Android)   │            │  (Web)               │                    │
│   └────────┬────────┘            └─────────┬───────────┘                     │
│            │                                │                                │
└────────────┼────────────────────────────────┼────────────────────────────────┘
             │ HTTPS / WSS                    │ HTTPS
             │                                │
┌────────────┼────────────────────────────────┼────────────────────────────────┐
│            │                                │                                │
│            │       XEX PLAY INFRASTRUCTURE (completely separate from Exchange)│
│            │                                │                                │
│            └───────────────┬────────────────┘                                │
│                            │                                                 │
│  ┌─────────────────────────┼──────────────────────────────┐                  │
│  │               XEX PLAY GO API SERVER                    │                  │
│  │                                                        │                  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────┐         │                  │
│  │  │   REST   │  │WebSocket │  │ JWT Validator │         │                  │
│  │  │ Handlers │  │  Hub     │  │ (shared secret│         │                  │
│  │  │          │  │          │  │  w/ Exchange) │         │                  │
│  │  └────┬─────┘  └────┬─────┘  └──────┬───────┘         │                  │
│  │       │              │               │                  │                  │
│  │  ┌────┴──────────────┴───────────────┴─────┐           │                  │
│  │  │           SERVICE LAYER                 │           │                  │
│  │  │  (Business Logic / Use Cases)           │           │                  │
│  │  └────────────────┬────────────────────────┘           │                  │
│  │                   │                                    │                  │
│  │  ┌────────────────┴──────────────────────┐             │                  │
│  │  │         REPOSITORY LAYER              │             │                  │
│  │  │  (Data Access / Abstractions)         │             │                  │
│  │  └────┬──────────┬──────────┬────────────┘             │                  │
│  │       │          │          │                           │                  │
│  └───────┼──────────┼──────────┼──────────────────────────┘                  │
│          │          │          │                                              │
│  ┌───────┴──────┐ ┌─┴──────┐ ┌┴─────────────────┐                           │
│  │ PostgreSQL   │ │ Redis  │ │ Firebase (FCM)    │                           │
│  │ (XEX PLAY    │ │ (Play  │ │                   │                           │
│  │  OWN DB)     │ │  only) │ │- Push             │                           │
│  │              │ │        │ │  Notifications     │                           │
│  │- Play Users  │ │- Cache │ │                   │                           │
│  │- Cards       │ │- Leader│ │                   │                           │
│  │- Sessions    │ │  boards│ │                   │                           │
│  │- Answers     │ │- Rate  │ │                   │                           │
│  │- Events      │ │  Limits│ │                   │                           │
│  │- etc.        │ │        │ │                   │                           │
│  └──────────────┘ └────────┘ └───────────────────┘                           │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘

  ════════════════════════  DATA ISOLATION BOUNDARY  ════════════════════════

┌──────────────────────────────────────────────────────────────────────────────┐
│                                                                              │
│                     XEX EXCHANGE (separate system)                            │
│                                                                              │
│  ┌─────────────────────┐  ┌───────────────┐  ┌────────────────────────┐      │
│  │  Exchange Go API    │  │ PostgreSQL    │  │ Redis / Azure Vaults   │      │
│  │  (Gin, gRPC)        │  │ (Exchange DB) │  │ (wallets, keys, KYC)  │      │
│  │                     │  │               │  │                        │      │
│  │  Issues JWTs with   │  │ - Users       │  │ - TSS/MPC keys         │      │
│  │  JWT_SECRET ────────┼──┼── shared ─────┼──┼→ XEX Play validates    │      │
│  │                     │  │ - Wallets     │  │   with same secret     │      │
│  │                     │  │ - Orders      │  │                        │      │
│  │                     │  │ - KYC, AML    │  │                        │      │
│  └─────────────────────┘  └───────────────┘  └────────────────────────┘      │
│                                                                              │
│  XEX Play has ZERO access to Exchange DB, wallets, keys, or financial data.  │
│  The ONLY shared element is the JWT signing secret for token validation.      │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

### Data Flow Summary

1. **Flutter App** communicates with the **XEX Play Go API** via REST (HTTPS) and WebSocket (WSS).
2. **Next.js Admin Panel** communicates with the **XEX Play Go API** via REST (HTTPS).
3. **XEX Play Go API** validates user JWTs using the **shared `JWT_SECRET`** — the same secret used by XEX Exchange to sign tokens. No network call to Exchange is needed.
4. **XEX Play** has its **own PostgreSQL database** (completely separate from Exchange). It never reads or writes Exchange data.
5. **XEX Play** has its **own Redis instance** for caching, leaderboards, and rate limiting.
6. **XEX Play Go API** sends push notifications via Firebase Cloud Messaging.
7. **WebSocket** connections deliver real-time updates (card resolution, leaderboard changes).

### Why Separate Databases?

The XEX Exchange database contains encrypted private keys, wallet addresses, KYC/AML data, financial balances, and trading ledgers. Sharing a database with a game would:

- Expand the attack surface for the exchange (a game bug could expose financial data).
- Create migration conflicts (142+ exchange migrations vs. game migrations).
- Cause resource contention (game traffic during World Cup finals competing with trading queries).
- Violate data isolation principles for financial systems.

XEX Play's **only connection** to the exchange is the shared JWT signing secret, a single environment variable.

---

## 2. Tech Stack Overview

| Component            | Technology                     | Rationale                                                                     |
| -------------------- | ------------------------------ | ----------------------------------------------------------------------------- |
| **Mobile App**       | Flutter (Dart)                 | Single codebase for iOS & Android, fast UI, great DX                          |
| **Backend API**      | Go (Golang)                    | High performance, low latency, excellent concurrency                          |
| **Admin Panel**      | Next.js (React)                | Rapid development, SSR, rich ecosystem                                        |
| **Database**         | PostgreSQL                     | Reliable, supports complex queries, JSON support                              |
| **Cache/Queue**      | Redis                          | In-memory speed for leaderboards, sessions, rate limiting                     |
| **Auth**             | Shared JWT (HS256)             | Validates Exchange-issued JWTs with shared secret, zero Exchange code changes |
| **Push**             | Firebase Cloud Messaging (FCM) | Cross-platform push notifications                                             |
| **Real-Time**        | WebSocket (gorilla/websocket)  | Live updates for card resolution, leaderboards                                |
| **Containerization** | Docker + Docker Compose        | Consistent environments, easy deployment                                      |
| **CI/CD**            | GitHub Actions                 | Automated testing, building, deployment                                       |
| **Reverse Proxy**    | Nginx or Traefik               | Load balancing, TLS termination, routing                                      |

---

## 3. Authentication & Authorization

### 3.1 Shared JWT Strategy (No OAuth Server Needed)

After investigating the XEX Exchange codebase, we found that the exchange is an **OAuth client** (consuming Google/Apple auth) — it does **not** expose an OAuth 2.0 server for third-party apps. Building a full OAuth server into the exchange would be significant work and unnecessary for this use case.

Instead, XEX Play uses the **Shared JWT Secret** approach: both the Exchange API and the Play API are configured with the same `JWT_SECRET` environment variable. Tokens issued by the exchange are directly valid on the Play API, no token exchange, no network calls between services.

```
┌──────────────┐
│  Flutter App  │
│  (XEX Play)   │
└──────┬───────┘
       │
       │ 1. User taps "Login with XEX Exchange"
       │
       │ 2. App opens Exchange login
       │    (deep link / WebView)
       ↓
┌──────────────────┐
│  XEX Exchange    │
│  Login Flow      │
│                  │  User authenticates via:
│  - Magic Link    │  - Magic link (email)
│  - Google OAuth  │  - Google / Apple OAuth
│  - Apple OAuth   │  - Passkey / WebAuthn
│  - Passkey       │
│                  │
│  Exchange issues │
│  JWT (HS256)     │
│  signed with     │
│  JWT_SECRET      │──── 3. JWT returned to Flutter app
└──────────────────┘
       │
       │ 4. Flutter app sends Exchange JWT
       │    to XEX Play API
       ↓
┌──────────────────┐
│  XEX Play        │
│  Go API          │
│                  │
│  Validates JWT   │  Same JWT_SECRET → signature is valid
│  using SAME      │  Extracts: user_id, email, role
│  JWT_SECRET      │
│                  │
│  5. Upserts local│  Creates Play user record on first login
│     user record  │  (xex_user_id, display_name, email)
│                  │
│  6. Returns Play │  Play-specific session data
│     session data │  (game state, streak, etc.)
└──────────────────┘
```

### How It Works

1. User taps "Login with XEX Exchange" in the XEX Play Flutter app.
2. App opens the XEX Exchange login flow (deep link to Exchange app, or WebView fallback). The user authenticates using any method the exchange supports: magic link, Google, Apple, or passkey.
3. Exchange issues a JWT pair (access + refresh) signed with `JWT_SECRET` (HS256). The JWT contains: `user_id`, `email`, `role`, `token_type`.
4. Flutter app stores the JWT and sends it to the XEX Play API in the `Authorization: Bearer` header.
5. XEX Play API validates the JWT using the **same `JWT_SECRET`** — no network call to the exchange. On first login, it creates a local `users` record with `xex_user_id` mapped from the JWT's `sub` claim.
6. XEX Play returns game-specific data (session state, streak, leaderboard position, etc.).

### 3.2 XEX Exchange JWT Structure (as-is)

The exchange already issues JWTs with this claims structure:

```json
{
  "sub": "uuid-of-exchange-user",
  "user_id": "uuid-of-exchange-user",
  "email": "user@example.com",
  "role": "user",
  "token_type": "access",
  "iss": "nyyu",
  "iat": 1700000000,
  "exp": 1700086400
}
```

- **Access Token:** 24-hour expiry (configurable via `JWT_ACCESS_EXPIRY`).
- **Refresh Token:** 30-day expiry (configurable via `JWT_REFRESH_EXPIRY`).
- **Algorithm:** HS256 (HMAC-SHA256).
- **Secret:** Shared `JWT_SECRET` env var (minimum 32 characters).

XEX Play reads the `user_id` and `email` claims, it ignores exchange-specific fields and never needs access to exchange data.

### 3.3 Token Refresh

When the access token expires, the Flutter app uses the refresh token to get a new pair from the **Exchange API** (not Play API). XEX Play only validates tokens, it never issues or refreshes them.

```
Flutter App → POST /auth/refresh → XEX Exchange API → New token pair
Flutter App → Uses new access token → XEX Play API (validates it)
```

### 3.4 Role-Based Access Control (RBAC)

| Role      | Access                                                                     |
| --------- | -------------------------------------------------------------------------- |
| **user**  | Play game, view leaderboards, manage profile                               |
| **admin** | All user access + admin panel, card management, user management, analytics |

- The exchange JWT `role` field is `user`, `admin`, or `super_admin`.
- XEX Play maps `admin` and `super_admin` to its own admin role.
- Admin panel routes are protected by role-checking middleware.
- XEX Play also stores a `role` field in its own `users` table, which can override the exchange role (e.g., someone can be a Play admin without being an exchange admin). The Play API checks **both** the JWT role and the local role.

### 3.5 What XEX Play Does NOT Have Access To

This is critical for security:

| XEX Play CAN access | XEX Play CANNOT access      |
| ------------------- | --------------------------- |
| User ID (UUID)      | Wallets / private keys      |
| Email address       | Account balances            |
| Display name        | Trading history / orders    |
| Role (user/admin)   | KYC documents / personal ID |
| Token type & expiry | 2FA secrets / backup codes  |
|                     | Withdrawal addresses        |
|                     | Any exchange database table |

---

## 4. Database Design

> **Important:** XEX Play has its **own dedicated PostgreSQL database** (e.g., `xexplay`), completely separate from the XEX Exchange database (`nyyu`). The two databases run on separate instances with separate credentials. XEX Play has zero access to exchange tables.

### 4.1 Entity-Relationship Overview

```
users ─────────┐
               │ 1:N
               ├──→ user_sessions ──→ user_answers
               │ 1:N
               ├──→ streaks
               │ 1:N
               ├──→ user_achievements
               │ 1:N
               ├──→ referrals
               │ 1:N
               └──→ mini_league_members

events ────────┐
               │ 1:N
               └──→ matches ──→ cards ──→ daily_basket_cards

card_profiles (lookup table)

leaderboard_entries

mini_leagues ──→ mini_league_members

achievements (lookup table)
```

### 4.2 PostgreSQL Schema

```sql
-- =============================================
-- USERS
-- =============================================
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    xex_user_id     UUID UNIQUE NOT NULL,            -- Maps to 'sub'/'user_id' from Exchange JWT
    display_name    VARCHAR(100) NOT NULL,            -- Extracted from JWT on first login
    email           VARCHAR(255),                     -- Extracted from JWT on first login
    avatar_url      TEXT,
    role            VARCHAR(20) DEFAULT 'user',       -- 'user' | 'admin' (Play-local role override)
    referral_code   VARCHAR(20) UNIQUE NOT NULL,
    referred_by     UUID REFERENCES users(id),
    total_points    INTEGER DEFAULT 0,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
    -- NOTE: This is XEX Play's own users table. It does NOT reference
    -- the Exchange users table. The link is xex_user_id (from JWT claims).
);

CREATE INDEX idx_users_xex_user_id ON users(xex_user_id);
CREATE INDEX idx_users_referral_code ON users(referral_code);

-- =============================================
-- EVENTS / TOURNAMENTS
-- =============================================
CREATE TABLE events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,           -- "FIFA World Cup 2026"
    slug            VARCHAR(100) UNIQUE NOT NULL,    -- "world-cup-2026"
    description     TEXT,
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    is_active       BOOLEAN DEFAULT FALSE,
    scoring_multiplier DECIMAL(3,2) DEFAULT 1.00,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- =============================================
-- MATCHES
-- =============================================
CREATE TABLE matches (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id        UUID NOT NULL REFERENCES events(id),
    home_team       VARCHAR(100) NOT NULL,
    away_team       VARCHAR(100) NOT NULL,
    kickoff_time    TIMESTAMPTZ NOT NULL,
    status          VARCHAR(20) DEFAULT 'upcoming',  -- 'upcoming' | 'live' | 'completed' | 'cancelled'
    home_score      INTEGER,
    away_score      INTEGER,
    result_data     JSONB,                           -- Flexible result storage
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_matches_event_id ON matches(event_id);
CREATE INDEX idx_matches_kickoff_time ON matches(kickoff_time);
CREATE INDEX idx_matches_status ON matches(status);

-- =============================================
-- CARD PROFILES
-- =============================================
CREATE TABLE card_profiles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(50) UNIQUE NOT NULL,     -- 'balanced', 'lean_yes', 'lean_no', 'high_risk', 'low_risk'
    yes_points      INTEGER NOT NULL,
    no_points       INTEGER NOT NULL,
    description     TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Seed default profiles
INSERT INTO card_profiles (name, yes_points, no_points, description) VALUES
    ('balanced',  10, 10, 'Equal points for Yes and No'),
    ('lean_yes',  20, 10, 'Higher reward for Yes'),
    ('lean_no',   10, 20, 'Higher reward for No'),
    ('high_risk', 20,  5, 'High reward for Yes, low for No'),
    ('low_risk',  10,  5, 'Moderate reward, low-risk No');

-- =============================================
-- CARDS (Questions)
-- =============================================
CREATE TABLE cards (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id        UUID NOT NULL REFERENCES matches(id),
    profile_id      UUID NOT NULL REFERENCES card_profiles(id),
    question_text   TEXT NOT NULL,
    tier            VARCHAR(10) NOT NULL,             -- 'white' | 'bronze' | 'silver' | 'gold'
    correct_answer  BOOLEAN,                          -- NULL until resolved, TRUE=Yes, FALSE=No
    is_resolved     BOOLEAN DEFAULT FALSE,
    available_date  DATE NOT NULL,                    -- The day this card appears in baskets
    expires_at      TIMESTAMPTZ NOT NULL,             -- Match kickoff time
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_cards_match_id ON cards(match_id);
CREATE INDEX idx_cards_available_date ON cards(available_date);
CREATE INDEX idx_cards_tier ON cards(tier);
CREATE INDEX idx_cards_is_resolved ON cards(is_resolved);

-- =============================================
-- DAILY BASKETS
-- =============================================
CREATE TABLE daily_baskets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    basket_date     DATE NOT NULL,
    event_id        UUID NOT NULL REFERENCES events(id),
    is_published    BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_daily_baskets_date_event ON daily_baskets(basket_date, event_id);

CREATE TABLE daily_basket_cards (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    basket_id       UUID NOT NULL REFERENCES daily_baskets(id),
    card_id         UUID NOT NULL REFERENCES cards(id),
    position        INTEGER NOT NULL,                -- Order in basket (1-15)
    UNIQUE(basket_id, card_id),
    UNIQUE(basket_id, position)
);

-- =============================================
-- USER SESSIONS
-- =============================================
CREATE TABLE user_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    basket_id       UUID NOT NULL REFERENCES daily_baskets(id),
    shuffle_order   INTEGER[] NOT NULL,              -- Array of card positions in shuffled order
    current_index   INTEGER DEFAULT 0,               -- Which card the user is currently on
    answers_used    INTEGER DEFAULT 0,
    skips_used      INTEGER DEFAULT 0,
    bonus_answers   INTEGER DEFAULT 0,               -- From streaks/referrals
    bonus_skips     INTEGER DEFAULT 0,               -- From streaks/referrals
    status          VARCHAR(20) DEFAULT 'active',    -- 'active' | 'completed' | 'expired'
    started_at      TIMESTAMPTZ DEFAULT NOW(),
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_user_sessions_user_basket ON user_sessions(user_id, basket_id);
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_status ON user_sessions(status);

-- =============================================
-- USER ANSWERS
-- =============================================
CREATE TABLE user_answers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      UUID NOT NULL REFERENCES user_sessions(id),
    card_id         UUID NOT NULL REFERENCES cards(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    answer          BOOLEAN NOT NULL,                -- TRUE=Yes, FALSE=No
    points_earned   INTEGER DEFAULT 0,               -- Set after card resolution
    is_correct      BOOLEAN,                         -- NULL until resolved
    answered_at     TIMESTAMPTZ DEFAULT NOW(),
    resolved_at     TIMESTAMPTZ,
    UNIQUE(session_id, card_id)
);

CREATE INDEX idx_user_answers_user_id ON user_answers(user_id);
CREATE INDEX idx_user_answers_card_id ON user_answers(card_id);
CREATE INDEX idx_user_answers_is_correct ON user_answers(is_correct);

-- =============================================
-- LEADERBOARD ENTRIES
-- =============================================
CREATE TABLE leaderboard_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    event_id        UUID REFERENCES events(id),      -- NULL for all-time
    period_type     VARCHAR(20) NOT NULL,             -- 'daily' | 'weekly' | 'tournament' | 'all_time'
    period_key      VARCHAR(20) NOT NULL,             -- '2026-06-15' | '2026-W24' | 'world-cup-2026' | 'all_time'
    total_points    INTEGER DEFAULT 0,
    correct_answers INTEGER DEFAULT 0,
    wrong_answers   INTEGER DEFAULT 0,
    total_answers   INTEGER DEFAULT 0,
    rank            INTEGER,
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_leaderboard_user_period ON leaderboard_entries(user_id, period_type, period_key);
CREATE INDEX idx_leaderboard_ranking ON leaderboard_entries(period_type, period_key, total_points DESC, wrong_answers ASC);

-- =============================================
-- STREAKS
-- =============================================
CREATE TABLE streaks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID UNIQUE NOT NULL REFERENCES users(id),
    current_streak  INTEGER DEFAULT 0,
    longest_streak  INTEGER DEFAULT 0,
    last_played_date DATE,
    bonus_skips_earned   INTEGER DEFAULT 0,
    bonus_answers_earned INTEGER DEFAULT 0,
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- =============================================
-- ACHIEVEMENTS
-- =============================================
CREATE TABLE achievements (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key             VARCHAR(50) UNIQUE NOT NULL,      -- 'first_prediction', 'perfect_day', etc.
    name            VARCHAR(100) NOT NULL,
    description     TEXT NOT NULL,
    badge_icon      VARCHAR(255),
    condition_type  VARCHAR(50) NOT NULL,             -- 'streak', 'score', 'referral', etc.
    condition_value INTEGER NOT NULL,                 -- threshold value
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE user_achievements (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    achievement_id  UUID NOT NULL REFERENCES achievements(id),
    earned_at       TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, achievement_id)
);

CREATE INDEX idx_user_achievements_user_id ON user_achievements(user_id);

-- =============================================
-- REFERRALS
-- =============================================
CREATE TABLE referrals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    referrer_id     UUID NOT NULL REFERENCES users(id),
    referred_id     UUID NOT NULL REFERENCES users(id),
    status          VARCHAR(20) DEFAULT 'signed_up',  -- 'signed_up' | 'first_session_completed'
    reward_granted  BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(referred_id)
);

CREATE INDEX idx_referrals_referrer_id ON referrals(referrer_id);

-- =============================================
-- MINI-LEAGUES (Private Groups)
-- =============================================
CREATE TABLE mini_leagues (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL,
    invite_code     VARCHAR(20) UNIQUE NOT NULL,
    created_by      UUID NOT NULL REFERENCES users(id),
    max_members     INTEGER DEFAULT 50,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE mini_league_members (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    league_id       UUID NOT NULL REFERENCES mini_leagues(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    joined_at       TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(league_id, user_id)
);

-- =============================================
-- FCM TOKENS (Push Notifications)
-- =============================================
CREATE TABLE fcm_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    token           TEXT NOT NULL,
    device_type     VARCHAR(10) NOT NULL,             -- 'ios' | 'android'
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_fcm_tokens_user_id ON fcm_tokens(user_id);
```

---

## 5. API Design

### 5.1 Base URL & Conventions

```
Base URL:    https://api.xexplay.com/v1
Content-Type: application/json
Auth Header:  Authorization: Bearer <jwt_token>
```

**Response Envelope:**

```json
{
  "success": true,
  "data": { ... },
  "error": null,
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 150
  }
}
```

### 5.2 REST Endpoints

#### Auth

| Method | Endpoint       | Description                                             | Auth   |
| ------ | -------------- | ------------------------------------------------------- | ------ |
| POST   | `/auth/login`  | Validate Exchange JWT, upsert Play user, return profile | Public |
| POST   | `/auth/logout` | Clear Play session state                                | User   |

> **Note:** XEX Play does not issue or refresh tokens. Token refresh is handled by the Exchange API directly. The `/auth/login` endpoint simply validates the Exchange-issued JWT (shared secret), creates/updates the local Play user record on first login, and returns the user's Play profile + game state.

#### User Profile

| Method | Endpoint           | Description                      | Auth |
| ------ | ------------------ | -------------------------------- | ---- |
| GET    | `/me`              | Get current user profile         | User |
| PUT    | `/me`              | Update display name / avatar     | User |
| GET    | `/me/stats`        | Get user stats (points, streaks) | User |
| GET    | `/me/achievements` | Get user achievements            | User |
| GET    | `/me/history`      | Get past sessions & answers      | User |

#### Game Session

| Method | Endpoint                   | Description                             | Auth |
| ------ | -------------------------- | --------------------------------------- | ---- |
| POST   | `/sessions/start`          | Start or resume today's session         | User |
| GET    | `/sessions/current`        | Get current session state               | User |
| GET    | `/sessions/current/card`   | Get the current card to answer          | User |
| POST   | `/sessions/current/answer` | Submit answer (Yes/No) for current card | User |
| POST   | `/sessions/current/skip`   | Skip the current card                   | User |

#### Leaderboards

| Method | Endpoint                            | Description             | Auth |
| ------ | ----------------------------------- | ----------------------- | ---- |
| GET    | `/leaderboards/daily`               | Today's leaderboard     | User |
| GET    | `/leaderboards/weekly`              | This week's leaderboard | User |
| GET    | `/leaderboards/tournament/:eventId` | Tournament leaderboard  | User |
| GET    | `/leaderboards/all-time`            | All-time leaderboard    | User |
| GET    | `/leaderboards/friends`             | Friend leaderboard      | User |
| GET    | `/leaderboards/league/:leagueId`    | Mini-league leaderboard | User |

#### Events

| Method | Endpoint      | Description        | Auth |
| ------ | ------------- | ------------------ | ---- |
| GET    | `/events`     | List active events | User |
| GET    | `/events/:id` | Get event details  | User |

#### Social

| Method | Endpoint          | Description                       | Auth |
| ------ | ----------------- | --------------------------------- | ---- |
| GET    | `/referral/code`  | Get user's referral code          | User |
| GET    | `/referral/stats` | Get referral statistics           | User |
| POST   | `/leagues`        | Create a mini-league              | User |
| POST   | `/leagues/join`   | Join a mini-league by invite code | User |
| GET    | `/leagues`        | List user's mini-leagues          | User |
| GET    | `/leagues/:id`    | Get mini-league details           | User |

#### Push Notifications

| Method | Endpoint            | Description          | Auth |
| ------ | ------------------- | -------------------- | ---- |
| POST   | `/devices/register` | Register FCM token   | User |
| DELETE | `/devices/:token`   | Unregister FCM token | User |

#### Admin Endpoints (all require admin role)

| Method | Endpoint                     | Description                       |
| ------ | ---------------------------- | --------------------------------- |
| GET    | `/admin/events`              | List all events                   |
| POST   | `/admin/events`              | Create event                      |
| PUT    | `/admin/events/:id`          | Update event                      |
| GET    | `/admin/matches`             | List matches (filterable)         |
| POST   | `/admin/matches`             | Create match                      |
| PUT    | `/admin/matches/:id`         | Update match (including results)  |
| GET    | `/admin/cards`               | List cards (filterable)           |
| POST   | `/admin/cards`               | Create card                       |
| PUT    | `/admin/cards/:id`           | Update card                       |
| POST   | `/admin/cards/:id/resolve`   | Resolve card (set correct answer) |
| GET    | `/admin/baskets`             | List daily baskets                |
| POST   | `/admin/baskets`             | Create daily basket               |
| PUT    | `/admin/baskets/:id`         | Update basket (add/remove cards)  |
| POST   | `/admin/baskets/:id/publish` | Publish basket (make live)        |
| GET    | `/admin/users`               | List users                        |
| GET    | `/admin/users/:id`           | Get user details                  |
| PUT    | `/admin/users/:id`           | Update user (ban, role change)    |
| GET    | `/admin/analytics/overview`  | Dashboard analytics               |
| GET    | `/admin/analytics/retention` | User retention metrics            |
| GET    | `/admin/analytics/cards`     | Card answer distribution          |
| POST   | `/admin/notifications/send`  | Send custom push notification     |

### 5.3 WebSocket Endpoints

```
WSS: wss://api.xexplay.com/v1/ws
```

**Connection:** Client connects with Exchange-issued JWT as query param or in first message. Server validates using shared secret.

**Server-Sent Events:**

```json
// Card resolved
{
  "type": "card_resolved",
  "data": {
    "card_id": "uuid",
    "correct_answer": true,
    "your_answer": true,
    "points_earned": 20
  }
}

// Leaderboard update
{
  "type": "leaderboard_update",
  "data": {
    "period_type": "daily",
    "your_rank": 42,
    "your_points": 65
  }
}

// Card expiring
{
  "type": "card_expiring",
  "data": {
    "card_id": "uuid",
    "expires_in_seconds": 300
  }
}

// Session card expired (removed from active session)
{
  "type": "card_expired",
  "data": {
    "card_id": "uuid"
  }
}

// Achievement unlocked
{
  "type": "achievement_unlocked",
  "data": {
    "achievement_key": "perfect_day",
    "achievement_name": "Sharpshooter",
    "description": "All 5 predictions correct in one day"
  }
}
```

---

## 6. Flutter App Architecture

### 6.1 Feature-First Folder Structure

```
lib/
├── main.dart
├── app.dart                          # MaterialApp, routing, theme
├── core/
│   ├── constants/
│   │   ├── api_constants.dart
│   │   ├── app_colors.dart
│   │   └── app_strings.dart
│   ├── errors/
│   │   ├── exceptions.dart
│   │   └── failures.dart
│   ├── network/
│   │   ├── api_client.dart           # Dio HTTP client
│   │   ├── websocket_client.dart     # WebSocket connection
│   │   └── interceptors/
│   │       ├── auth_interceptor.dart
│   │       └── error_interceptor.dart
│   ├── storage/
│   │   └── secure_storage.dart       # JWT token storage
│   ├── routing/
│   │   └── app_router.dart           # GoRouter
│   └── theme/
│       └── app_theme.dart
├── features/
│   ├── auth/
│   │   ├── data/
│   │   │   ├── auth_repository.dart
│   │   │   └── auth_remote_source.dart
│   │   ├── domain/
│   │   │   └── auth_state.dart
│   │   ├── presentation/
│   │   │   ├── login_screen.dart
│   │   │   └── widgets/
│   │   └── providers/
│   │       └── auth_provider.dart
│   ├── game/
│   │   ├── data/
│   │   │   ├── session_repository.dart
│   │   │   └── card_models.dart
│   │   ├── domain/
│   │   │   ├── session_state.dart
│   │   │   └── card_entity.dart
│   │   ├── presentation/
│   │   │   ├── game_screen.dart
│   │   │   ├── card_widget.dart      # Swipeable card
│   │   │   ├── timer_widget.dart     # 40s countdown
│   │   │   ├── session_summary.dart
│   │   │   └── widgets/
│   │   └── providers/
│   │       └── game_provider.dart
│   ├── leaderboard/
│   │   ├── data/
│   │   ├── presentation/
│   │   │   ├── leaderboard_screen.dart
│   │   │   └── widgets/
│   │   └── providers/
│   ├── profile/
│   │   ├── data/
│   │   ├── presentation/
│   │   │   ├── profile_screen.dart
│   │   │   ├── achievements_screen.dart
│   │   │   ├── history_screen.dart
│   │   │   └── widgets/
│   │   └── providers/
│   ├── social/
│   │   ├── data/
│   │   ├── presentation/
│   │   │   ├── referral_screen.dart
│   │   │   ├── mini_leagues_screen.dart
│   │   │   └── widgets/
│   │   └── providers/
│   └── notifications/
│       ├── data/
│       ├── services/
│       │   └── fcm_service.dart
│       └── providers/
└── shared/
    ├── widgets/
    │   ├── loading_widget.dart
    │   ├── error_widget.dart
    │   └── bottom_nav.dart
    └── utils/
        ├── date_utils.dart
        └── share_utils.dart
```

### 6.2 State Management: Riverpod

Riverpod is chosen for its compile-time safety, testability, and provider-based architecture.

**Key Providers:**

```dart
// Auth state
final authProvider = StateNotifierProvider<AuthNotifier, AuthState>(...);

// Current game session
final gameSessionProvider = StateNotifierProvider<GameNotifier, GameState>(...);

// Leaderboard data
final leaderboardProvider = FutureProvider.family<LeaderboardData, LeaderboardType>(...);

// WebSocket stream
final wsEventsProvider = StreamProvider<WsEvent>(...);

// User profile
final userProfileProvider = FutureProvider<UserProfile>(...);
```

### 6.3 Navigation

Using **GoRouter** for declarative, deep-link-ready routing:

```
/                       → Splash / Auth check
/login                  → Login via XEX Exchange (deep link / WebView)
/home                   → Main screen (daily basket overview)
/game                   → Active game session (card swipe)
/game/summary           → Session summary
/leaderboard            → Leaderboard tabs (daily/weekly/tournament/all-time)
/leaderboard/:type      → Specific leaderboard
/profile                → User profile
/profile/achievements   → Achievements list
/profile/history        → Session history
/social/referral        → Referral screen
/social/leagues         → Mini-leagues list
/social/leagues/:id     → Mini-league detail
```

### 6.4 Offline Support

- Session state is cached locally using **Hive** or **SharedPreferences**.
- If the user loses connectivity mid-session, the app preserves local state.
- On reconnect, the app syncs with the server to resume from the correct position.
- Cards with expired timers during offline are handled on reconnect (server is source of truth).

### 6.5 Card Swipe Animation

The card swipe uses Flutter's **Draggable** or a library like `flutter_card_swiper`:

- **Right swipe** → Green overlay → "YES" confirmed
- **Left swipe** → Red overlay → "NO" confirmed
- **Down swipe / tap button** → Gray overlay → "SKIP" confirmed
- Spring-back animation if swipe is not committed (threshold not reached)

---

## 7. Go Backend Architecture

### 7.1 Project Layout (Clean Architecture)

```
xexplay-api/
├── cmd/
│   └── server/
│       └── main.go                   # Entry point
├── internal/
│   ├── config/
│   │   └── config.go                 # Env vars, configuration (incl. shared JWT_SECRET)
│   ├── domain/                       # Domain entities (no dependencies)
│   │   ├── user.go
│   │   ├── event.go
│   │   ├── match.go
│   │   ├── card.go
│   │   ├── card_profile.go
│   │   ├── basket.go
│   │   ├── session.go
│   │   ├── answer.go
│   │   ├── leaderboard.go
│   │   ├── streak.go
│   │   ├── achievement.go
│   │   ├── referral.go
│   │   └── mini_league.go
│   ├── repository/                   # Data access interfaces + implementations
│   │   ├── interfaces.go            # Repository interfaces
│   │   ├── postgres/
│   │   │   ├── user_repo.go
│   │   │   ├── event_repo.go
│   │   │   ├── match_repo.go
│   │   │   ├── card_repo.go
│   │   │   ├── basket_repo.go
│   │   │   ├── session_repo.go
│   │   │   ├── answer_repo.go
│   │   │   ├── leaderboard_repo.go
│   │   │   ├── streak_repo.go
│   │   │   ├── achievement_repo.go
│   │   │   ├── referral_repo.go
│   │   │   └── mini_league_repo.go
│   │   └── redis/
│   │       ├── cache_repo.go
│   │       ├── leaderboard_cache.go
│   │       └── session_cache.go
│   ├── service/                      # Business logic (use cases)
│   │   ├── auth_service.go
│   │   ├── game_service.go          # Session start, answer, skip logic
│   │   ├── card_service.go          # Card resolution
│   │   ├── leaderboard_service.go
│   │   ├── streak_service.go
│   │   ├── achievement_service.go
│   │   ├── referral_service.go
│   │   ├── notification_service.go
│   │   └── shuffle_service.go       # Smart Shuffle Algorithm
│   ├── handler/                      # HTTP handlers (controllers)
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── game_handler.go
│   │   ├── leaderboard_handler.go
│   │   ├── social_handler.go
│   │   ├── device_handler.go
│   │   ├── admin/
│   │   │   ├── event_handler.go
│   │   │   ├── match_handler.go
│   │   │   ├── card_handler.go
│   │   │   ├── basket_handler.go
│   │   │   ├── user_handler.go
│   │   │   ├── analytics_handler.go
│   │   │   └── notification_handler.go
│   │   └── ws/
│   │       └── websocket_handler.go
│   ├── middleware/
│   │   ├── auth.go                  # Exchange JWT validation (shared secret)
│   │   ├── admin.go                 # Admin role check
│   │   ├── rate_limiter.go
│   │   ├── cors.go
│   │   ├── logger.go
│   │   └── recovery.go
│   └── pkg/
│       ├── jwt/
│       │   └── jwt.go
│       ├── response/
│       │   └── response.go          # Standard API response helpers
│       ├── validator/
│       │   └── validator.go
│       └── ws/
│           ├── hub.go               # WebSocket hub (connection manager)
│           └── client.go            # WebSocket client
├── migrations/
│   ├── 001_initial_schema.up.sql
│   ├── 001_initial_schema.down.sql
│   └── ...
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
├── go.mod
├── go.sum
└── Makefile
```

### 7.2 Key Libraries

| Purpose     | Library                        |
| ----------- | ------------------------------ |
| HTTP Router | `gin` (matches Exchange stack) |
| Database    | `pgx` (PostgreSQL driver)      |
| Migrations  | `golang-migrate`               |
| Redis       | `go-redis/redis`               |
| WebSocket   | `gorilla/websocket`            |
| JWT         | `golang-jwt/jwt`               |
| Validation  | `go-playground/validator`      |
| Config      | `viper` or env vars            |
| Logging     | `zerolog` or `zap`             |
| Testing     | `testify`                      |

### 7.3 Middleware Chain

```
Request → Logger → Recovery → CORS → RateLimiter → Auth (JWT) → [Admin Check] → Handler
```

### 7.4 Game Service — Core Logic

The `game_service.go` contains the most critical business logic:

```go
// StartSession creates or resumes a user's daily session
func (s *GameService) StartSession(ctx context.Context, userID uuid.UUID) (*Session, error)

// GetCurrentCard returns the next card for the user to answer
func (s *GameService) GetCurrentCard(ctx context.Context, sessionID uuid.UUID) (*Card, error)

// SubmitAnswer records the user's Yes/No answer
func (s *GameService) SubmitAnswer(ctx context.Context, sessionID uuid.UUID, answer bool) (*AnswerResult, error)

// SkipCard skips the current card
func (s *GameService) SkipCard(ctx context.Context, sessionID uuid.UUID) error

// ResolveCard sets the correct answer and scores all user answers
func (s *GameService) ResolveCard(ctx context.Context, cardID uuid.UUID, correctAnswer bool) error
```

---

## 8. Next.js Admin Panel

### 8.1 Pages

```
/login                      → Admin login (Exchange JWT, admin role required)
/dashboard                  → Analytics overview (DAU, sessions, conversions)
/events                     → Event CRUD list
/events/[id]                → Event detail & matches
/matches                    → Match list (filter by event, date, status)
/matches/[id]               → Match detail, enter results
/cards                      → Card list (filter by date, tier, status)
/cards/new                  → Create card form
/cards/[id]                 → Edit card, resolve card
/baskets                    → Daily basket list
/baskets/[date]             → Basket builder (add/remove/reorder cards)
/users                      → User list (search, filter, paginate)
/users/[id]                 → User detail (sessions, answers, stats)
/leaderboards               → Leaderboard viewer & export
/notifications              → Send push notifications
/settings                   → Scoring profiles, app config
```

### 8.2 Key Features

- **Basket Builder:** Drag-and-drop interface for composing daily baskets of 15 cards.
- **Card Resolution Panel:** After a match completes, admin sets the correct answer. System automatically resolves all user answers and updates leaderboards.
- **Live Dashboard:** Real-time metrics showing active sessions, answers being submitted, leaderboard changes.
- **User Moderation:** Ban/suspend users, view detailed activity logs.
- **Notification Composer:** Send targeted push notifications to all users, specific segments, or individuals.

### 8.3 Tech Details

- **Framework:** Next.js 14+ (App Router)
- **UI Library:** Tailwind CSS + shadcn/ui components
- **State:** React Query (TanStack Query) for server state
- **Auth:** Validates Exchange JWT (shared secret), checks admin role in JWT + local DB
- **Charts:** Recharts or Chart.js for analytics dashboards

---

## 9. Real-Time System

### 9.1 WebSocket Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Flutter #1  │     │  Flutter #2  │     │  Flutter #N  │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │ WSS                │ WSS                │ WSS
       └────────────────────┼────────────────────┘
                            │
                   ┌────────┴────────┐
                   │   WebSocket Hub │
                   │   (Go Server)   │
                   │                 │
                   │  Clients Map    │
                   │  [userID →      │
                   │   *Connection]  │
                   └────────┬────────┘
                            │
              ┌─────────────┼─────────────┐
              │             │             │
       Card Resolved  Leaderboard   Card Expiring
         Event         Update         Event
```

### 9.2 Event Flow — Card Resolution

1. Admin marks match as completed and enters results.
2. Admin resolves each card (sets correct answer = Yes/No).
3. `CardService.ResolveCard()` runs:
   - Fetches all `user_answers` for this card.
   - Calculates points for each correct answer.
   - Updates `user_answers.points_earned` and `user_answers.is_correct`.
   - Updates `leaderboard_entries` for all affected users.
   - Updates Redis leaderboard cache.
4. WebSocket Hub broadcasts `card_resolved` events to all connected users who answered that card.
5. Push notifications sent to users who answered (via FCM).

### 9.3 Event Flow — Card Expiry

1. A background goroutine monitors cards approaching their `expires_at` time.
2. At `expires_at - 30 minutes`: sends `card_expiring` WebSocket event.
3. At `expires_at`: sends `card_expired` WebSocket event.
4. Server removes the card from active sessions (users who haven't seen it yet lose it).

---

## 10. Push Notifications

### 10.1 Firebase Cloud Messaging Setup

```
┌──────────────┐     ┌──────────────┐     ┌───────────────┐
│  Go Server   │────→│  Firebase    │────→│  APNs / FCM   │
│              │     │  Admin SDK   │     │  (iOS/Android) │
└──────────────┘     └──────────────┘     └───────────────┘
```

### 10.2 Notification Types

| Type                | Trigger                       | Priority |
| ------------------- | ----------------------------- | -------- |
| Daily basket ready  | Cron job (morning)            | Normal   |
| Card expiring soon  | 30 min before match kickoff   | High     |
| Prediction resolved | Card resolved by admin        | High     |
| Streak at risk      | Evening if user hasn't played | Normal   |
| Friend joined       | Referral signup               | Normal   |
| Rank change         | Leaderboard recalculation     | Low      |
| Custom (admin)      | Admin sends manually          | Variable |

### 10.3 Implementation

- FCM tokens are stored in the `fcm_tokens` table.
- A notification service in Go handles token management and message dispatch.
- Batch sending for broadcast notifications (e.g., daily basket ready).
- Individual sending for user-specific notifications (e.g., prediction resolved).
- Token cleanup: mark inactive tokens that return errors from FCM.

---

## 11. Caching Strategy

### 11.1 Redis Usage Map

| Key Pattern                        | Data                        | TTL            |
| ---------------------------------- | --------------------------- | -------------- |
| `leaderboard:daily:{date}`         | Sorted set (userID → score) | 48 hours       |
| `leaderboard:weekly:{week}`        | Sorted set (userID → score) | 2 weeks        |
| `leaderboard:tournament:{eventId}` | Sorted set (userID → score) | Event duration |
| `leaderboard:alltime`              | Sorted set (userID → score) | Persistent     |
| `session:{userId}:{date}`          | Hash (session state)        | 24 hours       |
| `basket:{date}`                    | JSON (basket cards)         | 24 hours       |
| `user:profile:{userId}`            | JSON (user profile)         | 1 hour         |
| `rate_limit:{userId}:{endpoint}`   | Counter                     | 1 minute       |

### 11.2 Leaderboard Implementation

Redis Sorted Sets are ideal for leaderboards:

```
ZADD leaderboard:daily:2026-06-15 65 "user_uuid_1"
ZADD leaderboard:daily:2026-06-15 80 "user_uuid_2"

// Get top 100
ZREVRANGE leaderboard:daily:2026-06-15 0 99 WITHSCORES

// Get user's rank
ZREVRANK leaderboard:daily:2026-06-15 "user_uuid_1"
```

### 11.3 Cache Invalidation

- Leaderboard caches are updated **in real-time** when cards are resolved (write-through).
- Session cache is updated on every user action (answer/skip).
- User profile cache is invalidated on profile update.
- Basket cache is invalidated when admin publishes a new basket.

---

## 12. Session Persistence

### 12.1 Server-Side Session State

The user's game session is stored in both PostgreSQL (durable) and Redis (fast access):

```json
// Redis: session:{userId}:{date}
{
  "session_id": "uuid",
  "basket_id": "uuid",
  "shuffle_order": [3, 7, 1, 12, 5, 9, 2, 14, 6, 10, 4, 8, 11, 13, 15],
  "current_index": 4,
  "answers_used": 3,
  "skips_used": 1,
  "bonus_answers": 0,
  "bonus_skips": 1,
  "status": "active"
}
```

### 12.2 Resume Flow

1. User opens app → calls `POST /sessions/start`.
2. Server checks Redis for active session.
3. If found: returns session state, user continues from `current_index`.
4. If not in Redis but in PostgreSQL: restores to Redis, returns state.
5. If no session exists: creates new session with shuffled card order.

### 12.3 Disconnect Handling

- The **server is the source of truth**. The Flutter app caches session state locally but always syncs with the server on reconnect.
- If a user's connection drops mid-session, their progress is safe, the last confirmed action (answer/skip) was already persisted server-side.
- On reconnect, the app calls `GET /sessions/current` to get the authoritative state.
- Cards that expired during the disconnect are skipped automatically (no resource consumed).

---

## 13. Smart Shuffle Algorithm

### 13.1 Purpose

The shuffle algorithm ensures fairness: every user sees a balanced distribution of card tiers in their first 8 actionable positions, while the order remains unpredictable.

### 13.2 Algorithm

```
Input:  15 cards in today's basket, categorized by tier
Output: Ordered list of 15 card positions for this user

Step 1: Categorize cards by tier
  gold_cards   = cards where tier == 'gold'     (typically 1-2)
  silver_cards = cards where tier == 'silver'   (typically 2-3)
  bronze_cards = cards where tier == 'bronze'   (typically 3-4)
  white_cards  = cards where tier == 'white'    (typically 6-8)

Step 2: Select guaranteed cards for first 8 positions
  selected = []
  selected += random_sample(gold_cards, 1)
  selected += random_sample(silver_cards, 2)
  selected += random_sample(bronze_cards, 2)
  selected += random_sample(white_cards, 3)
  // Total: 8 cards

Step 3: Shuffle the 8 selected cards randomly
  shuffle(selected)

Step 4: Remaining cards (15 - 8 = 7) go after position 8
  remaining = all_cards - selected
  shuffle(remaining)

Step 5: Final order = selected + remaining

Step 6: Store the shuffle order in the session
```

### 13.3 Fairness Guarantees

- Every user is guaranteed to see exactly **1 Gold, 2 Silver, 2 Bronze, 3 White** cards before they exhaust their 8 actions.
- The order within those 8 positions is random, so no user gets the Gold card first or last predictably.
- The remaining 7 cards act as a buffer, they're only reached if the user has actions left (e.g., streak bonuses).
- This is not disclosed to the user.

---

## 14. Security Considerations

### 14.1 Rate Limiting

| Endpoint Category  | Limit        | Window   |
| ------------------ | ------------ | -------- |
| Auth endpoints     | 10 requests  | 1 minute |
| Game actions       | 30 requests  | 1 minute |
| Leaderboard reads  | 60 requests  | 1 minute |
| Admin endpoints    | 100 requests | 1 minute |
| WebSocket messages | 20 messages  | 1 minute |

Implemented using Redis counters with TTL.

### 14.2 Anti-Cheat Measures

- **Server-side validation:** All game logic runs on the server. The client only sends "answer Yes/No" or "skip", the server determines which card is being answered.
- **Timer enforcement:** The 40-second timer is enforced server-side. Client-side timer is for UX only.
- **Action sequence validation:** The server tracks `current_index` and rejects out-of-order actions.
- **Resource validation:** The server validates remaining answers/skips before processing an action.
- **Session uniqueness:** Only one active session per user per day. Duplicate session starts return the existing session.
- **Answer immutability:** Once an answer is recorded, it cannot be changed.

### 14.3 Input Validation

- All API inputs are validated using `go-playground/validator`.
- UUIDs, enums, string lengths, and data types are strictly checked.
- SQL injection prevention via parameterized queries (pgx).
- XSS prevention: admin-created card text is sanitized before serving to clients.

### 14.4 CORS Configuration

```
Allowed Origins: https://admin.xexplay.com (admin panel only)
Allowed Methods: GET, POST, PUT, DELETE, OPTIONS
Allowed Headers: Authorization, Content-Type
Max Age:         86400 (24 hours)
```

Mobile apps don't use CORS (native HTTP), but the admin panel does.

### 14.5 Data Isolation from XEX Exchange

This is the most critical security property of the architecture:

- **Separate database:** XEX Play has its own PostgreSQL instance with its own credentials. It cannot query, read, or write any Exchange table.
- **Separate Redis:** XEX Play has its own Redis instance. No shared cache keys.
- **No network access:** XEX Play API has no network route to the Exchange database or internal services.
- **Shared JWT secret is the only link:** The `JWT_SECRET` environment variable is the single shared credential. If XEX Play is compromised, an attacker gets:
  - The ability to forge JWTs (mitigated by secret rotation and monitoring).
  - Access to XEX Play game data only (cards, scores, leaderboards).
  - **No access** to exchange wallets, balances, private keys, KYC data, or trading systems.
- **Secret rotation:** If the shared JWT secret needs rotation, both services update their `JWT_SECRET` env var simultaneously during a maintenance window.

### 14.6 Additional Security

- **HTTPS everywhere** — TLS termination at the reverse proxy.
- **JWT secrets** rotated periodically (coordinated with Exchange team).
- **Database credentials** stored in environment variables, never in code.
- **Admin actions** logged with audit trail (who did what, when).
- **FCM server key** stored securely, not exposed to clients.
- **Dependency scanning** via GitHub Dependabot / Snyk.

---

## 15. Deployment Strategy

### 15.1 Infrastructure

```
┌───────────────────────────────────────────────────────────────┐
│                   XEX Play — Production                        │
│                                                               │
│  ┌────────────┐   ┌────────────┐                              │
│  │  Go API    │   │  Go API    │  (2+ replicas)               │
│  │  Container │   │  Container │                               │
│  └─────┬──────┘   └─────┬──────┘                              │
│        └────────┬────────┘                                    │
│           ┌─────┴─────┐                                       │
│           │  Nginx /  │                                       │
│           │  Traefik  │                                       │
│           └───────────┘                                       │
│                                                               │
│  ┌─────────────────┐   ┌──────────────┐                      │
│  │ PostgreSQL      │   │   Redis      │                       │
│  │ (xexplay DB)    │   │ (Play only)  │                       │
│  │ SEPARATE from   │   │              │                       │
│  │ Exchange DB     │   │              │                       │
│  └─────────────────┘   └──────────────┘                      │
│                                                               │
│  ┌────────────────┐                                           │
│  │  Next.js Admin │  (Static + SSR)                          │
│  │  Container     │                                           │
│  └────────────────┘                                           │
│                                                               │
│  Shared env: JWT_SECRET (same value as Exchange)              │
└───────────────────────────────────────────────────────────────┘
```

### 15.2 Docker Configuration

Each service has its own Dockerfile:

- `docker/api.Dockerfile` — Multi-stage Go build
- `docker/admin.Dockerfile` — Next.js build
- `docker/docker-compose.yml` — Full local development stack
- `docker/docker-compose.prod.yml` — Production overrides

### 15.3 Environments

| Environment    | Purpose             | Database                                              |
| -------------- | ------------------- | ----------------------------------------------------- |
| **local**      | Developer machine   | Local PostgreSQL (`xexplay`)                          |
| **staging**    | Pre-release testing | Staging DB (separate instance)                        |
| **production** | Live users          | Production DB (separate instance, no Exchange access) |

### 15.4 CI/CD Pipeline (GitHub Actions)

```
Push to main branch
    ↓
Run unit tests (Go, Flutter, Next.js)
    ↓
Run linters (golangci-lint, flutter analyze, eslint)
    ↓
Build Docker images
    ↓
Push to container registry
    ↓
Deploy to staging (auto)
    ↓
Manual approval gate
    ↓
Deploy to production
```

### 15.5 Database Migrations

- Managed by `golang-migrate`.
- Migrations are versioned SQL files in `migrations/`.
- Applied automatically on deployment (before the new API version starts serving traffic).
- Rollback scripts (`*.down.sql`) for every migration.

---

## 16. Development Phases / Roadmap

### Phase 1: MVP (4–6 weeks)

**Goal:** Core game loop playable end-to-end.

| Component    | Scope                                                                            |
| ------------ | -------------------------------------------------------------------------------- |
| **Backend**  | Auth (shared JWT validation), cards, baskets, sessions, answers, card resolution |
| **Flutter**  | Login, game session (swipe cards), basic session summary                         |
| **Admin**    | Create events, matches, cards, baskets; resolve cards                            |
| **Database** | Core tables: users, events, matches, cards, baskets, sessions, answers           |
| **Infra**    | Docker Compose local dev, PostgreSQL, Redis                                      |

**Not included:** Leaderboards, streaks, achievements, referrals, push notifications, WebSocket.

### Phase 2: v1.0 — Competition & Engagement (3–4 weeks)

**Goal:** Full competitive experience with leaderboards and streaks.

| Component    | Scope                                                            |
| ------------ | ---------------------------------------------------------------- |
| **Backend**  | Leaderboard service, streak system, push notifications (FCM)     |
| **Flutter**  | Leaderboard screens, streak display, push notification handling  |
| **Admin**    | Leaderboard management, analytics dashboard, notification sender |
| **Database** | Leaderboard entries, streaks, FCM tokens tables                  |
| **Infra**    | Firebase project setup, staging environment                      |

### Phase 3: v1.5 — Social & Growth (3–4 weeks)

**Goal:** Viral growth features and social engagement.

| Component    | Scope                                                                                 |
| ------------ | ------------------------------------------------------------------------------------- |
| **Backend**  | Referral system, achievements, mini-leagues, WebSocket                                |
| **Flutter**  | Referral screen, achievements/badges, mini-leagues, social sharing, real-time updates |
| **Admin**    | User moderation tools, referral analytics                                             |
| **Database** | Achievements, referrals, mini-leagues tables                                          |
| **Infra**    | WebSocket infrastructure, deep linking                                                |

### Phase 4: v2.0 — Exchange Integration & Scale (3–4 weeks)

**Goal:** Full XEX Exchange integration and production readiness.

| Component   | Scope                                                         |
| ----------- | ------------------------------------------------------------- |
| **Backend** | Exchange integration APIs, points redemption, exclusive cards |
| **Flutter** | Exchange prompts, reward redemption, trader benefits display  |
| **Admin**   | Prize pool management, exchange metrics                       |
| **Infra**   | Production deployment, monitoring, alerting, load testing     |

### Total Estimated Timeline: 13–18 weeks

```
Week:  1──2──3──4──5──6──7──8──9──10──11──12──13──14──15──16──17──18
       ├───── MVP ──────┤
                        ├────── v1.0 ──────┤
                                           ├───── v1.5 ──────┤
                                                             ├──── v2.0 ────┤
```

---

_This document serves as the definitive technical reference for XEX Play. All implementation decisions should align with the architecture and patterns described here._
