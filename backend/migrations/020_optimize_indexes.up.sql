-- Migration 020: Add missing indexes for query optimization.
-- All indexes use IF NOT EXISTS for idempotency.
-- CONCURRENTLY cannot be used inside a transaction block, so we use regular
-- CREATE INDEX here (migrations typically run in a transaction).

-- =============================================================================
-- EVENTS
-- =============================================================================

-- FindActive() queries: WHERE is_active = true ORDER BY start_date ASC.
-- Partial index covers only the active rows, which are a small subset.
CREATE INDEX IF NOT EXISTS idx_events_active_start_date
    ON events (start_date ASC)
    WHERE is_active = true;

-- =============================================================================
-- CARDS
-- =============================================================================

-- FindUnresolvedByMatch() queries: WHERE match_id = $1 AND is_resolved = false.
-- Composite partial index avoids scanning resolved cards entirely.
CREATE INDEX IF NOT EXISTS idx_cards_match_unresolved
    ON cards (match_id)
    WHERE is_resolved = false;

-- =============================================================================
-- DAILY_BASKET_CARDS
-- =============================================================================

-- GetCardsForBasket() and Publish() JOIN on card_id. The existing UNIQUE
-- constraints on (basket_id, card_id) and (basket_id, position) help lookups
-- by basket_id, but a standalone index on card_id speeds up the reverse lookup
-- (e.g., finding which baskets contain a given card).
CREATE INDEX IF NOT EXISTS idx_daily_basket_cards_card_id
    ON daily_basket_cards (card_id);

-- =============================================================================
-- USER_ANSWERS
-- =============================================================================

-- FindBySession() queries: WHERE session_id = $1 ORDER BY answered_at ASC.
-- session_id has no standalone index despite the UNIQUE(session_id, card_id).
-- The unique constraint helps, but an index with answered_at included lets
-- the planner satisfy the ORDER BY directly.
CREATE INDEX IF NOT EXISTS idx_user_answers_session_id
    ON user_answers (session_id, answered_at);

-- BulkResolve() queries: WHERE card_id = $1 AND is_correct IS NULL.
-- Partial index targets only unresolved answers, which shrinks over time.
CREATE INDEX IF NOT EXISTS idx_user_answers_card_unresolved
    ON user_answers (card_id)
    WHERE is_correct IS NULL;

-- =============================================================================
-- LEADERBOARD_ENTRIES
-- =============================================================================

-- GetRanking()/GetUserRank() JOINs users on le.user_id, but the existing
-- idx_leaderboard_user already covers that. However, event_id is a FK with
-- no index, needed for potential event-scoped leaderboard queries.
CREATE INDEX IF NOT EXISTS idx_leaderboard_event_id
    ON leaderboard_entries (event_id)
    WHERE event_id IS NOT NULL;

-- =============================================================================
-- STREAKS
-- =============================================================================

-- FindStreaksAtRisk() queries: WHERE current_streak > 0 AND last_played_date = yesterday.
-- Partial index on active streaks lets the scheduler find at-risk users efficiently.
CREATE INDEX IF NOT EXISTS idx_streaks_at_risk
    ON streaks (last_played_date)
    WHERE current_streak > 0;

-- =============================================================================
-- REWARD_CONFIGS
-- =============================================================================

-- FindActiveConfigs() queries: WHERE is_active = true AND period_type = $1 ORDER BY rank_from.
-- No indexes existed on this table at all.
CREATE INDEX IF NOT EXISTS idx_reward_configs_active_period
    ON reward_configs (period_type, rank_from)
    WHERE is_active = true;

-- =============================================================================
-- REWARD_DISTRIBUTIONS
-- =============================================================================

-- FindPendingByUser() queries: WHERE user_id = $1 AND status = 'pending'.
-- The existing idx_reward_dist_status is (user_id, status) which covers this,
-- but a partial index on pending-only rows is narrower and faster.
CREATE INDEX IF NOT EXISTS idx_reward_dist_pending
    ON reward_distributions (user_id, created_at DESC)
    WHERE status = 'pending';

-- ClaimReward() queries: WHERE id = $1 AND user_id = $2 AND status = 'pending'.
-- CreditReward() queries: WHERE id = $1 AND user_id = $2 AND status = 'claimed'.
-- PK on id already covers the lookup, but adding status as a partial index
-- avoids rechecking rows. However, these are single-row lookups by PK, so
-- the benefit is marginal. Skipping to avoid index bloat.

-- =============================================================================
-- AUDIT_LOGS
-- =============================================================================

-- FindByEntity() queries: WHERE entity_type = $1 AND entity_id = $2 ORDER BY created_at DESC.
-- No composite index exists for this query pattern.
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity
    ON audit_logs (entity_type, entity_id, created_at DESC);

-- =============================================================================
-- ABUSE_FLAGS
-- =============================================================================

-- CountByUserAndType() queries: WHERE user_id = $1 AND flag_type = $2.
-- The existing idx_abuse_flags_user_id covers user_id alone, but adding
-- flag_type avoids heap fetches for the type filter.
CREATE INDEX IF NOT EXISTS idx_abuse_flags_user_type
    ON abuse_flags (user_id, flag_type);

-- FindPending() queries: WHERE status = 'pending' ORDER BY created_at DESC.
-- Partial index targets only the pending subset for admin review queue.
CREATE INDEX IF NOT EXISTS idx_abuse_flags_pending
    ON abuse_flags (created_at DESC)
    WHERE status = 'pending';

-- =============================================================================
-- REFERRALS
-- =============================================================================

-- CountByReferrer() uses: WHERE referrer_id = $1 with FILTER (WHERE status = 'first_session').
-- Composite index lets Postgres do an index-only scan for the count + filter.
CREATE INDEX IF NOT EXISTS idx_referrals_referrer_status
    ON referrals (referrer_id, status);

-- =============================================================================
-- USERS
-- =============================================================================

-- GetStats() and other admin queries may filter by is_active.
-- Partial index covers only active users (the majority, but useful for
-- admin queries that filter inactive users).
CREATE INDEX IF NOT EXISTS idx_users_active
    ON users (id)
    WHERE is_active = true;
