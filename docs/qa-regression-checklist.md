# XEX Play — QA Regression Checklist

> Full manual QA regression script for task **4.11.6** — covers all features across Phases 1-4.
>
> **How to use:** Work through each section sequentially. Check off each item as it passes. Record failures with the test case number and observed behavior.

---

## 1. Authentication

### 1.1 Login via Exchange JWT

- [ ] **AUTH-01** — Open app with no stored token. **Expected:** Splash screen redirects to login screen with "Login with XEX Exchange" button.
- [ ] **AUTH-02** — Tap "Login with XEX Exchange" and complete the Exchange login flow. **Expected:** Exchange JWT is received, `POST /auth/login` succeeds, user is upserted in Play DB with correct `xex_user_id` from JWT `sub` claim, and app navigates to home screen.
- [ ] **AUTH-03** — Login with a brand-new Exchange account (first time in Play). **Expected:** New user record is created with `display_name`, `email`, `avatar_url` from JWT claims; `total_points` = 0; unique `referral_code` is generated.
- [ ] **AUTH-04** — Login with an existing Play account. **Expected:** Existing user record is returned (not duplicated); game state (points, streaks, rewards) is preserved.
- [ ] **AUTH-05** — Close and reopen the app while JWT is still valid. **Expected:** Splash screen validates stored JWT and auto-navigates to home screen without requiring re-login.

### 1.2 Token Expiry & Refresh

- [ ] **AUTH-06** — Make an API call with an expired JWT (> 24h old). **Expected:** Backend returns 401 Unauthorized; app detects token expiry via error interceptor and redirects to login screen.
- [ ] **AUTH-07** — Make an API call with a tampered/invalid JWT signature. **Expected:** Backend returns 401 Unauthorized.
- [ ] **AUTH-08** — Make an API call with a JWT missing required claims (`user_id`, `email`). **Expected:** Backend returns 401 Unauthorized.
- [ ] **AUTH-09** — Make an API call with a JWT signed with a different secret. **Expected:** Backend returns 401 Unauthorized.

### 1.3 Unauthorized Access

- [ ] **AUTH-10** — Call any authenticated endpoint (`/me`, `/sessions/*`, `/leaderboards/*`) without a Bearer token. **Expected:** Backend returns 401 Unauthorized.
- [ ] **AUTH-11** — Call any admin endpoint (`/admin/*`) with a valid user JWT that has a non-admin role. **Expected:** Backend returns 403 Forbidden.
- [ ] **AUTH-12** — Call any admin endpoint with a valid admin JWT. **Expected:** Request succeeds.
- [ ] **AUTH-13** — Tap logout. **Expected:** Stored JWT is cleared from secure storage, app navigates to login screen, subsequent API calls fail with 401.

---

## 2. Game Session

### 2.1 Start Session

- [ ] **GAME-01** — From home screen, with a published basket for today, tap "Start Session". **Expected:** `POST /sessions/start` creates a new session with 15 cards shuffled (Fisher-Yates, deterministic per user+date), `current_index` = 0, `answers_used` = 0, `skips_used` = 0, status = `in_progress`. App displays the first card.
- [ ] **GAME-02** — Home screen with no published basket for today. **Expected:** Start button is disabled or hidden; message indicates no basket is available.
- [ ] **GAME-03** — Home screen with an already-completed session for today. **Expected:** Status shows "Completed" with session summary; start button is not available.
- [ ] **GAME-04** — Verify the shuffle order differs between two different users for the same basket. **Expected:** Card order is different (deterministic per `user_id + date` seed).

### 2.2 Answer Cards

- [ ] **GAME-05** — Swipe right on a card (Yes answer). **Expected:** Green "YES" overlay appears during swipe; card flies off-screen with velocity-matched animation; `POST /sessions/current/answer` is called with `answer: true`; `answers_used` increments by 1; next card scales from 0.95 to 1.0 with spring animation (300ms, damping 0.9).
- [ ] **GAME-06** — Swipe left on a card (No answer). **Expected:** Red "NO" overlay appears during swipe; card flies off-screen; answer recorded with `answer: false`; `answers_used` increments by 1.
- [ ] **GAME-07** — Swipe below the threshold distance, then release. **Expected:** Card springs back to center position; no answer is recorded; `answers_used` unchanged.
- [ ] **GAME-08** — Answer all 10 allowed answers. **Expected:** After the 10th answer, if skips remain, swipe gestures for Yes/No are disabled; only Skip (swipe up) is available for remaining cards.

### 2.3 Skip Cards

- [ ] **GAME-09** — Swipe up on a card (Skip). **Expected:** Gray "SKIP" overlay appears; card animates out; `POST /sessions/current/skip` is called; `skips_used` increments by 1; no points awarded.
- [ ] **GAME-10** — Use all 5 skips. **Expected:** After the 5th skip, swipe-up gesture is disabled; remaining cards must be answered (Yes/No only).
- [ ] **GAME-11** — Attempt to skip when all skips are exhausted (via API). **Expected:** Backend returns 400 Bad Request with "no skips remaining" error.

### 2.4 Timer

- [ ] **GAME-12** — New card is presented. **Expected:** 40-second circular countdown timer starts; mono font displays remaining seconds; timer ring depletes smoothly.
- [ ] **GAME-13** — Timer reaches < 10 seconds. **Expected:** Timer turns warning color; pulse animation begins.
- [ ] **GAME-14** — Timer reaches < 5 seconds. **Expected:** Timer turns negative/danger color.
- [ ] **GAME-15** — Timer reaches 0 without user action. **Expected:** Card is auto-skipped (server enforces 42s with 2s grace); next card is presented; `skips_used` increments.
- [ ] **GAME-16** — Submit an answer after the 40s timer has expired but within the 2s server grace window. **Expected:** Answer is accepted.
- [ ] **GAME-17** — Submit an answer after the 42s server deadline. **Expected:** Backend rejects with 400; card is treated as auto-skipped.

### 2.5 Resume Session

- [ ] **GAME-18** — Start a session, answer 5 cards, then force-close the app. Reopen the app. **Expected:** `POST /sessions/start` (or `GET /sessions/current`) detects existing in-progress session; resumes at card 6 with correct `answers_used` and `skips_used` counts; previously answered cards are not replayed.
- [ ] **GAME-19** — Lose network connectivity mid-session. **Expected:** App shows offline indicator; pending actions are queued or an error is displayed; no data is lost.
- [ ] **GAME-20** — Regain connectivity after brief offline period. **Expected:** Session resumes from server state; any missed timer expirations are resolved server-side.

### 2.6 Session Completion

- [ ] **GAME-21** — Process all 15 cards (10 answers + 5 skips). **Expected:** Session status changes to `completed`; full-screen session summary is displayed with: large total score (display1 + mono font), cards answered count, correct predictions count, points earned, streak count, rank change (green/red arrow).
- [ ] **GAME-22** — On session summary, verify score counts up with animation (800ms ease-out). **Expected:** Score animates from 0 to final value.
- [ ] **GAME-23** — Tap "Done" on session summary. **Expected:** Returns to home screen; home screen shows today's session as completed.

### 2.7 Resource Counters

- [ ] **GAME-24** — During a session, verify the resource counters display correctly. **Expected:** "Answers remaining" shows `10 - answers_used`; "Skips remaining" shows `5 - skips_used` (plus any bonus resources from streaks).
- [ ] **GAME-25** — Verify card progress indicator. **Expected:** Shows "Card X of 15" (e.g., "Card 7 of 15") and updates after each action.

---

## 3. Card Tiers

### 3.1 Visual Differentiation

- [ ] **TIER-01** — Display a Gold tier card. **Expected:** Card has a gold shimmer border effect; tier label says "Gold".
- [ ] **TIER-02** — Display a Silver tier card. **Expected:** Card has a silver metallic border effect; tier label says "Silver".
- [ ] **TIER-03** — Display a White tier card. **Expected:** Card has a clean white border; tier label says "White".
- [ ] **TIER-04** — Display a VIP tier card (for active Exchange traders). **Expected:** Card has a distinct VIP indicator; trader badge is visible.

### 3.2 Scoring — Correct Answers

- [ ] **TIER-05** — Answer a Gold card correctly. **Expected:** `points_earned` = 20 after card is resolved.
- [ ] **TIER-06** — Answer a Silver card correctly. **Expected:** `points_earned` = 15 after card is resolved.
- [ ] **TIER-07** — Answer a White card correctly. **Expected:** `points_earned` = 10 after card is resolved.

### 3.3 Scoring — Incorrect Answers

- [ ] **TIER-08** — Answer a Gold card incorrectly. **Expected:** `points_earned` = -5 (penalty) after card is resolved.
- [ ] **TIER-09** — Answer a Silver card incorrectly. **Expected:** `points_earned` = -10 (penalty) after card is resolved.
- [ ] **TIER-10** — Answer a White card incorrectly. **Expected:** `points_earned` = -10 (penalty) after card is resolved.

### 3.4 Basket Composition

- [ ] **TIER-11** — Verify a published basket always contains exactly 3 Gold + 5 Silver + 7 White = 15 cards. **Expected:** Backend enforces this composition on basket publish; admin cannot publish a basket with incorrect tier counts.
- [ ] **TIER-12** — Attempt to publish a basket with incorrect tier composition via admin panel. **Expected:** Publish is rejected with a validation error showing the expected counts.

### 3.5 Event Scoring Multiplier

- [ ] **TIER-13** — Play a session in an event with `scoring_multiplier` > 1.0. **Expected:** All points for that event's cards are multiplied accordingly.

---

## 4. Leaderboards

### 4.1 Leaderboard Types

- [ ] **LB-01** — Navigate to Leaderboard tab, select "Daily". **Expected:** Shows today's rankings sorted by `total_points` descending; user's own rank is always visible (highlighted row with `surfaceRaised` background + primary left border).
- [ ] **LB-02** — Select "Weekly". **Expected:** Shows current week's cumulative rankings.
- [ ] **LB-03** — Select "Tournament" and choose an active event. **Expected:** Shows rankings scoped to that event.
- [ ] **LB-04** — Select "All-Time". **Expected:** Shows lifetime cumulative rankings.
- [ ] **LB-05** — Select "Friends". **Expected:** Shows rankings filtered to users connected via referrals or mini-leagues.

### 4.2 Leaderboard Display

- [ ] **LB-06** — Verify top 3 entries have Gold, Silver, Bronze accent colors respectively. **Expected:** Visual distinction for rank 1, 2, 3.
- [ ] **LB-07** — Verify each row shows: rank (mono font), avatar (32px, full radius), username (label), points (mono, right-aligned). **Expected:** All elements present and correctly formatted.
- [ ] **LB-08** — Scroll down the leaderboard past the initial page. **Expected:** Pagination loads more entries; no duplicate rows; loading indicator during fetch.

### 4.3 Real-Time Updates

- [ ] **LB-09** — Have a card resolved while viewing the leaderboard. **Expected:** WebSocket delivers `leaderboard_update` event; leaderboard refreshes with new rankings without manual pull-to-refresh.
- [ ] **LB-10** — Pull to refresh on the leaderboard. **Expected:** Data is refreshed from the API; loading indicator is shown during refresh.

### 4.4 Tiebreaker Logic

- [ ] **LB-11** — Two users have the same `total_points`. **Expected:** Tiebreaker applies in order: (1) fewer incorrect answers, (2) more higher-tier correct answers, (3) earlier submission time, (4) longer streak. The user with the better tiebreaker metric ranks higher.

### 4.5 Leaderboard Resets

- [ ] **LB-12** — After midnight (server time), check the daily leaderboard. **Expected:** Daily leaderboard is reset to empty; previous day's data is archived; rewards for the previous day are distributed.
- [ ] **LB-13** — After Monday midnight, check the weekly leaderboard. **Expected:** Weekly leaderboard is reset; previous week's rewards are distributed.

---

## 5. Streaks

### 5.1 Streak Tracking

- [ ] **STREAK-01** — Complete a session for the first time. **Expected:** `current_streak` = 1; `last_played_date` = today.
- [ ] **STREAK-02** — Complete a session on the next consecutive day. **Expected:** `current_streak` = 2; `longest_streak` updates if applicable.
- [ ] **STREAK-03** — Skip a day (no session completed), then play the following day. **Expected:** `current_streak` resets to 1; `longest_streak` retains its previous value.

### 5.2 Streak Display

- [ ] **STREAK-04** — View streak widget on profile/home. **Expected:** Shows current streak count in mono font, streak badge with ring progress indicator, and milestone markers for upcoming milestones.
- [ ] **STREAK-05** — Session summary includes streak count. **Expected:** Streak counter is visible on the session completion screen.

### 5.3 Milestone Bonuses

- [ ] **STREAK-06** — Reach a 7-day streak. **Expected:** Milestone celebration overlay with confetti animation; user receives +1 bonus skip for next session.
- [ ] **STREAK-07** — Reach a 10-day streak. **Expected:** Milestone celebration; user receives +1 bonus skip + token reward.
- [ ] **STREAK-08** — Reach a 14-day streak. **Expected:** Milestone celebration; user receives +1 bonus answer for next session.
- [ ] **STREAK-09** — Start a session after reaching a milestone. **Expected:** `bonus_skips` and/or `bonus_answers` are applied to the new session; resource counters reflect the bonuses (e.g., 11 answers + 5 skips, or 10 answers + 6 skips).
- [ ] **STREAK-10** — Verify milestone thresholds: 3, 7, 10, 14, 21, 30 days. **Expected:** Each milestone triggers the appropriate celebration and bonus.

### 5.4 Streak-at-Risk Notification

- [ ] **STREAK-11** — Have an active streak and do not play by evening. **Expected:** Push notification is sent warning that the streak is at risk (triggered by evening cron).

---

## 6. Rewards

### 6.1 Reward Configuration (Admin)

- [ ] **REWARD-01** — Admin creates a reward config: reward_type = `leaderboard_daily`, rank range 1-3, token amount = 100. **Expected:** Config is saved and appears in active configs list.
- [ ] **REWARD-02** — Admin edits an existing reward config. **Expected:** Changes are saved; future distributions use the updated config.

### 6.2 Reward Distribution

- [ ] **REWARD-03** — Admin triggers reward distribution for a completed daily period. **Expected:** `POST /admin/rewards/distribute` matches leaderboard rankings against reward configs; `reward_distribution` entries are created with status = `pending` for qualifying users.
- [ ] **REWARD-04** — Automatic daily distribution after leaderboard reset (cron). **Expected:** Distribution runs automatically at midnight after the daily leaderboard freezes.
- [ ] **REWARD-05** — Admin manually grants a reward to a specific user via `POST /admin/rewards/grant`. **Expected:** Reward appears in user's pending rewards.

### 6.3 Reward Claiming (User)

- [ ] **REWARD-06** — User navigates to Rewards screen. **Expected:** Pending rewards are listed with token amounts (mono font) and a "Claim" button.
- [ ] **REWARD-07** — User taps "Claim" on a pending reward. **Expected:** Confirmation dialog appears; on confirm, `POST /me/rewards/claim` is called; success animation plays (score count-up, 800ms ease-out); reward status changes to `claimed`.
- [ ] **REWARD-08** — User attempts to claim a reward before the 24h hold period has elapsed. **Expected:** Claim is rejected with a message indicating the hold period remaining.
- [ ] **REWARD-09** — User with an account < 7 days old attempts to claim token rewards. **Expected:** Claim is rejected with a message about minimum account age.
- [ ] **REWARD-10** — View claimed rewards history. **Expected:** Previously claimed rewards are listed with dates and amounts.

### 6.4 Exchange Token Credits

- [ ] **REWARD-11** — Successfully claim a reward. **Expected:** Backend calls Exchange API (or internal queue) to credit tokens to the user's Exchange account; reward status updates to `credited`.
- [ ] **REWARD-12** — Claim a trading fee discount reward. **Expected:** Discount is applied on the Exchange side; user sees confirmation.

### 6.5 WebSocket Notification

- [ ] **REWARD-13** — A reward is earned (e.g., daily leaderboard distribution). **Expected:** WebSocket delivers `reward_earned` event; in-app toast notification appears.

---

## 7. Achievements

### 7.1 Unlock Conditions

- [ ] **ACH-01** — Complete the first prediction session ever. **Expected:** "First Prediction" achievement is unlocked.
- [ ] **ACH-02** — Get all 10 answers correct in a single session (after resolution). **Expected:** "Perfect Day" achievement is unlocked.
- [ ] **ACH-03** — Maintain a 10-day streak. **Expected:** "10-Day Streak" achievement is unlocked.
- [ ] **ACH-04** — Maintain a 30-day streak. **Expected:** "30-Day Streak" achievement is unlocked.
- [ ] **ACH-05** — Finish #1 on a daily leaderboard. **Expected:** "Champion" achievement is unlocked.
- [ ] **ACH-06** — Refer 5 friends who complete their first session. **Expected:** Referral milestone achievement is unlocked (badge + token reward).

### 7.2 Achievement Display

- [ ] **ACH-07** — Navigate to Achievements screen. **Expected:** Grid of badges is displayed; unlocked achievements show in full color with subtle glow; locked achievements show in `textTertiary` color with lock overlay.
- [ ] **ACH-08** — View `GET /me/achievements`. **Expected:** Returns both earned and unearned achievements with unlock status and conditions.

### 7.3 Achievement Notifications

- [ ] **ACH-09** — When an achievement is unlocked, WebSocket delivers `achievement_unlocked` event. **Expected:** Celebration overlay appears in-app with confetti + scale animation.
- [ ] **ACH-10** — When an achievement is unlocked, push notification is sent. **Expected:** Push notification arrives with the achievement name and description (localized).

---

## 8. Social

### 8.1 Referrals

- [ ] **SOC-01** — View referral screen. **Expected:** Displays unique referral code with a copy button and a share link.
- [ ] **SOC-02** — Tap copy button on referral code. **Expected:** Code is copied to clipboard; confirmation feedback is shown.
- [ ] **SOC-03** — Share referral link via system share sheet. **Expected:** Share sheet opens with a pre-formatted message containing the referral link.
- [ ] **SOC-04** — A new user signs up using a referral code. **Expected:** Referral record is created; referrer receives +1 bonus skip.
- [ ] **SOC-05** — Referred user completes their first session. **Expected:** Referral status updates to `first_session_completed`; referrer receives +1 bonus answer.
- [ ] **SOC-06** — Referrer reaches 5 successful referrals. **Expected:** Referrer receives badge + token reward.
- [ ] **SOC-07** — Referrer reaches 10 successful referrals. **Expected:** Referrer receives permanent +1 skip per session.
- [ ] **SOC-08** — View referral stats (`GET /referral/stats`). **Expected:** Shows total referrals, successful completions, and rewards earned.

### 8.2 Mini-Leagues

- [ ] **SOC-09** — Create a mini-league. **Expected:** `POST /leagues` creates a league with a generated invite code; creator is added as a member.
- [ ] **SOC-10** — Share the invite code with a friend; friend joins via `POST /leagues/join`. **Expected:** Friend is added to the league members list.
- [ ] **SOC-11** — View mini-league detail screen. **Expected:** Shows league name, members list, and per-tournament leaderboard within the league.
- [ ] **SOC-12** — View mini-league leaderboard (`GET /leaderboards/league/:leagueId`). **Expected:** Rankings are scoped to league members only.

### 8.3 Social Sharing

- [ ] **SOC-13** — Share session results. **Expected:** Branded image card is generated (dark background, Gold accent) with score and stats; optimized for Instagram Stories / Twitter.
- [ ] **SOC-14** — Share streak milestone. **Expected:** Branded share card with streak count.
- [ ] **SOC-15** — Share leaderboard position. **Expected:** Branded share card with rank and points.
- [ ] **SOC-16** — Share an unlocked achievement badge. **Expected:** Branded share card with badge image.
- [ ] **SOC-17** — Shared content includes a deep link back to the app. **Expected:** Tapping the shared link opens XEX Play (or app store if not installed).

### 8.4 Deep Links

- [ ] **SOC-18** — Open a referral deep link. **Expected:** App opens; if user is new, referral code is applied during signup.
- [ ] **SOC-19** — Open a shared content deep link. **Expected:** App opens and navigates to the relevant screen (leaderboard, profile, etc.).

---

## 9. Admin Panel

### 9.1 Authentication & Access

- [ ] **ADMIN-01** — Navigate to admin panel login page. **Expected:** JWT login form is displayed.
- [ ] **ADMIN-02** — Login with a valid admin JWT. **Expected:** Redirected to dashboard; sidebar navigation is visible (Events, Matches, Cards, Baskets, Users, Leaderboards, Rewards, Notifications, Analytics, Abuse, Exchange Metrics).
- [ ] **ADMIN-03** — Login with a non-admin JWT. **Expected:** Access is denied with an appropriate error message.
- [ ] **ADMIN-04** — Access a dashboard route without authentication. **Expected:** Redirected to login page.

### 9.2 Event CRUD

- [ ] **ADMIN-05** — Create an event with JSONB name (en/fa/ar), slug, start/end dates, scoring multiplier. **Expected:** Event is created and appears in the events list.
- [ ] **ADMIN-06** — Edit an existing event. **Expected:** Changes are saved and reflected in the list.
- [ ] **ADMIN-07** — View the events list. **Expected:** All events are displayed with status (active/inactive), dates, and scoring multiplier.

### 9.3 Match CRUD

- [ ] **ADMIN-08** — Create a match linked to an event with home/away teams, kickoff time. **Expected:** Match is created with status = `upcoming`.
- [ ] **ADMIN-09** — Update a match with final scores and result. **Expected:** Match status changes to `completed`; `home_score`, `away_score`, and `result_data` are saved.
- [ ] **ADMIN-10** — Filter matches by event, date, and status. **Expected:** List is filtered correctly.

### 9.4 Card CRUD & Resolution

- [ ] **ADMIN-11** — Create a card with JSONB question text (en/fa/ar), tier selection (Gold/Silver/White), linked match, `high_answer_is_yes` flag for Gold/Silver cards. **Expected:** Card is created and appears in the cards list.
- [ ] **ADMIN-12** — Resolve a card: select correct answer (Yes/No), confirm via modal. **Expected:** Card `is_resolved` = true; `correct_answer` is set; all `user_answers` for this card are scored (points_earned calculated per tier); leaderboards are updated.
- [ ] **ADMIN-13** — Filter cards by date, tier, and resolved status. **Expected:** List is filtered correctly.

### 9.5 Basket Management

- [ ] **ADMIN-14** — Create a daily basket for a specific date + event. **Expected:** Basket is created in draft status.
- [ ] **ADMIN-15** — Add cards to a basket. **Expected:** Cards are added; tier count is displayed (e.g., "2G / 5S / 6W — need 1 more Gold, 1 more White").
- [ ] **ADMIN-16** — Attempt to publish a basket with incorrect tier composition. **Expected:** Publish is rejected with validation error requiring exactly 3G + 5S + 7W.
- [ ] **ADMIN-17** — Publish a basket with correct tier composition (3G + 5S + 7W). **Expected:** Basket status changes to `published`; basket becomes available to users for that date.

### 9.6 User Management

- [ ] **ADMIN-18** — View users list with search and pagination. **Expected:** Users are listed with display name, email, total points, role, status.
- [ ] **ADMIN-19** — View user detail page. **Expected:** Shows sessions history, answers, stats, referral tree, activity log.
- [ ] **ADMIN-20** — Ban a user. **Expected:** User status changes to banned; user cannot start new sessions or claim rewards.
- [ ] **ADMIN-21** — Unban a user. **Expected:** User status returns to normal; user can resume playing.

### 9.7 Other Admin Pages

- [ ] **ADMIN-22** — View Leaderboard viewer page with CSV export. **Expected:** Leaderboards are displayed; CSV export downloads a file with correct data.
- [ ] **ADMIN-23** — View Analytics dashboard. **Expected:** DAU/WAU/MAU charts, session completion rates, card answer distribution, user retention stats are displayed.
- [ ] **ADMIN-24** — Send a push notification to all users. **Expected:** Compose message (title + body), select target (all / segment), send; delivery stats are shown.
- [ ] **ADMIN-25** — View Translation status page. **Expected:** Shows which cards are missing translations per language; flags baskets that cannot be published due to incomplete translations.
- [ ] **ADMIN-26** — View Anti-abuse dashboard. **Expected:** Flagged accounts and suspicious patterns are listed in a review queue; admin can approve/reject flagged rewards.
- [ ] **ADMIN-27** — View Exchange metrics page. **Expected:** Shows users who navigated to Exchange, conversion rates, and trading activity correlation.
- [ ] **ADMIN-28** — View Prize Pool management page. **Expected:** Can create tournament prize pools (total tokens, distribution percentages) and view active pools.
- [ ] **ADMIN-29** — Verify admin audit log. **Expected:** All admin actions (card resolution, user bans, reward grants, notifications) are logged with admin ID and timestamp.

---

## 10. Push Notifications

### 10.1 Device Registration

- [ ] **PUSH-01** — On first login, app requests push notification permission. **Expected:** If granted, FCM token is obtained and registered with backend via `POST /devices/register`.
- [ ] **PUSH-02** — FCM token refreshes. **Expected:** New token is re-registered with backend automatically.
- [ ] **PUSH-03** — User logs out. **Expected:** FCM token is deactivated via `DELETE /devices/:token`.

### 10.2 Notification Triggers

- [ ] **PUSH-04** — Daily basket is published (morning cron). **Expected:** Push notification is sent to all users: "Today's basket is ready! 15 new predictions await."
- [ ] **PUSH-05** — A card the user answered is resolved (correct). **Expected:** Push notification: "Your prediction was correct! +X points."
- [ ] **PUSH-06** — A card the user answered is resolved (incorrect). **Expected:** Push notification: "Your prediction was incorrect. -X points."
- [ ] **PUSH-07** — User has an active streak and hasn't played by evening. **Expected:** Push notification: "Don't lose your X-day streak! Play today's cards now."
- [ ] **PUSH-08** — User earns a token reward. **Expected:** Push notification with reward details.
- [ ] **PUSH-09** — Achievement is unlocked. **Expected:** Push notification with achievement name and description.

### 10.3 Notification Behavior

- [ ] **PUSH-10** — Receive a push notification while the app is in the foreground. **Expected:** In-app banner is displayed (not a system notification).
- [ ] **PUSH-11** — Receive a push notification while the app is in the background. **Expected:** System notification appears.
- [ ] **PUSH-12** — Tap a push notification. **Expected:** App opens and navigates to the relevant screen (game, rewards, achievements, etc.).
- [ ] **PUSH-13** — Verify all push notification text is localized to the user's language preference. **Expected:** EN/FA/AR users receive notifications in their respective language.

### 10.4 Admin Custom Notifications

- [ ] **PUSH-14** — Admin sends a custom notification to all users. **Expected:** All active devices receive the notification; delivery stats are recorded.
- [ ] **PUSH-15** — Admin sends a notification to a specific segment. **Expected:** Only users in the target segment receive the notification.

---

## 11. Localization

### 11.1 Language Support

- [ ] **L10N-01** — Set app language to English (en). **Expected:** All UI strings, button labels, and navigation items are in English.
- [ ] **L10N-02** — Set app language to Persian (fa). **Expected:** All UI strings are in Persian; layout switches to RTL.
- [ ] **L10N-03** — Set app language to Arabic (ar). **Expected:** All UI strings are in Arabic; layout switches to RTL.
- [ ] **L10N-04** — Change language via `PUT /me` (update user's `language` field). **Expected:** Subsequent API responses return localized content; app UI updates.

### 11.2 RTL Layout

- [ ] **L10N-05** — In FA or AR mode, verify the bottom navigation bar is mirrored. **Expected:** Tab order is reversed (right-to-left reading order).
- [ ] **L10N-06** — In FA or AR mode, verify card swipe directions are correct. **Expected:** Swipe gestures remain: right = Yes, left = No, up = Skip (gestures are NOT mirrored).
- [ ] **L10N-07** — In FA or AR mode, verify text alignment is right-aligned where appropriate. **Expected:** Body text, labels, and descriptions are right-aligned.
- [ ] **L10N-08** — In FA or AR mode, verify the leaderboard layout is mirrored. **Expected:** Rank on right, points on left.
- [ ] **L10N-09** — In FA or AR mode, verify the admin panel displays correctly. **Expected:** Admin panel renders properly (admin is LTR by default if not localized, or renders correctly if localized).

### 11.3 Localized Content

- [ ] **L10N-10** — View a card question in EN. **Expected:** Question text is in English (from JSONB `question_text.en`).
- [ ] **L10N-11** — View the same card question in FA. **Expected:** Question text is in Persian (from JSONB `question_text.fa`).
- [ ] **L10N-12** — View a card where FA translation is missing. **Expected:** Falls back to EN text gracefully (no empty/broken UI).
- [ ] **L10N-13** — View event and match names in each language. **Expected:** JSONB `name` fields are resolved to the user's language.
- [ ] **L10N-14** — Achievement names and descriptions are localized. **Expected:** Correct language is displayed per user preference.

### 11.4 Date & Number Formatting

- [ ] **L10N-15** — In EN mode, verify dates are in Gregorian format (e.g., "March 10, 2026"). **Expected:** Correct date format.
- [ ] **L10N-16** — In FA mode, verify dates are displayed appropriately for the locale. **Expected:** Persian date format or Gregorian with Persian numerals, depending on implementation.
- [ ] **L10N-17** — In FA/AR mode, verify numbers use the correct numeral system where appropriate. **Expected:** Numbers display correctly per locale conventions.
- [ ] **L10N-18** — Verify `Accept-Language` header is sent with API requests. **Expected:** Backend locale middleware parses the header and uses it for response localization.

---

## 12. Performance

### 12.1 API Response Times

- [ ] **PERF-01** — `POST /sessions/start` responds in < 200ms (p95). **Expected:** Measured under normal load.
- [ ] **PERF-02** — `POST /sessions/current/answer` responds in < 200ms (p95). **Expected:** Measured under normal load.
- [ ] **PERF-03** — `GET /leaderboards/daily` responds in < 200ms (p95) with 10K+ users. **Expected:** Redis sorted set serves the data.
- [ ] **PERF-04** — `POST /admin/cards/:id/resolve` resolves and scores 10K+ user_answers in < 5 seconds. **Expected:** Batch processing with DB transaction.

### 12.2 WebSocket

- [ ] **PERF-05** — WebSocket connection is established within 2 seconds. **Expected:** JWT auth via query param, connection upgrade succeeds.
- [ ] **PERF-06** — WebSocket maintains connection with ping/pong. **Expected:** Connection stays alive during inactivity; no unexpected disconnects.
- [ ] **PERF-07** — WebSocket auto-reconnects after connection drop. **Expected:** Client detects disconnect and reconnects with exponential backoff.
- [ ] **PERF-08** — WebSocket broadcasts `card_resolved` event to affected users within 1 second of card resolution. **Expected:** Minimal latency between resolution and notification.

### 12.3 App Animation Smoothness

- [ ] **PERF-09** — Card swipe animation runs at 60fps. **Expected:** No frame drops during swipe, fly-off, and next-card scale-up.
- [ ] **PERF-10** — Timer countdown animation is smooth. **Expected:** No jank in the circular progress ring depletion.
- [ ] **PERF-11** — Achievement celebration overlay (confetti + scale) runs without lag. **Expected:** 60fps during the animation.
- [ ] **PERF-12** — Score count-up animation on session summary is smooth. **Expected:** 60fps, 800ms duration, ease-out curve.

### 12.4 Caching

- [ ] **PERF-13** — Redis cache hit rate > 90% for leaderboards, active sessions, and baskets. **Expected:** Verified via Prometheus `cache_hit` / `cache_miss` counters.
- [ ] **PERF-14** — Published basket data is served from Redis cache (not DB) on subsequent requests. **Expected:** Cache is populated on first request; subsequent requests hit cache.

---

## 13. Security

### 13.1 JWT Validation

- [ ] **SEC-01** — Backend validates JWT signature using shared `JWT_SECRET` (HS256). **Expected:** Only tokens signed with the correct secret are accepted.
- [ ] **SEC-02** — Backend rejects JWTs with `iss` claim other than `"nyyu"`. **Expected:** 401 Unauthorized.
- [ ] **SEC-03** — Backend rejects expired JWTs (exp < current time). **Expected:** 401 Unauthorized.
- [ ] **SEC-04** — Backend extracts `user_id`, `email`, `role` from validated JWT and injects into request context. **Expected:** Downstream handlers can access user identity.

### 13.2 Rate Limiting

- [ ] **SEC-05** — Exceed the rate limit for `/sessions/current/answer` (e.g., rapid-fire submissions). **Expected:** Backend returns 429 Too Many Requests after the limit is reached.
- [ ] **SEC-06** — Exceed the rate limit for `/auth/login`. **Expected:** Backend returns 429 Too Many Requests.
- [ ] **SEC-07** — Rate limiting is per-user (not global). **Expected:** One user hitting the limit does not affect other users.
- [ ] **SEC-08** — Rate limit state is stored in Redis. **Expected:** Limits persist across API server restarts and work correctly with multiple API replicas.

### 13.3 CORS

- [ ] **SEC-09** — Make a request to the API from the admin panel origin. **Expected:** Request succeeds; CORS headers are present.
- [ ] **SEC-10** — Make a request to the API from an unauthorized origin. **Expected:** CORS preflight fails; browser blocks the request.

### 13.4 Input Sanitization

- [ ] **SEC-11** — Submit a card answer with an invalid value (not boolean). **Expected:** Backend returns 400 Bad Request with validation error.
- [ ] **SEC-12** — Submit a request with SQL injection attempt in a string field (e.g., display name). **Expected:** No SQL injection occurs; parameterized queries via pgx prevent it; input is sanitized.
- [ ] **SEC-13** — Submit card question text with HTML/script tags via admin. **Expected:** Text is sanitized; no XSS when displayed in the app.
- [ ] **SEC-14** — Attempt to access another user's session via IDOR (e.g., `GET /sessions/:otherId`). **Expected:** Backend checks user ownership; returns 403 Forbidden.

### 13.5 Anti-Abuse

- [ ] **SEC-15** — Create multiple accounts from the same device fingerprint. **Expected:** Accounts are flagged for multi-account abuse; flagged in admin review queue.
- [ ] **SEC-16** — New account (< 7 days old) achieves a perfect score. **Expected:** Account is flagged for anomaly detection (perfect score from new account).
- [ ] **SEC-17** — User exceeds daily/weekly token reward velocity limits. **Expected:** Additional rewards are blocked; flagged for admin review.
- [ ] **SEC-18** — Device fingerprint and IP are recorded on login. **Expected:** `UpdateDeviceInfo` stores device_id and IP for abuse detection.

### 13.6 Data Isolation

- [ ] **SEC-19** — Verify XEX Play has zero network access to Exchange database. **Expected:** Separate resource groups, servers, and credentials; connection attempt from Play to Exchange DB fails.
- [ ] **SEC-20** — Verify XEX Play has its own Redis instance separate from Exchange. **Expected:** Separate Redis endpoints and credentials.

### 13.7 Infrastructure Security

- [ ] **SEC-21** — Docker containers run as non-root user. **Expected:** Verified in both API and Admin Dockerfiles.
- [ ] **SEC-22** — No secrets are baked into Docker image layers. **Expected:** Secrets are injected via environment variables at runtime.
- [ ] **SEC-23** — GitHub Dependabot is configured for Go, Flutter, and Next.js dependencies. **Expected:** `.github/dependabot.yml` exists and scans are active.
- [ ] **SEC-24** — Health check endpoints (`/health`, `/health/ready`) do not expose sensitive information. **Expected:** Only return status (ok/not ok) and component connectivity status.

---

## 14. Exchange Integration

### 14.1 Token Claims

- [ ] **EX-01** — Claim a pending reward. **Expected:** Backend calls Exchange API to credit tokens; reward status updates to `credited`.
- [ ] **EX-02** — Claim is rejected if the linked Exchange account is not in good standing. **Expected:** `verifyExchangeAccount` check fails; claim is rejected with an appropriate message.

### 14.2 VIP / Trader Benefits

- [ ] **EX-03** — Active Exchange trader logs in. **Expected:** `trading_tier` field is set; VIP card tier is unlocked; trader badge is displayed.
- [ ] **EX-04** — Non-trader user does not see VIP cards. **Expected:** VIP tier cards are not included in their basket.

### 14.3 Exchange Prompts

- [ ] **EX-05** — After completing a session, contextual Exchange prompt is displayed. **Expected:** Banner/card encouraging trading on XEX Exchange is shown at the appropriate moment.
- [ ] **EX-06** — On the rewards screen, Exchange prompt is displayed. **Expected:** Contextual prompt appears.
- [ ] **EX-07** — Tap "Trade on XEX Exchange" button. **Expected:** Opens Exchange app via deep link; if Exchange app is not installed, falls back to web URL.
- [ ] **EX-08** — `GET /me/exchange-prompts` returns contextual prompts. **Expected:** Prompts are returned based on user context (post-session, reward screen, achievement, leaderboard).

---

## Appendix: Test Environment Setup

Before running the regression:

1. **Staging environment** is deployed with the latest code from all components (backend, app, admin).
2. **Seed data** is loaded: at least 1 event, 4 matches, 15+ cards (3G + 5S + 7W), 1 published basket.
3. **Test accounts** are prepared: 1 admin account, 2+ regular user accounts, 1 new account (< 7 days old).
4. **FCM** is configured for the test devices (iOS + Android).
5. **WebSocket** endpoint is accessible from test devices.
6. **Redis** and **PostgreSQL** are running and healthy (verify via `/health/ready`).

---

_Last updated: 2026-03-10_
