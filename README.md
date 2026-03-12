# XEX Play

> Card-based sports prediction game, pick your battles, climb the leaderboard, earn rewards on XEX Exchange.

| Document                             | Description                                                              |
| ------------------------------------ | ------------------------------------------------------------------------ |
| **This file**                        | Product vision, game design, and mechanics                               |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | System architecture, tech stack, database schema, API design, deployment |

---

## Table of Contents

1. [Overview](#1-overview)
2. [Target Audience & Positioning](#2-target-audience--positioning)
3. [Core Game Mechanics](#3-core-game-mechanics)
4. [User Journey](#4-user-journey)
5. [Scoring & Ranking System](#5-scoring--ranking-system)
6. [Engagement & Growth Features](#6-engagement--growth-features)
7. [XEX Exchange Integration](#7-xex-exchange-integration)
8. [Admin Capabilities](#8-admin-capabilities) (includes 8.10 Sports Automation, 8.11 Settings)
9. [Monetization Strategy](#9-monetization-strategy)
10. [UI/UX Design Language](#10-uiux-design-language)

---

## 1. Overview

### What is XEX Play?

XEX Play is a card-based sports prediction game designed to engage users during live sports tournaments and create a daily habit loop. Each day, users receive a curated basket of 15 prediction cards tied to real upcoming matches. Users must answer 10 and skip 5 — they see every card but can't go back, creating genuine risk/reward decisions rather than mindless guessing. With real token rewards redeemable on XEX Exchange, every session has stakes.

### Elevator Pitch

> "Pick your battles. 15 cards, 10 answers, 5 skips — predict sports outcomes, climb the leaderboard, and earn real token rewards on XEX Exchange."

### Value Proposition

| For Users                          | For XEX Exchange                           |
| ---------------------------------- | ------------------------------------------ |
| Free, engaging daily sports game   | User acquisition funnel                    |
| Strategic gameplay, not just luck  | Daily active user engagement               |
| Compete with friends and globally  | Conversion to exchange trading             |
| Earn real token rewards             | Brand association with major sports events |
| Longer sessions (~10 minutes/day)  | Cross-promotion channel                    |

### Design Principles

- **Daily Habit Formation** — The game is designed to bring users back every single day through limited daily resources, streaks, and expiring cards.
- **Real Decision-Making** — 10 answers and 5 skips across 15 cards force users to evaluate risk/reward on every single card. Every card is seen, no going back.
- **Skip Management as Strategy** — Users see all 15 cards but must skip 5. Choosing when to skip is the core strategic decision — skip now hoping for better, or answer a mediocre card to save skips?
- **Live Sports Connection** — Cards are directly tied to real matches with real expiry times, keeping the game grounded in reality.
- **Multi-Language First** — The app, card content, and notifications are fully localized. Users play in their preferred language.
- **Exchange Funnel** — Every feature subtly guides users toward XEX Exchange without being intrusive.

---

## 2. Target Audience & Positioning

### Primary Audience

- **Sports enthusiasts** (18–40) who follow major tournaments (World Cup, Champions League, Euro, domestic leagues)
- **Casual gamers** who enjoy quick daily mobile games with competitive elements
- **Existing XEX Exchange users** looking for additional ways to engage with the platform
- **Crypto-curious sports fans** who can be introduced to XEX Exchange through the game

### Positioning

XEX Play sits at the intersection of sports prediction and strategic card games. Unlike traditional betting apps, XEX Play:

- Is **free to play** — no money required to participate
- Emphasizes **strategy over luck** — resource management is key
- Focuses on **engagement over gambling** — the goal is fun and competition, not wagering
- Acts as a **gateway to XEX Exchange** — rewards are exchange-based, not cash payouts

### Supported Events (Multi-Sport Architecture)

The platform supports multiple concurrent sports and leagues. Matches can be created manually or fetched automatically from The Odds API:

- **Soccer:** Premier League, Champions League, La Liga, Bundesliga, Serie A, World Cup, Euro, Copa America, AFC Asian Cup
- **Basketball:** NBA, EuroLeague
- **American Football:** NFL
- **Baseball:** MLB
- **Ice Hockey:** NHL
- **Tennis:** ATP
- **MMA:** UFC/MMA
- **Cricket:** IPL
- Any other sport/tournament the admin configures or that The Odds API supports

### Supported Languages

The entire experience — app UI, card questions, push notifications — is fully localized:

| Language    | Code  | Priority |
| ----------- | ----- | -------- |
| **English** | `en`  | Launch   |
| **Persian** | `fa`  | Launch   |
| **Arabic**  | `ar`  | Launch   |
| **Turkish** | `tr`  | Phase 2  |
| **Spanish** | `es`  | Phase 2  |
| **French**  | `fr`  | Phase 3  |

- The app detects the device language and defaults to the closest supported locale, falling back to English.
- Users can override their language preference in-app settings.
- **Card questions are translated by the admin or AI** — each card has a question text per supported language. AI-generated cards include automatic translations in EN, FA, and AR.
- RTL (right-to-left) layout is fully supported for Persian and Arabic.

---

## 3. Core Game Mechanics

### 3.1 The Daily Basket

Every day, the system provides a **basket of 15 prediction cards**. Each card is a yes/no question about an upcoming match.

```
Example cards:
┌─────────────────────────────────────────────┐
│  🥇 GOLD                                    │
│  "Will Brazil score more than 3 goals       │
│   against Sweden?"                          │
│                                             │
│  YES: +20 pts    NO: +5 pts                 │
│  Expires: 18:00 UTC (kickoff)               │
│                                             │
│       ← SWIPE LEFT (No)                     │
│                    SWIPE RIGHT (Yes) →       │
│              ↑ SKIP ↑                        │
└─────────────────────────────────────────────┘
```

### 3.2 User Resources (Per Day)

| Resource          | Quantity | Purpose                          |
| ----------------- | -------- | -------------------------------- |
| **Answers**       | 10       | Respond Yes or No to a card      |
| **Skips**         | 5        | Burn a card without answering    |
| **Total Actions** | 15       | Every card is seen, every card requires a decision |

Since 10 answers + 5 skips = 15 = total cards, **every user sees every card**. There are no hidden cards. The strategic tension comes from deciding which 5 cards to skip and which 10 to answer — with no going back.

- **Answer (Yes/No):** Consumes 1 answer resource. The card is locked and scored after the match.
- **Skip:** Consumes 1 skip resource. The card is burned and cannot be revisited. No points awarded.
- **No skips remaining:** Once all 5 skips are used, every remaining card **must** be answered. The UI makes this clear ("No skips remaining — you must answer all remaining cards").
- **No action within 40 seconds:** The card is automatically skipped (consumes a skip if available, otherwise the card expires with no resource consumed).

### 3.3 Card Display Rules

- Only **one card** is visible at a time.
- Users **cannot see** previous or upcoming cards.
- Card order is **shuffled uniquely** per user (random permutation, server-side).
- There is **no going back** — once a card is answered or skipped, it's gone.

### 3.4 The 40-Second Timer

Each card has a **40-second countdown timer**. If the timer expires:

- The card is automatically skipped (consuming a skip resource if available).
- This creates urgency and prevents indefinite deliberation.
- If no skips remain and the timer expires, the card simply expires without any resource consumption.

### 3.5 Swipe UX

The primary interaction model is **swipe-based**:

| Gesture        | Action         |
| -------------- | -------------- |
| Swipe Right    | Answer **Yes** |
| Swipe Left     | Answer **No**  |
| Swipe Up / Tap | **Skip**       |

Swiping up to skip feels natural — like scrolling past content in a feed. This creates a fast, tactile, mobile-native experience similar to dating apps but for sports predictions.

### 3.6 Card Tiers & Scoring

Cards are grouped into **3 tiers** with fixed point values. The tier determines the visual style and the scoring asymmetry. Each tier has a **fixed count** per daily basket — this is always the same, so users can plan their strategy around it.

| Tier       | Count  | Points (Option A) | Points (Option B) | Color           | Strategy                             |
| ---------- | ------ | ------------------ | ------------------ | --------------- | ------------------------------------ |
| **Gold**   | 3      | +20 / +5           | +5 / +20           | Shiny, premium  | High risk/reward — big swing either way |
| **Silver** | 5      | +15 / +10          | +10 / +15          | Cool metallic   | Medium risk — slight edge on one side   |
| **White**  | 7      | +10 / +10          | +10 / +10          | Clean, minimal  | Safe — equal points either way          |

**How scoring works:**

- **Gold cards** are always asymmetric: one answer is worth +20, the other +5. The admin decides per card whether Yes or No is the high-reward answer. This means you can't blindly swipe one direction on Gold cards — you need to actually think about the prediction.
- **Silver cards** have a mild asymmetry: one answer is +15, the other +10. Less punishing than Gold but still rewards correct assessment of which side is favored.
- **White cards** are always balanced at +10/+10. Safe picks — you get the same points regardless of which answer is correct.

**Fixed counts matter:** Since every basket always has exactly 3 Gold + 5 Silver + 7 White = 15 cards, and every user sees all 15, users can develop strategies around tier management. A risk-loving player might save all their answers for Gold and Silver cards, skipping Whites. A conservative player might answer every White card for guaranteed points. The fixed distribution makes this strategic planning possible.

---

## 4. User Journey

### 4.1 Onboarding

```
┌─────────────────────────────────────────────────────┐
│                                                     │
│  1. Download XEX Play from App Store / Play Store   │
│                     ↓                               │
│  2. "Login with XEX Exchange" (Shared JWT)           │
│     └─ No XEX account? → Create one (redirects)    │
│                     ↓                               │
│  3. Quick tutorial (3 screens):                     │
│     - "You get 15 cards daily"                      │
│     - "10 answers, 5 skips — see every card"       │
│     - "Climb the leaderboard, earn token rewards"   │
│                     ↓                               │
│  4. Drop into first daily basket immediately        │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### 4.2 Daily Play Loop

```
User opens app
    ↓
System checks: Active session today?
    ├── YES → Resume from last card position
    └── NO  → Generate new session:
                - Select 15 cards from today's basket
                - Shuffle card order (random permutation)
                - Store session state
    ↓
Show Card #1 (40s timer starts)
    ↓
User decides: Answer (Yes/No) or Skip
    ↓
Card is locked, next card appears
    ↓
Repeat until:
    - All 15 cards processed (10 answered + 5 skipped), OR
    - All 10 answers used (remaining cards must be answered), OR
    - All remaining cards expired
    ↓
Session summary:
    - Cards answered: X/10
    - Predictions pending: waiting for match results
    - Current streak: N days
    ↓
Results trickle in as matches conclude
    ↓
Push notification: "Your prediction was correct! +20 pts"
```

### 4.3 Post-Match Resolution

After each match concludes:

1. The system resolves all cards tied to that match.
2. Correct predictions earn points based on the card's scoring profile.
3. Incorrect predictions earn 0 points.
4. Users receive push notifications for each resolved card.
5. Leaderboards update in real-time.

### 4.4 Leaderboard & Competition

Users can view their ranking and compare with others:

- **Daily leaderboard** — resets each day
- **Weekly leaderboard** — resets each week
- **Tournament leaderboard** — spans the entire tournament
- **All-time leaderboard** — lifetime points
- **Friend leaderboard** — compete with added friends

### 4.5 Exchange Conversion Touchpoints

At strategic moments, users are nudged toward XEX Exchange:

- After a big win: "Celebrate your streak! Trade on XEX Exchange with a fee discount."
- On leaderboard: "Top 100 players get 50% trading fee discount this week."
- In rewards screen: "Redeem 500 pts for exchange fee credits."
- Achievement unlocked: "Active traders get exclusive Gold cards."

---

## 5. Scoring & Ranking System

### 5.1 Points Calculation

Points are awarded only for **correct predictions**:

```
Points Earned = Card Profile Points for the chosen answer (Yes or No)
              × Correctness (1 if correct, 0 if wrong)
```

**Example:**

- Card: "Will Brazil score more than 3 goals?" (High-Risk profile)
- User answers: Yes (+20 if correct, 0 if wrong)
- Result: Brazil scores 4 goals → User earns **+20 points**

### 5.2 Daily Score

A user's daily score is the sum of points from all correctly predicted cards that day.

```
Daily Score = Σ (Points for each correct prediction)
Maximum possible daily score = (3 × 20) + (5 × 15) + (2 × 10) = 155 points
  (answering all 3 Gold correctly on the +20 side, all 5 Silver on the +15 side,
   and 2 of 7 White cards — skipping the other 5)
```

### 5.3 Tiebreaker Rules

When two users have the same score, ties are broken in order:

| Priority | Tiebreaker              | Rationale                           |
| -------- | ----------------------- | ----------------------------------- |
| 1st      | Fewer incorrect answers | Rewards accuracy over volume        |
| 2nd      | Higher-tier cards answered correctly | Rewards risk-taking on Gold/Silver cards |
| 3rd      | Earlier submission time | Rewards decisive, confident players |
| 4th      | Longer active streak    | Rewards consistent daily engagement |

### 5.4 Seasonal Resets

- **Daily scores** reset every day at midnight (server timezone).
- **Weekly scores** reset every Monday at midnight.
- **Tournament scores** reset when a new tournament begins.
- **All-time scores** never reset.
- Historical data is preserved for stats and profile pages.

---

## 6. Engagement & Growth Features

### 6.1 Leaderboards

Multiple leaderboard views to sustain competition:

| Leaderboard | Reset Cycle       | Purpose                             |
| ----------- | ----------------- | ----------------------------------- |
| Daily       | Every 24 hours    | Quick wins, immediate gratification |
| Weekly      | Every Monday      | Medium-term competition             |
| Tournament  | Per event         | Long-term tournament engagement     |
| All-Time    | Never             | Legacy and prestige                 |
| Friends     | Mirrors all above | Social competition                  |

### 6.2 Referral Program

Users can invite friends to earn bonus resources:

| Milestone                         | Reward                                    |
| --------------------------------- | ----------------------------------------- |
| Friend signs up via referral link | +1 bonus skip for referrer (next day)     |
| Friend completes first session    | +1 bonus answer for referrer (next day)   |
| 5 friends referred                | Exclusive referral badge + token bonus    |
| 10 friends referred               | Permanent +1 daily skip + token bonus     |

### 6.3 Achievements & Badges

Achievements provide long-term goals and collectibility:

| Achievement          | Condition                            | Badge          |
| -------------------- | ------------------------------------ | -------------- |
| First Prediction     | Answer your first card               | Rookie         |
| Perfect Day          | All 10 predictions correct in one day | Sharpshooter   |
| 10-Day Streak        | Play 10 consecutive days             | Dedicated      |
| 30-Day Streak        | Play 30 consecutive days             | Loyal Fan      |
| 100-Day Streak       | Play 100 consecutive days            | Legend         |
| Gold Hunter          | Correctly predict 10 Gold cards      | Risk Taker     |
| Social Butterfly     | Refer 5 friends                      | Connector      |
| Leaderboard Champion | Finish #1 on weekly leaderboard      | Champion       |
| Tournament Winner    | Finish #1 on tournament leaderboard  | Tournament MVP |
| Exchange Explorer    | Make first trade on XEX Exchange     | Trader         |

### 6.4 Social Sharing

Users can share their achievements and predictions:

- **Share prediction results:** "I predicted Brazil 3+ goals correctly! Play XEX Play and challenge me."
- **Share streaks:** "I'm on a 15-day prediction streak on XEX Play!"
- **Share leaderboard position:** "I'm #7 in this week's XEX Play leaderboard!"
- **Share badges:** Visual badge cards optimized for Instagram Stories, Twitter, and Telegram.

Each shared item includes a **deep link** back to XEX Play (organic growth).

### 6.5 Mini-Leagues / Private Groups

Users can create or join private groups to compete with friends within any tournament:

- Create a mini-league with a custom name and invite code.
- Share the invite code with friends.
- **Per-tournament leaderboard** — each mini-league tracks scores per active tournament, so friends compete head-to-head within the same event.
- Dedicated mini-league leaderboard within the group (daily, weekly, and tournament views).
- Group chat (future enhancement).
- Great for friend groups, office pools, sports clubs.

### 6.6 Push Notification Strategy

| Trigger                       | Notification                                        | Timing                 |
| ----------------------------- | --------------------------------------------------- | ---------------------- |
| Daily basket available        | "Your 15 cards are ready! Start predicting."        | Morning (configurable) |
| Card expiring soon            | "A Gold card expires in 30 min! Don't miss it."     | 30 min before match    |
| Prediction resolved (correct) | "You nailed it! +20 pts for Brazil 3+ goals."       | After match ends       |
| Prediction resolved (wrong)   | "Close one! Brazil didn't hit 3+ goals. Try again!" | After match ends       |
| Streak at risk                | "Don't break your 7-day streak! Play today."        | Evening if not played  |
| Friend joined via referral    | "Your friend Ali just joined! You earned +1 skip."  | On signup              |
| Leaderboard milestone         | "You moved up to #15 on the weekly leaderboard!"    | On rank change         |
| Token reward earned           | "You earned 10 tokens for finishing #5 today!"      | After daily/weekly end |
| New tournament started        | "UEFA Champions League is live! Start predicting."  | Event launch day       |

### 6.7 Streak System

The streak system rewards consistent daily play:

| Streak Length | Reward                                           |
| ------------- | ------------------------------------------------ |
| 3 days        | Visual streak badge on profile                   |
| 7 days        | +1 bonus skip for the next day                   |
| 10 days       | +1 bonus skip + token bonus                      |
| 14 days       | +1 bonus answer for the next day                 |
| 21 days       | Exclusive streak achievement badge               |
| 30 days       | +1 permanent daily skip (until streak breaks)    |

**Streak Rules:**

- A day counts as "played" if the user opens a session and answers at least 1 card.
- Missing a single day resets the streak to 0.
- Streak bonuses are applied the following day.
- The 10-day milestone grants a token bonus in addition to the gameplay benefit.

---

## 7. XEX Exchange Integration

XEX Play's primary business purpose is to funnel users to XEX Exchange. Every integration point is designed to feel natural and rewarding, not forced.

### 7.1 Shared Authentication (Shared JWT)

XEX Play is a **fully separate application** with its own database and API, it does **not** share a database with XEX Exchange. This is a deliberate security decision: the exchange handles wallets, private keys, KYC data, and financial balances. A game must never have access to that data.

Instead, XEX Play reuses the exchange's **JWT tokens** via a shared signing secret:

- Users log in through the **XEX Exchange app or website** (magic link, Google, Apple, or passkey, all existing exchange auth methods).
- The exchange issues a JWT signed with its `JWT_SECRET`.
- The Flutter app sends that same JWT to the **XEX Play API**, which validates it using the same shared secret.
- On first login, XEX Play creates a local user record linked by the exchange's `user_id` from the JWT `sub` claim.
- No separate registration, no separate password, no new auth system.
- New users who don't have an XEX Exchange account are redirected to create one on the exchange first.

**Why this approach?**

- Zero changes needed to the exchange codebase.
- Complete data isolation, XEX Play has no access to exchange wallets, balances, or KYC data.
- If XEX Play is ever compromised, the exchange is unaffected.
- Users get seamless SSO with a single account across both products.

### 7.2 Token Rewards

XEX Play rewards winners with **real tokens** that can be claimed and traded on XEX Exchange. This is the primary incentive loop — play well, earn tokens, go to the exchange to use them.

#### Daily Rewards

| Leaderboard Position | Daily Token Reward                 |
| -------------------- | ---------------------------------- |
| #1 (Daily)           | Large token prize                  |
| Top 3 (Daily)        | Medium token prize                 |
| Top 10 (Daily)       | Small token prize                  |
| Top 50 (Daily)       | Micro token prize                  |

Daily rewards create a reason to come back every single day. Even small token amounts add up over time.

#### Weekly Rewards

| Leaderboard Position | Weekly Reward                                        |
| -------------------- | ---------------------------------------------------- |
| Top 10 (Weekly)      | Token prize + 50% trading fee discount for 1 week   |
| Top 50 (Weekly)      | Token prize + 25% trading fee discount for 1 week   |
| Top 100 (Weekly)     | Token prize + 10% trading fee discount for 1 week   |

#### Tournament Rewards

| Leaderboard Position | Tournament Reward                                      |
| -------------------- | ------------------------------------------------------ |
| #1 (Tournament)      | Grand token prize + VIP exchange status                |
| Top 3 (Tournament)   | Large token prize pool distribution                    |
| Top 10 (Tournament)  | Medium token prize                                     |
| Top 50 (Tournament)  | Small token prize                                      |

Tournament prize pools are announced at the start of each tournament and funded by XEX Exchange.

### 7.3 Token Claim Flow

Tokens earned in XEX Play are credited to the user's XEX Exchange account:

1. User earns tokens through daily/weekly/tournament leaderboard placement.
2. Tokens accumulate in XEX Play's reward balance (visible in-app).
3. User taps "Claim Rewards" → tokens are transferred to their XEX Exchange account.
4. User can then trade, hold, or withdraw tokens on the exchange.

This flow ensures every winner visits XEX Exchange, completing the funnel.

### 7.4 Exclusive Cards for Active Traders

Users who actively trade on XEX Exchange receive exclusive benefits in XEX Play:

- **Exchange Insider:** Special prediction cards about crypto market events.
- **VIP Tier:** Access to exclusive high-value prediction cards.

### 7.5 In-App Exchange Prompts

Strategic, non-intrusive prompts throughout the user journey:

- Post-session: "Turn your prediction skills into trading profits on XEX Exchange."
- Reward screen: "You have 50 tokens to claim! Go to XEX Exchange to trade them."
- Achievement: "You're a prediction expert! Try your hand at trading."
- Leaderboard: "Top players trade on XEX Exchange. Join them."

### 7.6 Tournament Prize Pools

For major events (World Cup, Champions League), XEX Exchange funds token prize pools:

- Prize pools (in tokens) are announced at the start of each tournament.
- Distribution is based on final tournament leaderboard positions.
- Tokens are credited to XEX Play reward balance, claimable to Exchange account.
- Creates a compelling reason to both play XEX Play and use XEX Exchange.
- Larger tournaments = larger prize pools = more user acquisition.

---

## 8. Admin Capabilities

### 8.1 Event & Tournament Management

- Create and manage events/tournaments (e.g., "World Cup 2026", "Champions League 2025-26").
- Set tournament start/end dates.
- Configure scoring multipliers per tournament.
- Manage multiple concurrent events.

### 8.2 Match Management

- Add matches with teams, date/time, and associated event.
- Set match status (upcoming, live, completed).
- Input match results for card resolution.

### 8.3 Card/Question Management

- Create daily prediction cards (exactly 15 per day: 3 Gold + 5 Silver + 7 White).
- Set the question text for each card **in all supported languages** (admin enters translations per language).
- Assign card tier (Gold, Silver, White) — scoring is determined by tier.
- For Gold cards: choose which answer (Yes/No) gets the +20 (the other gets +5).
- For Silver cards: choose which answer gets the +15 (the other gets +10).
- White cards are always +10/+10 (no admin choice needed).
- Link cards to specific matches (for expiry timing).
- **Basket validation:** The system enforces exactly 3 Gold + 5 Silver + 7 White = 15 cards. A basket cannot be published with incorrect tier counts.
- Preview the daily basket before it goes live.

### 8.4 Reward & Prize Management

- Configure daily token reward amounts per leaderboard position.
- Configure weekly token reward amounts per leaderboard position.
- Set up tournament prize pools (total tokens, distribution percentages).
- View reward distribution history and audit trail.
- Manually grant token rewards to specific users if needed.

### 8.5 User Management

- View user list, stats, and activity.
- View user sessions and answers.
- Ban/suspend users for violations.
- Grant bonus resources (skips/answers) to specific users.

### 8.6 Leaderboard Management

- View and export leaderboard data.
- Manually adjust scores if needed (with audit trail).
- Configure leaderboard reset schedules.
- Set up tournament prize pools.

### 8.7 Translation Management

- Manage card question translations for all supported languages.
- Translation status dashboard: see which cards are missing translations for which languages.
- A basket cannot be published unless all its cards have translations for all active languages.
- Manage push notification templates per language.
- Manage static content translations (achievement names, event names, team names).

### 8.8 Analytics Dashboard

- Daily/weekly/monthly active users.
- Session completion rates.
- Card answer distribution (Yes vs No per card).
- User retention and streak statistics.
- Exchange conversion metrics (users who navigate to exchange).
- Referral program performance.

### 8.9 Push Notification Management

- Send custom push notifications to all users or segments.
- Configure automated notification triggers and timing.
- View notification delivery and open rates.

### 8.10 Sports Data Automation

XEX Play includes an AI-powered automation system that can run alongside manual content creation:

- **Sports Manager** — Toggle which sports/leagues are active for automation (14 sports across Soccer, Basketball, NFL, MLB, NHL, Tennis, MMA, Cricket).
- **Auto Match Fetching** — Upcoming matches are fetched from The Odds API every 6 hours with odds data, deduplicated by external ID.
- **AI Card Generation** — Claude Haiku 4.5 generates yes/no prediction questions with multi-language translations (EN, FA, AR) and proper tier distribution (Gold/Silver/White).
- **Auto-Resolution** — Cards are automatically resolved when matches complete, using AI to evaluate complex questions against actual results.
- **Automation Dashboard** — View job status, last run times, trigger jobs manually, and browse activity logs.
- **Source Badges** — Matches show "Auto"/"Manual", cards show "AI"/"Manual", events show "Auto"/"Manual" badges across the admin panel.

### 8.11 Settings Management

- Manage API keys and configuration directly from the admin UI (no server restart needed).
- Secret values (API keys) are masked in the interface with show/hide toggle.
- Inline editing with save/cancel/clear actions.
- Database-stored settings with environment variable fallback on startup.

---

## 9. Monetization Strategy

XEX Play is **free to play** and does not directly monetize users. Revenue is generated indirectly through XEX Exchange:

### Revenue Model

```
Free Game → User Engagement → Exchange Awareness → Trading Activity → Exchange Revenue
```

### Key Revenue Drivers

| Driver                     | Mechanism                                              |
| -------------------------- | ------------------------------------------------------ |
| **User Acquisition**       | Free game attracts sports fans who create XEX accounts |
| **Daily Engagement**       | Daily play keeps XEX Exchange top-of-mind              |
| **Token Rewards**          | Winners claim tokens on exchange, driving visits       |
| **Fee Discount Rewards**   | Winners are incentivized to trade to use discounts     |
| **Social Virality**        | Referrals and sharing bring new users organically      |
| **Tournament Prize Pools** | Prize pools attract competitive users to exchange      |
| **Brand Building**         | Association with sports events builds trust            |

### Why Not Direct Monetization?

- The game should feel **generous and accessible** to maximize reach.
- Paywalls or in-app purchases would **reduce the funnel** to XEX Exchange.
- The real value is in **user lifetime value** on the exchange, not game revenue.
- Free-to-play with exchange rewards creates a **unique competitive advantage** over paid prediction apps.

---

## 10. UI/UX Design Language

### Design Philosophy

XEX Play follows a **Coinbase-inspired design language** — clean, dark, professional, and enterprise-grade. The app must feel like it was built by a large, trusted fintech company. Every screen should convey security, quality, and confidence. No playful/cartoon aesthetics — this is a premium product backed by a real exchange.

### Reference Apps

| App | What to Study |
| --- | --- |
| **Coinbase / Coinbase Earn** | Overall visual language, card layouts, dark mode, quiz flow, token reward screens |
| **Sorare** | Card tier visual hierarchy (Gold/Silver/White shimmer effects) |
| **Duolingo** | Streak UI, daily progress, celebration animations |
| **Tinder** | Swipe card stack physics, gesture feedback, card transitions |

### Color System

Built on a **dark-first palette** inspired by Coinbase's design system. Dark mode is the default.

#### Core Palette

| Token | Dark Mode | Light Mode | Usage |
| --- | --- | --- | --- |
| `background` | `#0A0B0D` (Woodsmoke) | `#FFFFFF` | App background |
| `surface` | `#141519` | `#F5F8FF` | Card surfaces, bottom sheets, modals |
| `surfaceRaised` | `#1E1F25` | `#EBEDF2` | Elevated cards, active states |
| `primary` | `#587BFA` | `#0052FF` | Primary actions, links, active indicators |
| `primaryBold` | `#0052FF` | `#0052FF` | High-emphasis buttons, key CTAs |
| `positive` | `#09A85A` | `#098551` | Correct predictions, success states, streaks |
| `negative` | `#CF202F` | `#CF202F` | Incorrect predictions, errors |
| `warning` | `#ED702F` | `#ED702F` | Timer warnings, expiring cards |
| `textPrimary` | `#FFFFFF` | `#0A0B0D` | Headlines, card questions |
| `textSecondary` | `#8A8F98` | `#5B616E` | Captions, labels, metadata |
| `textTertiary` | `#555961` | `#9DA3AE` | Disabled text, hints |
| `border` | `#1E1F25` | `#E2E4E9` | Dividers, card outlines |
| `borderSubtle` | `#141519` | `#F0F1F3` | Subtle separators |

#### Card Tier Colors

| Tier | Accent | Gradient / Effect |
| --- | --- | --- |
| **Gold** | `#FFD700` → `#FFA500` | Shimmering gold gradient border + subtle glow. Premium feel — metallic shine effect on card surface. |
| **Silver** | `#C0C0C0` → `#8A9BAE` | Cool metallic gradient border. Polished steel look, lighter than Gold but still distinct. |
| **White** | `#E2E4E9` → `#FFFFFF` | Clean, minimal border. No shimmer — subtle and calm. Safe and understated. |

### Typography

Use a **system font stack with a monospaced accent** — inspired by Coinbase's use of monospaced typefaces for financial/data contexts.

| Style | Font | Size | Weight | Line Height | Usage |
| --- | --- | --- | --- | --- | --- |
| `display1` | System Sans | 32px (2rem) | 600 | 40px | Splash headlines, tournament names |
| `title1` | System Sans | 24px (1.5rem) | 600 | 32px | Screen titles, leaderboard headers |
| `title2` | System Sans | 20px (1.25rem) | 600 | 28px | Section headers, card tier labels |
| `headline` | System Sans | 16px (1rem) | 600 | 24px | Card question text, button labels |
| `body` | System Sans | 16px (1rem) | 400 | 24px | Descriptions, explanations |
| `label` | System Sans | 14px (0.875rem) | 600 | 20px | Badges, tags, small buttons |
| `caption` | System Sans | 13px (0.8125rem) | 400 | 16px | Timestamps, metadata, helper text |
| `mono` | System Mono | 14px (0.875rem) | 600 | 20px | Points, scores, timers, countdowns |

- **System Sans** = SF Pro (iOS) / Roboto (Android) — native feel, no custom font loading.
- **System Mono** = SF Mono (iOS) / Roboto Mono (Android) — used for all numerical data (scores, timers, points) to give a financial-data feel.
- All numerical displays (points, timers, leaderboard ranks) use the **mono** style for a data-dashboard aesthetic.
- RTL languages (Persian, Arabic) use system defaults with proper mirroring.

### Spacing & Layout

Based on an **8px grid system** (matching Coinbase Design System):

| Token | Value | Usage |
| --- | --- | --- |
| `xs` | 4px | Tight gaps (icon-to-label) |
| `sm` | 8px | Inline spacing, small paddings |
| `md` | 16px | Standard padding, card internal spacing |
| `lg` | 24px | Section spacing, card-to-card gaps |
| `xl` | 32px | Screen edge padding, major section breaks |
| `xxl` | 48px | Top/bottom safe areas, hero spacing |

**Layout Rules:**
- Screen horizontal padding: `xl` (32px) on phones, responsive on tablets.
- Card internal padding: `md` (16px) on all sides.
- Between cards in a stack: `lg` (24px).
- Bottom navigation bar: 56px height + safe area inset.

### Border Radius

| Token | Value | Usage |
| --- | --- | --- |
| `none` | 0px | Flat edges (dividers) |
| `sm` | 8px | Small buttons, tags, badges |
| `md` | 12px | Cards, input fields, bottom sheets |
| `lg` | 16px | Modal dialogs, large cards |
| `xl` | 24px | Floating action buttons, pill buttons |
| `full` | 100000px | Avatars, circular indicators, round buttons |

### Elevation & Shadows

| Token | Value | Usage |
| --- | --- | --- |
| `elevation0` | none | Flat, inline elements |
| `elevation1` | `0px 8px 12px rgba(0,0,0,0.12)` | Cards, bottom sheets |
| `elevation2` | `0px 8px 24px rgba(0,0,0,0.12)` | Modals, overlays, floating elements |

In dark mode, shadows are less visible — rely on **surface color differentiation** instead of shadows to create depth (`background` → `surface` → `surfaceRaised`).

### Component Design Rules

#### Prediction Cards (Core UI)

- **Card stack:** Only one card visible at a time, with a subtle peek of the next card behind (2-3px offset, slightly scaled down, blurred).
- **Card surface:** Uses `surface` color with `md` border radius. Card tier determines the border style (Gold shimmer / Silver metallic / White clean).
- **Question text:** Centered, `headline` style, max 3 lines. If longer, scale down to `body`.
- **Points display:** Bottom of card, `mono` style. Show both outcomes: "Yes: +20 pts / No: +5 pts". Use `positive` color for the higher-value side.
- **Timer:** Top-right corner, circular progress ring. `mono` style. Turns `warning` color below 10 seconds. Turns `negative` below 5 seconds.
- **Swipe indicators:** Directional arrows appear with opacity based on swipe distance. Right swipe reveals green "YES" overlay. Left swipe reveals red "NO" overlay. Up swipe reveals neutral "SKIP" overlay.
- **Card transition:** Swiped card flies off-screen with physics-based animation (velocity-matched). Next card scales up from 0.95 to 1.0 with a spring curve.
- **Gold card shimmer:** Animated gradient that subtly moves across the card border, like light reflecting off metal. Not distracting — elegant and premium.

#### Buttons

- **Primary:** `primaryBold` background, white text, `md` border radius, 48px height. Used for main CTAs ("Start Today's Session", "Claim Rewards").
- **Secondary:** Transparent background, `primary` text, 1px `primary` border, `md` radius, 48px height.
- **Ghost:** No background, no border, `primary` text. Used for tertiary actions.
- **Destructive:** `negative` background, white text. Used sparingly.
- All buttons use `headline` typography (16px, weight 600).
- Pressed state: 0.96 scale + opacity 0.8 (quick spring animation).

#### Bottom Navigation

- 4 tabs: **Play**, **Leaderboard**, **Rewards**, **Profile**.
- Dark surface (`surface` color), 56px + safe area.
- Active tab: `primary` color icon + label. Inactive: `textTertiary`.
- Icons: 24px, outlined style (inactive), filled (active).
- No floating tab bar — flush with bottom edge for a professional, non-playful look.

#### Leaderboard

- Rows are simple list items — rank number (`mono`), avatar (32px, `full` radius), username (`label`), points (`mono`, right-aligned).
- Top 3 have special treatment: slightly larger row, Gold/Silver/Bronze accent on rank number.
- Current user's row is highlighted with `surfaceRaised` background and `primary` left border.
- Sticky header with leaderboard type selector (Daily / Weekly / Tournament / All-Time) as a horizontal pill-button group.

#### Session Summary / Results

- Full-screen card with stats in a clean grid layout.
- Large `display1` score at top.
- Stats in 2-column grid: cards answered, correct predictions, streak count, rank change.
- Each stat uses `mono` for the number and `caption` for the label.
- "Share Results" button generates a branded image card (dark background, Gold accent, score, streak — optimized for Instagram Stories / Twitter).
- Positive rank change: green arrow + `positive` color. Negative: red arrow + `negative`.

#### Streak & Achievement Badges

- Badges are 48px circular icons with a subtle glow matching their achievement color.
- Streak badge: ring progress indicator around the badge icon.
- Unlocked badges: full color. Locked: `textTertiary` with a lock overlay.
- Achievement unlock: full-screen celebration overlay with confetti particles (tasteful, not over-the-top — similar to Coinbase's reward claim animation).

### Motion & Animation

| Element | Animation | Duration | Curve |
| --- | --- | --- | --- |
| Card swipe | Physics-based throw | Velocity-matched | Spring (damping: 0.8) |
| Card enter | Scale 0.95→1.0 + fade in | 300ms | Spring (damping: 0.9) |
| Timer pulse | Scale 1.0→1.05→1.0 | 600ms | Ease-in-out (loops below 10s) |
| Button press | Scale 1.0→0.96 | 100ms | Ease-out |
| Tab switch | Cross-fade content | 200ms | Ease-in-out |
| Score count-up | Number rolls from 0 to final | 800ms | Ease-out (decelerate) |
| Achievement unlock | Scale 0→1.1→1.0 + glow | 500ms | Spring |
| Confetti burst | Particle physics | 2000ms | Gravity + drag |

- All animations must be **subtle and professional** — never flashy or juvenile.
- Respect `reduceMotion` accessibility setting — disable non-essential animations.
- Card swipe must feel **physically responsive** — the card follows the finger with zero lag.

### Accessibility

- Minimum contrast ratio: **4.5:1** for text, **3:1** for large text and UI components.
- All interactive elements: minimum **44x44px** touch target.
- Screen reader labels on all swipe actions ("Swipe right for Yes, left for No, up to Skip").
- Timer countdown announced to screen readers at 10s and 5s.
- Support for system font scaling (up to 200%).
- Full RTL layout mirroring for Persian and Arabic.

### Overall Aesthetic Summary

> **Dark, clean, trustworthy.** Think Coinbase meets a sports data dashboard. Every pixel should say "this is backed by a real exchange with real money." No rounded-cartoon icons, no playful illustrations, no bright gradients. Instead: sharp typography, monospaced numbers, metallic card tiers, and physics-based interactions. The app should feel like Bloomberg Terminal had a baby with Tinder — data-rich but swipeable, professional but fun.

---

_This document serves as the definitive product reference for XEX Play. All design and development decisions should align with the principles and mechanics described here._
