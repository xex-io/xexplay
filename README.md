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
8. [Admin Capabilities](#8-admin-capabilities)
9. [Monetization Strategy](#9-monetization-strategy)

---

## 1. Overview

### What is XEX Play?

XEX Play is a card-based sports prediction game designed to engage users during live sports tournaments and create a daily habit loop. Each day, users receive a curated basket of prediction cards tied to real upcoming matches. With limited resources (5 answers and 3 skips), users must strategically decide which predictions to make and which to skip, creating genuine risk/reward decisions rather than mindless guessing.

### Elevator Pitch

> "Pick your battles. 15 cards, 5 answers, 3 skips, predict sports outcomes, climb the leaderboard, and earn real rewards on XEX Exchange."

### Value Proposition

| For Users                          | For XEX Exchange                           |
| ---------------------------------- | ------------------------------------------ |
| Free, engaging daily sports game   | User acquisition funnel                    |
| Strategic gameplay, not just luck  | Daily active user engagement               |
| Compete with friends and globally  | Conversion to exchange trading             |
| Earn real rewards (exchange perks) | Brand association with major sports events |
| Quick sessions (~5 minutes/day)    | Cross-promotion channel                    |

### Design Principles

- **Daily Habit Formation** — The game is designed to bring users back every single day through limited daily resources, streaks, and expiring cards.
- **Real Decision-Making** — Limited answers and skips force users to evaluate risk/reward on every card.
- **Anti-Spam Scoring** — The scoring system prevents users from blindly answering only high-point cards. Resource management is key.
- **Live Sports Connection** — Cards are directly tied to real matches with real expiry times, keeping the game grounded in reality.
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

### Supported Events (Multi-Event Architecture)

The platform is designed to support multiple concurrent and sequential events:

- FIFA World Cup
- UEFA Champions League
- UEFA Euro
- Copa America
- Domestic leagues (Premier League, La Liga, Serie A, Bundesliga, etc.)
- AFC Asian Cup
- Any other sports tournament the admin configures

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
│              ↓ SKIP ↓                        │
└─────────────────────────────────────────────┘
```

### 3.2 User Resources (Per Day)

| Resource          | Quantity | Purpose                          |
| ----------------- | -------- | -------------------------------- |
| **Answers**       | 5        | Respond Yes or No to a card      |
| **Skips**         | 3        | Burn a card without answering    |
| **Total Actions** | 8        | Maximum interactions per session |

- **Answer (Yes/No):** Consumes 1 answer resource. The card is locked and scored after the match.
- **Skip:** Consumes 1 skip resource. The card is burned and cannot be revisited. No points awarded.
- **No action within 40 seconds:** The card is automatically skipped (consumes a skip if available, otherwise the card expires).

### 3.3 Card Display Rules

- Only **one card** is visible at a time.
- Users **cannot see** previous or upcoming cards.
- Card order is **shuffled uniquely** per user (via the Smart Shuffle Algorithm).
- There is **no going back** — once a card is answered or skipped, it's gone.

### 3.4 The 40-Second Timer

Each card has a **40-second countdown timer**. If the timer expires:

- The card is automatically skipped (consuming a skip resource if available).
- This creates urgency and prevents indefinite deliberation.
- If no skips remain and the timer expires, the card simply expires without any resource consumption.

### 3.5 Swipe UX

The primary interaction model is **swipe-based**:

| Gesture          | Action         |
| ---------------- | -------------- |
| Swipe Right      | Answer **Yes** |
| Swipe Left       | Answer **No**  |
| Swipe Down / Tap | **Skip**       |

This creates a fast, tactile, mobile-native experience similar to dating apps but for sports predictions.

### 3.6 Card Scoring Profiles

Each card is assigned a **scoring profile** by the admin. The profile determines how many points the Yes and No answers are worth. The scoring asymmetry is the core strategic element, users must decide whether the risk is worth the reward.

| Profile       | Yes Points | No Points | Color  | Strategy                           |
| ------------- | ---------- | --------- | ------ | ---------------------------------- |
| **Balanced**  | +10        | +10       | White  | Safe pick, equal either way        |
| **Lean-Yes**  | +20        | +10       | Bronze | High reward for Yes, decent for No |
| **Lean-No**   | +10        | +20       | Silver | High reward for No, decent for Yes |
| **High-Risk** | +20        | +5        | Gold   | Huge reward for Yes, low for No    |
| **Low-Risk**  | +10        | +5        | White  | Moderate reward, low-risk No       |

### 3.7 Card Tiers

Cards are visually distinguished by their tier, which correlates with their scoring profile:

| Tier       | Visual Style        | Typical Profile     | Rarity in Basket |
| ---------- | ------------------- | ------------------- | ---------------- |
| **White**  | Clean, minimal      | Balanced / Low-Risk | Common (6–8)     |
| **Bronze** | Warm metallic       | Lean-Yes            | Uncommon (3–4)   |
| **Silver** | Cool metallic       | Lean-No             | Rare (2–3)       |
| **Gold**   | Shiny, premium feel | High-Risk           | Very Rare (1–2)  |

### 3.8 Smart Shuffle Distribution

The system guarantees a fair distribution in the first 8 cards each user sees (since they can take at most 8 actions: 5 answers + 3 skips):

| Tier   | Guaranteed in First 8 |
| ------ | --------------------- |
| Gold   | 1 card                |
| Silver | 2 cards               |
| Bronze | 2 cards               |
| White  | 3 cards               |

The remaining 7 cards in the basket serve as a buffer for expired cards and provide variety. The shuffle order is randomized per user and stored server-side.

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
│     - "5 answers, 3 skips, choose wisely"          │
│     - "Climb the leaderboard, earn rewards"         │
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
                - Shuffle order (Smart Shuffle Algorithm)
                - Store session state
    ↓
Show Card #1 (40s timer starts)
    ↓
User decides: Answer (Yes/No) or Skip
    ↓
Card is locked, next card appears
    ↓
Repeat until:
    - All 5 answers used, OR
    - All 5 answers + 3 skips used, OR
    - All remaining cards expired, OR
    - No cards left in basket
    ↓
Session summary:
    - Cards answered: X/5
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
Maximum possible daily score = 5 × 20 = 100 points (if all 5 answers are correct on max-point cards)
```

### 5.3 Tiebreaker Rules

When two users have the same score, ties are broken in order:

| Priority | Tiebreaker              | Rationale                           |
| -------- | ----------------------- | ----------------------------------- |
| 1st      | Fewer incorrect answers | Rewards accuracy over volume        |
| 2nd      | Earlier submission time | Rewards decisive, confident players |
| 3rd      | Longer active streak    | Rewards consistent daily engagement |

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

| Milestone                         | Reward                                  |
| --------------------------------- | --------------------------------------- |
| Friend signs up via referral link | +1 bonus skip for referrer (next day)   |
| Friend completes first session    | +1 bonus answer for referrer (next day) |
| 5 friends referred                | Exclusive referral badge                |
| 10 friends referred               | Permanent +1 daily skip                 |

### 6.3 Achievements & Badges

Achievements provide long-term goals and collectibility:

| Achievement          | Condition                            | Badge          |
| -------------------- | ------------------------------------ | -------------- |
| First Prediction     | Answer your first card               | Rookie         |
| Perfect Day          | All 5 predictions correct in one day | Sharpshooter   |
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

Users can create or join private groups to compete:

- Create a mini-league with a custom name and invite code.
- Share the invite code with friends.
- Dedicated leaderboard within the group.
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
| New tournament started        | "UEFA Champions League is live! Start predicting."  | Event launch day       |

### 6.7 Streak System

The streak system rewards consistent daily play:

| Streak Length | Reward                                           |
| ------------- | ------------------------------------------------ |
| 3 days        | Visual streak badge on profile                   |
| 7 days        | +1 bonus skip for the next day                   |
| 10 days       | +1 bonus skip + 1 guaranteed Gold card in basket |
| 14 days       | +1 bonus answer for the next day                 |
| 21 days       | Exclusive streak achievement badge               |
| 30 days       | +1 permanent daily skip (until streak breaks)    |

**Streak Rules:**

- A day counts as "played" if the user opens a session and answers at least 1 card.
- Missing a single day resets the streak to 0.
- Streak bonuses are applied the following day.
- The 10-day milestone matches the original spec: +1 skip and seeing one additional Gold card (+20 pts).

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

### 7.2 Leaderboard Rewards

| Leaderboard Position | Exchange Reward                                |
| -------------------- | ---------------------------------------------- |
| Top 10 (Weekly)      | 50% trading fee discount for 1 week            |
| Top 50 (Weekly)      | 25% trading fee discount for 1 week            |
| Top 100 (Weekly)     | 10% trading fee discount for 1 week            |
| Top 3 (Tournament)   | Prize pool distribution (funded by exchange)   |
| #1 (Tournament)      | VIP exchange status + premium trading features |

### 7.3 Points Redemption

Users can redeem accumulated XEX Play points for exchange perks:

| Points | Reward                                   |
| ------ | ---------------------------------------- |
| 200    | Small trading fee credit                 |
| 500    | Medium trading fee credit                |
| 1000   | Bonus exchange tokens                    |
| 2500   | Premium exchange feature unlock (1 week) |

### 7.4 Exclusive Cards for Active Traders

Users who actively trade on XEX Exchange receive exclusive benefits in XEX Play:

- **Trader Card Pack:** Extra 2 cards added to daily basket (17 total).
- **Exchange Insider:** Special prediction cards about crypto market events.
- **VIP Tier:** Access to exclusive high-value prediction cards.

### 7.5 In-App Exchange Prompts

Strategic, non-intrusive prompts throughout the user journey:

- Post-session: "Turn your prediction skills into trading profits on XEX Exchange."
- Reward screen: "Your 500 points can become fee credits. Redeem now?"
- Achievement: "You're a prediction expert! Try your hand at trading."
- Leaderboard: "Top players trade on XEX Exchange. Join them."

### 7.6 Tournament Prize Pools

For major events (World Cup, Champions League), XEX Exchange funds prize pools:

- Prize pools are announced at the start of each tournament.
- Distribution is based on final tournament leaderboard positions.
- Prizes are credited directly to the user's XEX Exchange account.
- Creates a compelling reason to both play XEX Play and use XEX Exchange.

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

- Create daily prediction cards (up to 15 per day).
- Set the question text for each card.
- Assign a scoring profile (Balanced, Lean-Yes, Lean-No, High-Risk, Low-Risk).
- Assign card tier (White, Bronze, Silver, Gold).
- Link cards to specific matches (for expiry timing).
- Preview the daily basket before it goes live.

### 8.4 Scoring Profile Management

- View and manage scoring profiles.
- Create custom profiles if needed (beyond the default 5).
- Assign profiles to cards.

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

### 8.7 Analytics Dashboard

- Daily/weekly/monthly active users.
- Session completion rates.
- Card answer distribution (Yes vs No per card).
- User retention and streak statistics.
- Exchange conversion metrics (users who navigate to exchange).
- Referral program performance.

### 8.8 Push Notification Management

- Send custom push notifications to all users or segments.
- Configure automated notification triggers and timing.
- View notification delivery and open rates.

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
| **Points Redemption**      | Users visit exchange to redeem points                  |
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

_This document serves as the definitive product reference for XEX Play. All design and development decisions should align with the principles and mechanics described here._
