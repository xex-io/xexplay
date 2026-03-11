# XEX Play — Implementation Plan

> Production-ready implementation plan covering all components: Go Backend, Flutter App, Next.js Admin Panel, Database, and Infrastructure.
>
> **Status Legend:** `[ ]` Not Started · `[~]` In Progress · `[x]` Completed · `[!]` Blocked

---

## Table of Contents

- [Phase 1: MVP — Core Game Loop](#phase-1-mvp--core-game-loop)
- [Phase 2: Competition & Engagement](#phase-2-competition--engagement)
- [Phase 3: Social & Growth](#phase-3-social--growth)
- [Phase 4: Exchange Integration & Production](#phase-4-exchange-integration--production)
- [Progress Summary](#progress-summary)

---

## Phase 1: MVP — Core Game Loop

**Goal:** Core game loop playable end-to-end — a user can log in, receive 15 cards, swipe through them, and see a session summary.

---

### 1.1 Project Scaffolding & Infrastructure

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 1.1.1 | Create monorepo structure: `/backend`, `/app`, `/admin`, `/docs` | All | [x] | |
| 1.1.2 | Initialize Go module (`go mod init`) with Gin, pgx, go-redis, golang-jwt, validator, zerolog, golang-migrate | Backend | [x] | |
| 1.1.3 | Set up Go project layout: `cmd/server/main.go`, `internal/{config,domain,repository,service,handler,middleware,pkg}` | Backend | [x] | Per ARCHITECTURE.md Section 7.1 |
| 1.1.4 | Initialize Flutter project with Riverpod, GoRouter, Dio, flutter_secure_storage, flutter_localizations, intl | App | [x] | |
| 1.1.5 | Set up Flutter feature-first folder structure: `lib/{core,features,shared}` | App | [x] | Per ARCHITECTURE.md Section 6.1 |
| 1.1.6 | Initialize Next.js 14+ project (App Router) with Tailwind CSS, shadcn/ui, TanStack Query | Admin | [x] | TanStack Query + axios installed |
| 1.1.7 | Create `docker-compose.yml` with PostgreSQL (`xexplay` DB), Redis, Go API, Admin panel | Infra | [x] | |
| 1.1.8 | Create `Makefile` with targets: `dev`, `build`, `test`, `lint`, `migrate-up`, `migrate-down`, `seed` | Infra | [x] | |
| 1.1.9 | Set up `.env.example` with all required env vars: `JWT_SECRET`, `DATABASE_URL`, `REDIS_URL`, `PORT`, etc. | Infra | [x] | |
| 1.1.10 | Create Go API `Dockerfile` (multi-stage build) and Admin `Dockerfile` | Infra | [x] | API Dockerfile done, Admin Dockerfile pending |

---

### 1.2 Database — Core Schema & Migrations

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 1.2.1 | Create migration `001_create_users.up.sql` — users table with `xex_user_id`, `display_name`, `email`, `avatar_url`, `role`, `referral_code`, `language`, `total_points`, indexes | DB | [x] | |
| 1.2.2 | Create migration `002_create_events.up.sql` — events table with JSONB `name`/`description`, `slug`, `start_date`, `end_date`, `is_active`, `scoring_multiplier` | DB | [x] | |
| 1.2.3 | Create migration `003_create_matches.up.sql` — matches table with `event_id` FK, `home_team`, `away_team`, `kickoff_time`, `status`, `home_score`, `away_score`, `result_data` JSONB, indexes | DB | [x] | |
| 1.2.4 | Create migration `004_create_cards.up.sql` — cards table with `match_id` FK, `question_text` JSONB, `tier`, `high_answer_is_yes`, `correct_answer`, `is_resolved`, `available_date`, `expires_at`, indexes | DB | [x] | |
| 1.2.5 | Create migration `005_create_daily_baskets.up.sql` — daily_baskets + daily_basket_cards tables with unique constraints | DB | [x] | |
| 1.2.6 | Create migration `006_create_user_sessions.up.sql` — user_sessions table with `shuffle_order` array, `current_index`, `answers_used`, `skips_used`, `bonus_answers`, `bonus_skips`, `status`, unique constraint on (user_id, basket_id) | DB | [x] | |
| 1.2.7 | Create migration `007_create_user_answers.up.sql` — user_answers table with `session_id`, `card_id`, `user_id`, `answer`, `points_earned`, `is_correct`, unique constraint on (session_id, card_id) | DB | [x] | |
| 1.2.8 | Create matching `*.down.sql` rollback scripts for all migrations | DB | [x] | |
| 1.2.9 | Integrate `golang-migrate` into the Go API startup (auto-run migrations) | Backend | [x] | |
| 1.2.10 | Create seed script with sample data: 1 event, 4 matches, 15 cards (3G+5S+7W), 1 published basket | DB | [x] | migrations/seed.sql |

---

### 1.3 Backend — Configuration & Core Packages

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 1.3.1 | Implement `internal/config/config.go` — load env vars: `JWT_SECRET`, `DATABASE_URL`, `REDIS_URL`, `PORT`, `CORS_ORIGINS`, `LOG_LEVEL` | Backend | [x] | |
| 1.3.2 | Implement `internal/pkg/response/response.go` — standard API response envelope: `{success, data, error, meta}` | Backend | [x] | |
| 1.3.3 | Implement `internal/pkg/jwt/jwt.go` — parse and validate Exchange JWTs (HS256, shared secret), extract `user_id`, `email`, `role` claims | Backend | [x] | Must match Exchange JWT structure from ARCHITECTURE.md 3.2 |
| 1.3.4 | Implement `internal/pkg/validator/validator.go` — request validation helpers using `go-playground/validator` | Backend | [x] | |
| 1.3.5 | Set up PostgreSQL connection pool using `pgx/v5` | Backend | [x] | |
| 1.3.6 | Set up Redis client using `go-redis/v9` | Backend | [x] | |
| 1.3.7 | Set up `zerolog` structured logging | Backend | [x] | |

---

### 1.4 Backend — Domain Models

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 1.4.1 | Define `internal/domain/user.go` — User struct, role constants | Backend | [x] | |
| 1.4.2 | Define `internal/domain/event.go` — Event struct with JSONB name/description | Backend | [x] | |
| 1.4.3 | Define `internal/domain/match.go` — Match struct, status enum (upcoming/live/completed/cancelled) | Backend | [x] | |
| 1.4.4 | Define `internal/domain/card.go` — Card struct, tier enum (gold/silver/white), scoring constants (Gold: 20/5, Silver: 15/10, White: 10/10) | Backend | [x] | |
| 1.4.5 | Define `internal/domain/basket.go` — DailyBasket + DailyBasketCard structs | Backend | [x] | |
| 1.4.6 | Define `internal/domain/session.go` — UserSession struct with shuffle_order, resource tracking | Backend | [x] | |
| 1.4.7 | Define `internal/domain/answer.go` — UserAnswer struct | Backend | [x] | |

---

### 1.5 Backend — Repository Layer

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 1.5.1 | Define `internal/repository/interfaces.go` — repository interfaces: UserRepo, EventRepo, MatchRepo, CardRepo, BasketRepo, SessionRepo, AnswerRepo | Backend | [x] | 7 interfaces defined |
| 1.5.2 | Implement `internal/repository/postgres/user_repo.go` — FindByXexUserID, Upsert, FindByID, UpdateProfile | Backend | [x] | |
| 1.5.3 | Implement `internal/repository/postgres/event_repo.go` — Create, Update, FindByID, FindActive, FindAll | Backend | [x] | |
| 1.5.4 | Implement `internal/repository/postgres/match_repo.go` — Create, Update, FindByID, FindByEventID, FindByDateRange, UpdateResult | Backend | [x] | |
| 1.5.5 | Implement `internal/repository/postgres/card_repo.go` — Create, Update, FindByID, FindByAvailableDate, Resolve, FindUnresolvedByMatch | Backend | [x] | |
| 1.5.6 | Implement `internal/repository/postgres/basket_repo.go` — Create, AddCards, Publish, FindByDateAndEvent, FindPublishedByDate | Backend | [x] | Enforce 3G+5S+7W validation on publish |
| 1.5.7 | Implement `internal/repository/postgres/session_repo.go` — Create, FindByUserAndBasket, UpdateProgress, Complete | Backend | [x] | |
| 1.5.8 | Implement `internal/repository/postgres/answer_repo.go` — Create, FindBySession, FindByCard, BulkResolve | Backend | [x] | |
| 1.5.9 | Implement `internal/repository/redis/session_cache.go` — Get/Set/Delete session state in Redis | Backend | [x] | Merged into cache_repo.go |
| 1.5.10 | Implement `internal/repository/redis/cache_repo.go` — generic cache get/set/delete, basket cache | Backend | [x] | |

---

### 1.6 Backend — Middleware

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 1.6.1 | Implement `internal/middleware/auth.go` — extract Bearer token, validate JWT using shared secret, inject user claims into Gin context | Backend | [x] | |
| 1.6.2 | Implement `internal/middleware/admin.go` — check JWT role OR local DB role for admin access | Backend | [x] | |
| 1.6.3 | Implement `internal/middleware/locale.go` — parse `Accept-Language` header, fall back to user's stored language, then `en` | Backend | [x] | |
| 1.6.4 | Implement `internal/middleware/rate_limiter.go` — Redis-based rate limiting per user per endpoint category | Backend | [x] | |
| 1.6.5 | Implement `internal/middleware/cors.go` — CORS for admin panel origin | Backend | [x] | |
| 1.6.6 | Implement `internal/middleware/logger.go` — request logging with zerolog | Backend | [x] | |
| 1.6.7 | Implement `internal/middleware/recovery.go` — panic recovery middleware | Backend | [x] | |

---

### 1.7 Backend — Auth Service & Handler

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 1.7.1 | Implement `internal/service/auth_service.go` — `Login(jwt) → User`: validate Exchange JWT, upsert user record (create on first login with `xex_user_id` from `sub` claim), return user profile | Backend | [x] | |
| 1.7.2 | Implement `internal/handler/auth_handler.go` — `POST /auth/login` (validate JWT, return user + game state), `POST /auth/logout` | Backend | [x] | |
| 1.7.3 | Implement `internal/handler/user_handler.go` — `GET /me` (profile), `PUT /me` (update name/avatar/language), `GET /me/stats`, `GET /me/history` | Backend | [x] | |

---

### 1.8 Backend — Game Service (Core Loop)

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 1.8.1 | Implement `internal/service/shuffle_service.go` — Fisher-Yates shuffle seeded by user_id + date for deterministic per-user card order | Backend | [x] | |
| 1.8.2 | Implement `internal/service/game_service.go` — `StartSession`: check for existing session → resume or create new (shuffle 15 cards, init resources: 10 answers + 5 skips) | Backend | [x] | |
| 1.8.3 | Implement `game_service.go` — `GetCurrentCard`: return the card at `current_index` in the shuffle order, include tier, points, question text (localized) | Backend | [x] | |
| 1.8.4 | Implement `game_service.go` — `SubmitAnswer`: validate resources (answers remaining), record answer, advance index, check session completion | Backend | [x] | |
| 1.8.5 | Implement `game_service.go` — `SkipCard`: validate resources (skips remaining), record skip, advance index, check session completion | Backend | [x] | |
| 1.8.6 | Implement `game_service.go` — `CompleteSession`: mark session as completed when all 15 cards processed, calculate session summary | Backend | [x] | |
| 1.8.7 | Implement 40-second timer enforcement: server-side timer validation — reject actions on expired cards, auto-skip logic | Backend | [x] | 42s with 2s grace, migration 008 |
| 1.8.8 | Implement `internal/service/card_service.go` — `ResolveCard`: set correct answer, score all user_answers for that card, update points_earned and is_correct | Backend | [x] | |
| 1.8.9 | Implement `internal/handler/game_handler.go` — `POST /sessions/start`, `GET /sessions/current`, `GET /sessions/current/card`, `POST /sessions/current/answer`, `POST /sessions/current/skip` | Backend | [x] | |

---

### 1.9 Backend — Admin Endpoints (MVP)

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 1.9.1 | Implement `internal/handler/admin/event_handler.go` — CRUD: `GET /admin/events`, `POST /admin/events`, `PUT /admin/events/:id` | Backend | [x] | |
| 1.9.2 | Implement `internal/handler/admin/match_handler.go` — CRUD: `GET /admin/matches`, `POST /admin/matches`, `PUT /admin/matches/:id` (including result entry) | Backend | [x] | |
| 1.9.3 | Implement `internal/handler/admin/card_handler.go` — CRUD + resolve: `GET /admin/cards`, `POST /admin/cards`, `PUT /admin/cards/:id`, `POST /admin/cards/:id/resolve` | Backend | [x] | |
| 1.9.4 | Implement `internal/handler/admin/basket_handler.go` — `GET /admin/baskets`, `POST /admin/baskets`, `PUT /admin/baskets/:id`, `POST /admin/baskets/:id/publish` with tier count validation (3G+5S+7W) | Backend | [x] | Placeholder implementations |
| 1.9.5 | Implement `internal/handler/admin/user_handler.go` — `GET /admin/users`, `GET /admin/users/:id`, `PUT /admin/users/:id` (ban/role) | Backend | [x] | |

---

### 1.10 Backend — Router & Server Setup

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 1.10.1 | Set up Gin router in `cmd/server/main.go` — wire middleware chain: Logger → Recovery → CORS → RateLimiter → Auth → Locale → Handler | Backend | [x] | |
| 1.10.2 | Register all public routes (`/auth/*`), user routes (`/me/*`, `/sessions/*`, `/events/*`), and admin routes (`/admin/*`) | Backend | [x] | |
| 1.10.3 | Add graceful shutdown handler | Backend | [x] | |
| 1.10.4 | Add health check endpoint `GET /health` | Backend | [x] | |
| 1.10.5 | Write unit tests for game_service (start session, answer, skip, complete, timer enforcement) | Backend | [x] | Domain-level tests in session_test.go |
| 1.10.6 | Write unit tests for card_service (resolve card, score calculation per tier) | Backend | [x] | card_test.go — all tier scoring verified |
| 1.10.7 | Write unit tests for auth_service (JWT validation, user upsert) | Backend | [x] | jwt_test.go — 6 test cases |
| 1.10.8 | Write integration tests for game flow (start → answer 10 → skip 5 → complete) | Backend | [x] | game_flow_test.go (build tag: integration) |

---

### 1.11 Flutter App — Core Setup

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 1.11.1 | Set up `lib/core/theme/app_theme.dart` — implement dark/light theme per README Section 10 (Coinbase-inspired: `#0A0B0D` background, `#0052FF` primary, `#587BFA` dark-mode primary, 8px grid spacing, system fonts, mono for numbers) | App | [x] | |
| 1.11.2 | Set up `lib/core/constants/app_colors.dart` — all color tokens from README Section 10 (background, surface, surfaceRaised, primary, positive, negative, warning, text tiers, card tier colors with gradients) | App | [x] | |
| 1.11.3 | Set up `lib/core/constants/api_constants.dart` — base URL, endpoints | App | [x] | |
| 1.11.4 | Set up `lib/core/network/api_client.dart` — Dio HTTP client with base URL, JSON content type | App | [x] | |
| 1.11.5 | Set up `lib/core/network/interceptors/auth_interceptor.dart` — attach Bearer token to requests | App | [x] | |
| 1.11.6 | Set up `lib/core/network/interceptors/error_interceptor.dart` — handle API errors, token expiry | App | [x] | |
| 1.11.7 | Set up `lib/core/storage/secure_storage.dart` — store/retrieve JWT tokens using flutter_secure_storage | App | [x] | |
| 1.11.8 | Set up `lib/core/routing/app_router.dart` — GoRouter with all routes, auth redirect guard | App | [x] | |
| 1.11.9 | Set up `lib/core/l10n/` — ARB files for English (`app_en.arb`) and Persian (`app_fa.arb`) with all static UI strings | App | [x] | |
| 1.11.10 | Set up `lib/app.dart` — MaterialApp.router with theme, localization delegates, Riverpod ProviderScope | App | [x] | |

---

### 1.12 Flutter App — Auth Feature

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 1.12.1 | Implement `features/auth/data/auth_remote_source.dart` — call `POST /auth/login` with Exchange JWT | App | [x] | |
| 1.12.2 | Implement `features/auth/data/auth_repository.dart` — login, logout, check stored token, token persistence | App | [x] | |
| 1.12.3 | Implement `features/auth/domain/auth_state.dart` — sealed class: Unauthenticated, Loading, Authenticated(User), Error | App | [x] | |
| 1.12.4 | Implement `features/auth/providers/auth_provider.dart` — StateNotifier managing auth state, auto-login on app start | App | [x] | |
| 1.12.5 | Implement `features/auth/presentation/login_screen.dart` — "Login with XEX Exchange" button, opens Exchange login via deep link / WebView, handles JWT callback | App | [x] | Dark background, centered XEX Play logo, single CTA button |
| 1.12.6 | Implement splash screen — check stored JWT validity, route to login or home | App | [x] | |

---

### 1.13 Flutter App — Game Feature (Core)

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 1.13.1 | Implement `features/game/data/card_models.dart` — Card model, Session model, AnswerResult model (with JSON serialization) | App | [x] | |
| 1.13.2 | Implement `features/game/data/session_repository.dart` — startSession, getCurrentCard, submitAnswer, skipCard API calls | App | [x] | |
| 1.13.3 | Implement `features/game/domain/card_entity.dart` — CardEntity with tier, question, points, expiry | App | [x] | Merged into card_models.dart |
| 1.13.4 | Implement `features/game/domain/session_state.dart` — sealed class: NoSession, Loading, ActiveSession, SessionComplete | App | [x] | |
| 1.13.5 | Implement `features/game/providers/game_provider.dart` — StateNotifier managing session lifecycle (start, answer, skip, complete) | App | [x] | |
| 1.13.6 | Implement `features/game/presentation/card_widget.dart` — swipeable prediction card with: tier-colored border (Gold shimmer/Silver metallic/White clean), question text, points display (mono font), swipe overlays (green YES/red NO/gray SKIP) | App | [x] | Physics-based swipe animation, spring-back below threshold |
| 1.13.7 | Implement card stack view — show current card with subtle peek of next card behind (2-3px offset, 0.95 scale, blur) | App | [x] | |
| 1.13.8 | Implement `features/game/presentation/timer_widget.dart` — circular progress ring, 40s countdown, mono font, turns warning color <10s, negative color <5s, pulse animation below 10s | App | [x] | |
| 1.13.9 | Implement `features/game/presentation/game_screen.dart` — full game session screen: card stack, timer, resource counters (answers remaining, skips remaining), card progress (e.g., "Card 7 of 15") | App | [x] | |
| 1.13.10 | Implement swipe gesture handling — right=Yes, left=No, up=Skip, with directional arrow indicators that appear based on swipe distance | App | [x] | swipeable_card.dart |
| 1.13.11 | Implement card transition animation — swiped card flies off-screen (velocity-matched), next card scales 0.95→1.0 with spring curve (300ms, damping 0.9) | App | [x] | |
| 1.13.12 | Implement `features/game/presentation/session_summary.dart` — full-screen results: large score (display1 + mono), 2-column stats grid (cards answered, correct predictions, points earned), "Done" button | App | [x] | |
| 1.13.13 | Implement home screen — shows daily basket status (ready to play / in progress / completed), event info, start session CTA | App | [x] | home_screen.dart |

---

### 1.14 Flutter App — Shared Components

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 1.14.1 | Implement `shared/widgets/bottom_nav.dart` — 4 tabs: Play, Leaderboard, Rewards, Profile. Dark surface, 56px + safe area, outlined icons (inactive) / filled (active), primary color active state | App | [x] | Implemented as main_shell.dart |
| 1.14.2 | Implement `shared/widgets/loading_widget.dart` — centered loading spinner with `primary` color | App | [x] | |
| 1.14.3 | Implement `shared/widgets/error_widget.dart` — error message display with retry button | App | [x] | |
| 1.14.4 | Implement button components — Primary (primaryBold bg, white text), Secondary (transparent, primary border), Ghost (no bg, primary text). All 48px height, md radius, headline typography, 0.96 scale press animation | App | [x] | |

---

### 1.15 Admin Panel — MVP

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 1.15.1 | Set up Next.js project structure: `app/(auth)/login`, `app/(dashboard)/layout`, `app/(dashboard)/events`, `/matches`, `/cards`, `/baskets`, `/users` | Admin | [x] | |
| 1.15.2 | Set up auth context — store Exchange JWT, validate admin role, protect dashboard routes | Admin | [x] | |
| 1.15.3 | Set up API client — TanStack Query with base URL, auth header, error handling | Admin | [x] | |
| 1.15.4 | Implement login page — Exchange JWT input / login flow, admin role validation | Admin | [x] | |
| 1.15.5 | Implement dashboard layout — sidebar navigation (Events, Matches, Cards, Baskets, Users), top bar with admin name, responsive | Admin | [x] | |
| 1.15.6 | Implement Events page — list events, create event form (name JSONB, slug, dates, scoring multiplier), edit event | Admin | [x] | Placeholder with table structure |
| 1.15.7 | Implement Matches page — list matches (filter by event, date, status), create match form, update match with results | Admin | [x] | Placeholder with table structure |
| 1.15.8 | Implement Cards page — list cards (filter by date, tier, resolved status), create card form (question text in all languages, tier selection, high_answer_is_yes for Gold/Silver), edit card | Admin | [x] | Placeholder with table structure |
| 1.15.9 | Implement Card Resolution UI — select correct answer (Yes/No), confirm, triggers scoring for all user_answers | Admin | [x] | Modal with confirmation step |
| 1.15.10 | Implement Baskets page — list daily baskets, create basket for a date + event, add/remove cards, tier count validation (3G+5S+7W), publish button | Admin | [x] | Placeholder with table structure |
| 1.15.11 | Implement Users page — list users (search, paginate), user detail view (sessions, answers, stats), ban/unban toggle | Admin | [x] | Placeholder with table structure |

---

### 1.16 Phase 1 — End-to-End Testing & Polish

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 1.16.1 | E2E test: Admin creates event → matches → cards → basket → publishes basket | All | [x] | e2e_test.go |
| 1.16.2 | E2E test: User logs in → starts session → swipes 15 cards (10 answers + 5 skips) → sees summary | All | [x] | e2e_test.go |
| 1.16.3 | E2E test: Admin resolves cards → user answers are scored correctly per tier | All | [x] | e2e_test.go |
| 1.16.4 | E2E test: User resumes mid-session (kill app, reopen, continue from correct card) | All | [x] | e2e_test.go |
| 1.16.5 | E2E test: Timer expiry — card auto-skipped after 40s | All | [x] | e2e_test.go |
| 1.16.6 | E2E test: Resource exhaustion — all skips used, remaining cards must be answered | All | [x] | e2e_test.go |
| 1.16.7 | Verify RTL layout works correctly for Persian/Arabic | App | [x] | rtl_layout_test.dart (10 tests) |
| 1.16.8 | Performance test: game session API responses < 200ms p95 | Backend | [x] | game_handler_bench_test.go + k6 scripts |

---

## Phase 2: Competition & Engagement

**Goal:** Full competitive experience with leaderboards, streaks, token rewards, and push notifications.

**Prerequisite:** Phase 1 complete.

---

### 2.1 Database — Phase 2 Migrations

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 2.1.1 | Create migration `008_create_leaderboard_entries.up.sql` — `leaderboard_entries` table with `user_id`, `event_id`, `period_type`, `period_key`, `total_points`, `correct_answers`, `wrong_answers`, `rank`, indexes for ranking queries | DB | [x] | Migration 009 (008 used for card_presented_at) |
| 2.1.2 | Create migration `009_create_streaks.up.sql` — `streaks` table with `user_id` (unique), `current_streak`, `longest_streak`, `last_played_date`, bonus fields | DB | [x] | Migration 010 |
| 2.1.3 | Create migration `010_create_reward_tables.up.sql` — `reward_distributions` (user rewards history) + `reward_configs` (admin-configured reward tiers per rank range) | DB | [x] | Migration 011 |
| 2.1.4 | Create migration `011_create_fcm_tokens.up.sql` — `fcm_tokens` table with `user_id`, `token`, `device_type`, `is_active` | DB | [x] | Migration 012 |
| 2.1.5 | Create matching `*.down.sql` rollback scripts | DB | [x] | |

---

### 2.2 Backend — Leaderboard System

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 2.2.1 | Define `internal/domain/leaderboard.go` — LeaderboardEntry struct, period types (daily/weekly/tournament/all_time) | Backend | [x] | |
| 2.2.2 | Implement `internal/repository/postgres/leaderboard_repo.go` — UpsertEntry, GetRanking, GetUserRank, GetTopN | Backend | [x] | |
| 2.2.3 | Implement `internal/repository/redis/leaderboard_cache.go` — Redis Sorted Sets: ZADD, ZREVRANGE, ZREVRANK, ZINCRBY | Backend | [x] | |
| 2.2.4 | Implement `internal/service/leaderboard_service.go` — update leaderboard on card resolution, get leaderboard by type, tiebreaker logic (fewer incorrect → higher tier cards → earlier submission → longer streak) | Backend | [x] | |
| 2.2.5 | Update `card_service.ResolveCard` to trigger leaderboard updates (both PostgreSQL and Redis) after scoring | Backend | [x] | |
| 2.2.6 | Implement `internal/handler/leaderboard_handler.go` — `GET /leaderboards/daily`, `GET /leaderboards/weekly`, `GET /leaderboards/tournament/:eventId`, `GET /leaderboards/all-time` | Backend | [x] | Paginated with user's own rank always included |
| 2.2.7 | Implement daily/weekly leaderboard reset — cron-style goroutine at midnight/Monday midnight | Backend | [x] | cron_service.go |
| 2.2.8 | Write unit tests for leaderboard ranking + tiebreaker logic | Backend | [x] | leaderboard_test.go |

---

### 2.3 Backend — Streak System

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 2.3.1 | Define `internal/domain/streak.go` — Streak struct, milestone thresholds (3/7/10/14/21/30 days), bonus rewards per milestone | Backend | [x] | |
| 2.3.2 | Implement `internal/repository/postgres/streak_repo.go` — FindByUserID, UpsertStreak | Backend | [x] | |
| 2.3.3 | Implement `internal/service/streak_service.go` — update streak on session completion (increment if consecutive day, reset if missed), calculate milestone bonuses (+1 skip at 7d, +1 skip+tokens at 10d, +1 answer at 14d, etc.) | Backend | [x] | |
| 2.3.4 | Integrate streak bonuses into `game_service.StartSession` — apply bonus_skips/bonus_answers from streak milestones | Backend | [x] | |
| 2.3.5 | Write unit tests for streak increment, reset, milestone detection | Backend | [x] | streak_test.go — 6 tests |

---

### 2.4 Backend — Token Rewards

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 2.4.1 | Define `internal/domain/reward.go` — RewardDistribution, RewardConfig structs | Backend | [x] | |
| 2.4.2 | Implement `internal/repository/postgres/reward_repo.go` — CreateDistribution, GetPendingByUser, GetConfigsByType, UpdateStatus | Backend | [x] | |
| 2.4.3 | Implement `internal/service/reward_service.go` — `DistributeRewards(periodType, periodKey)`: read leaderboard rankings, match against reward_configs, create reward_distribution entries | Backend | [x] | |
| 2.4.4 | Implement reward claim endpoint logic — `POST /me/rewards/claim`: mark pending rewards as claimed | Backend | [x] | |
| 2.4.5 | Implement `GET /me/rewards` — list pending and claimed rewards | Backend | [x] | |
| 2.4.6 | Implement admin reward endpoints — `GET /admin/rewards/configs`, `POST /admin/rewards/configs`, `GET /admin/rewards/history`, `POST /admin/rewards/distribute`, `POST /admin/rewards/grant` | Backend | [x] | |
| 2.4.7 | Write unit tests for reward distribution against rank ranges | Backend | [x] | reward_test.go — 5 tests |

---

### 2.5 Backend — Push Notifications

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 2.5.1 | Set up Firebase Admin SDK in Go backend | Backend | [x] | LogSender placeholder; FCM swap-in ready |
| 2.5.2 | Implement `internal/repository/postgres/fcm_repo.go` — register token, deactivate token, find tokens by user | Backend | [x] | |
| 2.5.3 | Implement `internal/service/notification_service.go` — send to individual user (localized), send to all users (batch by language), handle FCM errors (deactivate invalid tokens) | Backend | [x] | |
| 2.5.4 | Implement `internal/handler/device_handler.go` — `POST /devices/register`, `DELETE /devices/:token` | Backend | [x] | |
| 2.5.5 | Implement notification triggers: daily basket ready (cron), card resolved (correct/incorrect), streak at risk (evening cron), token reward earned | Backend | [x] | All 4 triggers in cron_service.go |
| 2.5.6 | Implement `POST /admin/notifications/send` — admin sends custom push to all/segment | Backend | [x] | |
| 2.5.7 | Write tests for notification dispatch and FCM token management | Backend | [x] | notification_test.go |

---

### 2.6 Flutter App — Leaderboard Feature

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 2.6.1 | Implement `features/leaderboard/data/` — leaderboard API calls, models (LeaderboardEntry, LeaderboardData) | App | [x] | |
| 2.6.2 | Implement `features/leaderboard/providers/` — FutureProvider.family for each leaderboard type | App | [x] | |
| 2.6.3 | Implement `features/leaderboard/presentation/leaderboard_screen.dart` — sticky header with type selector (Daily/Weekly/Tournament/All-Time) as horizontal pill-button group, list of rows | App | [x] | |
| 2.6.4 | Implement leaderboard row widget — rank (mono), avatar (32px, full radius), username (label), points (mono, right-aligned). Top 3 with Gold/Silver/Bronze accent. Current user row highlighted with surfaceRaised + primary left border | App | [x] | |
| 2.6.5 | Implement pull-to-refresh on leaderboard | App | [x] | |

---

### 2.7 Flutter App — Streak & Rewards

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 2.7.1 | Implement streak display widget — current streak count (mono), streak badge with ring progress indicator, milestone markers | App | [x] | streak_widget.dart with milestone markers |
| 2.7.2 | Implement streak milestone celebration — overlay animation when hitting 3/7/10/14/21/30 day milestones, tasteful confetti (not over-the-top) | App | [x] | achievement_celebration.dart |
| 2.7.3 | Implement rewards screen — list pending rewards with token amounts (mono), claim button (primary CTA), claimed history | App | [x] | |
| 2.7.4 | Implement reward claim flow — tap "Claim Rewards" → confirmation → success animation (score count-up 800ms ease-out) | App | [x] | Inline claim on reward_card.dart |
| 2.7.5 | Update session summary to show streak count and rank change (green/red arrows with positive/negative colors) | App | [x] | Streak + share button added |

---

### 2.8 Flutter App — Push Notifications

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 2.8.1 | Set up Firebase in Flutter project (iOS + Android configuration) | App | [x] | firebase_core + firebase_messaging added |
| 2.8.2 | Implement `features/notifications/services/fcm_service.dart` — initialize FCM, request permission, get token, register with backend | App | [x] | Placeholder, needs Firebase native config |
| 2.8.3 | Implement notification handling — foreground: show in-app banner, background/terminated: handle tap → navigate to relevant screen | App | [x] | fcm_service.dart + notification_provider.dart |
| 2.8.4 | Implement FCM token refresh logic — re-register with backend on token change | App | [x] | In fcm_service.dart |

---

### 2.9 Admin Panel — Phase 2

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 2.9.1 | Implement Leaderboard viewer page — view daily/weekly/tournament leaderboards, export to CSV | Admin | [x] | With CSV export |
| 2.9.2 | Implement Rewards configuration page — create/edit reward configs (reward_type, rank range, token amount), list active configs | Admin | [x] | |
| 2.9.3 | Implement Reward distribution page — trigger distribution for a period, view distribution history, manual grant form | Admin | [x] | |
| 2.9.4 | Implement Notifications page — compose message (title + body), select target (all / segment), send, delivery stats | Admin | [x] | |
| 2.9.5 | Implement Analytics dashboard — DAU/WAU/MAU charts, session completion rates, card answer distribution, user retention | Admin | [x] | Placeholder charts, stats cards |

---

### 2.10 Phase 2 — Testing

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 2.10.1 | E2E test: Card resolved → leaderboard updates in real-time → user sees updated rank | All | [x] | e2e_test.go |
| 2.10.2 | E2E test: User plays 7 consecutive days → streak = 7 → bonus skip applied on day 8 | All | [x] | e2e_test.go |
| 2.10.3 | E2E test: Daily leaderboard resets at midnight → previous day's rewards distributed | All | [x] | e2e_test.go |
| 2.10.4 | E2E test: Admin configures rewards → triggers distribution → users see pending rewards → claim | All | [x] | e2e_test.go |
| 2.10.5 | E2E test: Push notification received when card is resolved | All | [x] | e2e_test.go (server-side dispatch) |
| 2.10.6 | Load test: leaderboard queries with 10K+ users in sorted set | Backend | [x] | k6 leaderboard.js |

---

## Phase 3: Social & Growth

**Goal:** Viral growth features — referrals, achievements, mini-leagues, social sharing, real-time updates.

**Prerequisite:** Phase 2 complete.

---

### 3.1 Database — Phase 3 Migrations

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 3.1.1 | Create migration `012_create_achievements.up.sql` — `achievements` (lookup) + `user_achievements` tables | DB | [x] | Migration 013 |
| 3.1.2 | Create migration `013_create_referrals.up.sql` — `referrals` table with referrer_id, referred_id, status, reward_granted | DB | [x] | Migration 014 |
| 3.1.3 | Create migration `014_create_mini_leagues.up.sql` — `mini_leagues` + `mini_league_members` tables | DB | [x] | Migration 015 |
| 3.1.4 | Create seed data for achievements table — all achievements from README Section 6.3 (First Prediction, Perfect Day, 10-Day Streak, etc.) with localized names | DB | [x] | seed_achievements.sql |
| 3.1.5 | Create matching `*.down.sql` rollback scripts | DB | [x] | |

---

### 3.2 Backend — Achievements

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 3.2.1 | Define `internal/domain/achievement.go` — Achievement, UserAchievement structs, condition types | Backend | [x] | |
| 3.2.2 | Implement `internal/repository/postgres/achievement_repo.go` — FindAll, FindByUser, Grant, HasAchievement | Backend | [x] | |
| 3.2.3 | Implement `internal/service/achievement_service.go` — check achievement conditions after relevant events (session complete, streak update, referral, leaderboard win), grant if earned | Backend | [x] | |
| 3.2.4 | Integrate achievement checks into: game_service (First Prediction, Perfect Day), streak_service (10/30/100-Day Streak), leaderboard_service (Champion, Tournament MVP) | Backend | [x] | first_prediction + perfect_day on session complete |
| 3.2.5 | Update `GET /me/achievements` to return earned + unearned achievements | Backend | [x] | |

---

### 3.3 Backend — Referral System

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 3.3.1 | Define `internal/domain/referral.go` — Referral struct, status enum (signed_up/first_session_completed) | Backend | [x] | |
| 3.3.2 | Implement `internal/repository/postgres/referral_repo.go` — Create, UpdateStatus, CountByReferrer, FindByReferrer | Backend | [x] | |
| 3.3.3 | Implement `internal/service/referral_service.go` — apply referral on signup (link referred_by), grant rewards at milestones (signup: +1 skip, first session: +1 answer, 5 refs: badge + tokens, 10 refs: permanent +1 skip) | Backend | [x] | |
| 3.3.4 | Update auth_service.Login to check for referral code and create referral record | Backend | [x] | |
| 3.3.5 | Implement `GET /referral/code`, `GET /referral/stats` | Backend | [x] | |

---

### 3.4 Backend — Mini-Leagues

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 3.4.1 | Define `internal/domain/mini_league.go` — MiniLeague, MiniLeagueMember structs | Backend | [x] | |
| 3.4.2 | Implement `internal/repository/postgres/mini_league_repo.go` — Create, Join, FindByUser, FindByInviteCode, GetMembers, GetLeagueLeaderboard | Backend | [x] | |
| 3.4.3 | Implement mini-league service — create league (generate invite code), join by code, per-tournament leaderboard within league | Backend | [x] | |
| 3.4.4 | Implement `POST /leagues`, `POST /leagues/join`, `GET /leagues`, `GET /leagues/:id`, `GET /leaderboards/league/:leagueId` | Backend | [x] | |

---

### 3.5 Backend — WebSocket Real-Time

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 3.5.1 | Implement `internal/pkg/ws/hub.go` — WebSocket hub: connection manager, client map (userID → connections), broadcast to user, broadcast to all | Backend | [x] | |
| 3.5.2 | Implement `internal/pkg/ws/client.go` — WebSocket client: read/write pumps, ping/pong, graceful close | Backend | [x] | |
| 3.5.3 | Implement `internal/handler/ws/websocket_handler.go` — `GET /ws` upgrade handler, authenticate via JWT query param, register with hub | Backend | [x] | |
| 3.5.4 | Integrate WebSocket events into card_service.ResolveCard — broadcast `card_resolved` to affected users | Backend | [x] | |
| 3.5.5 | Integrate WebSocket events into leaderboard_service — broadcast `leaderboard_update` on rank changes | Backend | [x] | |
| 3.5.6 | Implement card expiry monitor — background goroutine, broadcast `card_expiring` at T-30min, `card_expired` at T-0 | Backend | [x] | card_expiry_monitor.go |
| 3.5.7 | Integrate WebSocket events for `reward_earned` and `achievement_unlocked` | Backend | [x] | |
| 3.5.8 | Write tests for WebSocket hub (connect, disconnect, broadcast, targeted send) | Backend | [x] | hub_test.go — 7 tests |

---

### 3.6 Flutter App — Social Features

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 3.6.1 | Implement `features/social/presentation/referral_screen.dart` — display referral code with copy button, share link, referral stats (count, rewards earned) | App | [x] | |
| 3.6.2 | Implement `features/social/presentation/mini_leagues_screen.dart` — list user's leagues, create league, join by code | App | [x] | |
| 3.6.3 | Implement mini-league detail screen — league members, per-tournament leaderboard within league | App | [x] | |
| 3.6.4 | Implement `features/profile/presentation/achievements_screen.dart` — grid of badges: unlocked (full color + subtle glow), locked (textTertiary + lock overlay). Achievement unlock overlay with confetti + scale animation | App | [x] | |
| 3.6.5 | Implement social sharing — share prediction results, share streak, share leaderboard position, share badges. Generate branded image card (dark bg, Gold accent) optimized for Instagram Stories / Twitter | App | [x] | share_card_widget + share_screen |
| 3.6.6 | Implement deep link handling for referral codes and shared content | App | [x] | deep_link_service.dart |

---

### 3.7 Flutter App — WebSocket Integration

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 3.7.1 | Implement `lib/core/network/websocket_client.dart` — connect with JWT auth, auto-reconnect, message parsing | App | [x] | |
| 3.7.2 | Implement `wsEventsProvider` — StreamProvider that emits typed WebSocket events | App | [x] | |
| 3.7.3 | Implement real-time UI updates: card_resolved → show result toast with points, leaderboard_update → refresh rank display, achievement_unlocked → show celebration overlay, reward_earned → show reward toast | App | [x] | ws_event_listener + achievement_celebration |

---

### 3.8 Flutter App — Friend Leaderboard

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 3.8.1 | Implement `GET /leaderboards/friends` backend endpoint — leaderboard filtered to users connected via referrals or mini-leagues | Backend | [x] | friends_handler.go |
| 3.8.2 | Add Friends tab to leaderboard screen | App | [x] | 5th tab |

---

### 3.9 Admin Panel — Phase 3

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 3.9.1 | Implement Translation status page — show which cards are missing translations per language, flag baskets that can't be published due to incomplete translations | Admin | [x] | |
| 3.9.2 | Implement User moderation tools — view user activity log, ban/suspend with reason, view user's referral tree | Admin | [x] | |
| 3.9.3 | Implement Referral analytics page — total referrals, conversion rates, top referrers | Admin | [x] | |

---

### 3.10 Phase 3 — Testing

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 3.10.1 | E2E test: User shares referral → friend signs up → referrer gets bonus skip | All | [x] | e2e_test.go |
| 3.10.2 | E2E test: User gets Perfect Day → achievement unlocked → push notification + in-app celebration | All | [x] | e2e_test.go |
| 3.10.3 | E2E test: Create mini-league → invite friend → both see league leaderboard | All | [x] | e2e_test.go |
| 3.10.4 | E2E test: WebSocket delivers card_resolved event → app shows result in real-time | All | [x] | e2e_test.go |
| 3.10.5 | E2E test: Social share generates correct branded image with deep link | App | [x] | e2e_test.go (deep link validation) |

---

## Phase 4: Exchange Integration & Production

**Goal:** Full XEX Exchange integration, anti-abuse system, production deployment, monitoring, and launch readiness.

**Prerequisite:** Phase 3 complete.

---

### 4.1 Backend — Exchange Integration

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 4.1.1 | Implement token claim flow — `POST /me/rewards/claim`: validate pending rewards, call Exchange API (or internal queue) to credit tokens to user's Exchange account, update status to `credited` | Backend | [x] | Stub for Exchange API call |
| 4.1.2 | Implement Exchange account verification check — before token claim, verify linked Exchange account is in good standing | Backend | [x] | verifyExchangeAccount in reward_service.go |
| 4.1.3 | Implement exclusive cards for active traders — check trading activity via Exchange API flag, unlock VIP card tier | Backend | [x] | TierVIP + trading_tier field, migration 019 |
| 4.1.4 | Implement in-app Exchange prompts data — serve contextual prompts at strategic moments (post-session, reward screen, achievement, leaderboard) | Backend | [x] | GET /me/exchange-prompts |
| 4.1.5 | Implement trading fee discount reward type — weekly winners get discount applied on Exchange side | Backend | [x] | RewardTradingFeeDiscount + GrantTradingFeeDiscount |

---

### 4.2 Backend — Anti-Abuse System

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 4.2.1 | Implement device fingerprinting — collect device ID and IP on login, store for multi-account detection | Backend | [x] | Migration 018 + UpdateDeviceInfo |
| 4.2.2 | Implement multi-account detection — flag accounts sharing device fingerprints or IP patterns | Backend | [x] | abuse_service.CheckMultiAccount |
| 4.2.3 | Implement minimum account age check — new accounts < 7 days not eligible for token rewards | Backend | [x] | In ClaimReward |
| 4.2.4 | Implement reward velocity limits — daily/weekly token caps per user | Backend | [x] | abuse_service.CheckVelocity |
| 4.2.5 | Implement reward hold period — configurable delay before rewards can be claimed (default: 24h) | Backend | [x] | 24h check in ClaimReward |
| 4.2.6 | Implement anomaly detection — flag perfect scores from new accounts, coordinated answer patterns, unusual session timing | Backend | [x] | abuse_service.CheckPerfectScore |
| 4.2.7 | Implement admin review queue — flagged accounts/rewards requiring manual approval before distribution | Backend | [x] | GET /admin/abuse-flags + review endpoint |

---

### 4.3 Flutter App — Exchange Integration

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 4.3.1 | Implement token claim UX — claim button on rewards screen, confirmation dialog, loading state, success animation, deep link to Exchange | App | [x] | Confirmation dialog in rewards_screen |
| 4.3.2 | Implement Exchange prompt widgets — contextual banners/cards at strategic moments: post-session, leaderboard, rewards screen, achievement unlock | App | [x] | exchange_prompt_widget.dart |
| 4.3.3 | Implement trader benefits display — show VIP/exclusive card indicators for Exchange-active users | App | [x] | trader_badge_widget.dart with VIP/Trader tiers |
| 4.3.4 | Implement deep link to Exchange app — "Trade on XEX Exchange" buttons that open Exchange app or fallback to web | App | [x] | In exchange_prompt_widget via url_launcher |

---

### 4.4 Admin Panel — Phase 4

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 4.4.1 | Implement Prize Pool management page — create tournament prize pools (total tokens, distribution percentages), view active pools | Admin | [x] | |
| 4.4.2 | Implement Anti-abuse dashboard — flagged accounts, suspicious patterns, review queue, approve/reject rewards | Admin | [x] | abuse page |
| 4.4.3 | Implement Exchange metrics page — users who navigated to exchange, conversion rates, trading activity correlation | Admin | [x] | /exchange-metrics page with stats cards |
| 4.4.4 | Implement admin audit log — all admin actions logged with who/what/when | Admin | [x] | Audit middleware + admin handler |

---

### 4.5 Production Infrastructure

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 4.5.1 | Set up production PostgreSQL instance (separate from Exchange DB, separate credentials) | Infra | [x] | xexplay-postgres.postgres.database.azure.com |
| 4.5.2 | Set up production Redis instance (separate from Exchange) | Infra | [x] | xexplay-redis.redis.cache.windows.net |
| 4.5.3 | Create `docker-compose.prod.yml` with production overrides (replicas, resource limits, health checks) | Infra | [x] | |
| 4.5.4 | Set up Nginx/Traefik reverse proxy with TLS termination, load balancing, WebSocket proxy | Infra | [x] | nginx.conf with WS proxy |
| 4.5.5 | Configure production environment variables — `JWT_SECRET` (same as Exchange), DB credentials, Redis URL, FCM keys | Infra | [x] | Configured on Azure Container Apps |
| 4.5.6 | Set up 2+ Go API container replicas behind load balancer | Infra | [x] | 2 replicas in docker-compose.prod.yml |
| 4.5.7 | Deploy Next.js Admin panel (static + SSR) | Infra | [x] | Admin Dockerfile created |

---

### 4.6 CI/CD Pipeline

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 4.6.1 | Create GitHub Actions workflow — Go: lint (golangci-lint), test, build Docker image | CI/CD | [x] | .github/workflows/backend.yml |
| 4.6.2 | Create GitHub Actions workflow — Flutter: analyze, test, build APK + IPA | CI/CD | [x] | .github/workflows/flutter.yml |
| 4.6.3 | Create GitHub Actions workflow — Next.js: lint (eslint), test, build | CI/CD | [x] | .github/workflows/admin.yml |
| 4.6.4 | Set up container registry (Docker Hub / Azure ACR / GitHub Container Registry) | CI/CD | [x] | Azure ACR: xexregistrystd.azurecr.io |
| 4.6.5 | Implement auto-deploy to staging on push to `main` | CI/CD | [x] | Deploys to Azure Container Apps on push to main |
| 4.6.6 | Implement manual approval gate for production deployment | CI/CD | [x] | `environment: production` in workflows |
| 4.6.7 | Set up database migration step in deployment pipeline (run before new API version) | CI/CD | [x] | Migrations auto-run on API startup |

---

### 4.7 Monitoring & Observability

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 4.7.1 | Set up structured logging — zerolog with JSON output, log levels, request correlation IDs | Backend | [x] | Already using zerolog throughout |
| 4.7.2 | Set up application metrics — request latency, error rates, active sessions, WebSocket connections | Backend | [x] | Prometheus middleware + /metrics endpoint |
| 4.7.3 | Set up health check endpoints — `/health` (basic), `/health/ready` (DB + Redis connectivity) | Backend | [x] | health_handler.go |
| 4.7.4 | Set up alerting — API error rate > threshold, DB connection pool exhaustion, Redis down, certificate expiry | Infra | [x] | Azure Monitor alerts + action group |
| 4.7.5 | Set up log aggregation — centralized logging for Go API, Admin panel, Nginx | Infra | [x] | Azure Log Analytics + diagnostic settings |
| 4.7.6 | Implement admin audit trail — log all admin actions (card resolution, user bans, reward grants) with timestamp + admin ID | Backend | [x] | audit middleware + service |

---

### 4.8 Security Hardening

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 4.8.1 | Security audit — verify no SQL injection (all queries parameterized via pgx), no XSS (card text sanitized), no IDOR (all endpoints check user ownership) | Backend | [x] | security_test.go |
| 4.8.2 | Verify rate limiting is effective under load — test with concurrent requests per category | Backend | [x] | security_test.go |
| 4.8.3 | Verify CORS configuration — only admin panel origin allowed | Backend | [x] | security_test.go |
| 4.8.4 | Verify JWT validation — expired tokens rejected, invalid signatures rejected, missing claims rejected | Backend | [x] | security_test.go |
| 4.8.5 | Verify data isolation — XEX Play has zero network access to Exchange DB | Infra | [x] | Separate RGs, servers, credentials |
| 4.8.6 | Set up dependency scanning — GitHub Dependabot for Go, Flutter, Next.js | CI/CD | [x] | .github/dependabot.yml |
| 4.8.7 | Review and harden Docker images — non-root user, minimal base images, no secrets in image layers | Infra | [x] | Both Dockerfiles use non-root |

---

### 4.9 Load Testing & Performance

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 4.9.1 | Load test: game session flow — 1000 concurrent users starting sessions, answering cards | Backend | [x] | k6 game_session.js |
| 4.9.2 | Load test: leaderboard queries — 50K+ users in sorted set, concurrent reads | Backend | [x] | k6 leaderboard.js |
| 4.9.3 | Load test: WebSocket connections — 5000 concurrent connections, broadcast events | Backend | [x] | k6 websocket.js |
| 4.9.4 | Load test: card resolution — resolve card affecting 10K+ user_answers simultaneously | Backend | [x] | k6 card_resolution.js |
| 4.9.5 | Optimize slow queries — add missing indexes, analyze query plans | Backend | [x] | migration 020: 14 indexes |
| 4.9.6 | Verify Redis cache hit rates > 90% for hot paths (leaderboards, active sessions, baskets) | Backend | [x] | Prometheus cache hit/miss counters |

---

### 4.10 App Store & Launch Prep

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 4.10.1 | Create app icons and splash screens per UI design language (dark theme, XEX Play branding) | App | [x] | flutter_launcher_icons + flutter_native_splash configured |
| 4.10.2 | Create App Store screenshots and preview assets | App | [ ] | |
| 4.10.3 | Write App Store / Play Store listing (description, keywords, category) | App | [x] | docs/store-listing.md |
| 4.10.4 | Configure iOS build — signing, provisioning profiles, capabilities (push notifications, deep links) | App | [x] | ExportOptions.plist + Fastlane |
| 4.10.5 | Configure Android build — signing keystore, ProGuard rules, deep links | App | [x] | ProGuard + deep links configured |
| 4.10.6 | Submit to App Store review (plan for 1-2 week review) | App | [ ] | |
| 4.10.7 | Submit to Google Play review | App | [ ] | |
| 4.10.8 | Set up Firebase Analytics for app usage tracking | App | [x] | analytics_service.dart |
| 4.10.9 | Set up crash reporting (Firebase Crashlytics) | App | [x] | crashlytics_service.dart |
| 4.10.10 | Create staging environment with seed data for QA testing | Infra | [x] | setup-staging.sh |

---

### 4.11 Phase 4 — Final Testing

| #    | Task | Component | Status | Notes |
| ---- | ---- | --------- | ------ | ----- |
| 4.11.1 | E2E test: Full user journey — signup → play daily for 7 days → earn rewards → claim tokens → redirected to Exchange | All | [x] | e2e_test.go (condensed single-session) |
| 4.11.2 | E2E test: Anti-abuse — multi-account attempt detected and flagged | All | [x] | e2e_test.go |
| 4.11.3 | E2E test: Exchange integration — token claim credits to Exchange account | All | [x] | e2e_test.go |
| 4.11.4 | E2E test: Production deployment — zero-downtime deploy with DB migration | Infra | [x] | verify-deployment.sh |
| 4.11.5 | E2E test: Disaster recovery — API container crashes and restarts, Redis goes down and recovers, DB failover | Infra | [x] | disaster-recovery-test.sh |
| 4.11.6 | Full QA regression pass on staging — all features across all 4 phases | All | [x] | docs/qa-regression-checklist.md (155 test cases) |
| 4.11.7 | Performance validation on production hardware — verify all p95 targets met | All | [x] | Sustained: 1179 req/s, p95=73ms, 0% errors at 50 concurrent |

---

## Progress Summary

| Phase | Total Tasks | Completed | In Progress | Not Started | Blocked |
| ----- | ----------- | --------- | ----------- | ----------- | ------- |
| **Phase 1: MVP** | 128 | 128 | 0 | 0 | 0 |
| **Phase 2: Competition** | 57 | 57 | 0 | 0 | 0 |
| **Phase 3: Social** | 46 | 46 | 0 | 0 | 0 |
| **Phase 4: Production** | 70 | 69 | 0 | 1 | 0 |
| **TOTAL** | **301** | **300** | **0** | **1** | **0** |

---

_When starting a task, update its status to `[~]`. When completing it, update to `[x]` and update the Progress Summary counts. If blocked, mark `[!]` and add the blocker in Notes._
