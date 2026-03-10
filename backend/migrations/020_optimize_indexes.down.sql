-- Rollback migration 020: Remove optimization indexes.

DROP INDEX IF EXISTS idx_events_active_start_date;
DROP INDEX IF EXISTS idx_cards_match_unresolved;
DROP INDEX IF EXISTS idx_daily_basket_cards_card_id;
DROP INDEX IF EXISTS idx_user_answers_session_id;
DROP INDEX IF EXISTS idx_user_answers_card_unresolved;
DROP INDEX IF EXISTS idx_leaderboard_event_id;
DROP INDEX IF EXISTS idx_streaks_at_risk;
DROP INDEX IF EXISTS idx_reward_configs_active_period;
DROP INDEX IF EXISTS idx_reward_dist_pending;
DROP INDEX IF EXISTS idx_audit_logs_entity;
DROP INDEX IF EXISTS idx_abuse_flags_user_type;
DROP INDEX IF EXISTS idx_abuse_flags_pending;
DROP INDEX IF EXISTS idx_referrals_referrer_status;
DROP INDEX IF EXISTS idx_users_active;
