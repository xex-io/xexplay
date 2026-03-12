DROP INDEX IF EXISTS idx_events_sport_key;
DROP INDEX IF EXISTS idx_cards_source;
DROP INDEX IF EXISTS idx_matches_source;
DROP INDEX IF EXISTS idx_matches_sport_key;
DROP INDEX IF EXISTS idx_matches_external_id;
DROP INDEX IF EXISTS idx_automation_logs_created_at;
DROP INDEX IF EXISTS idx_automation_logs_job_name;
DROP TABLE IF EXISTS automation_logs;

ALTER TABLE events DROP COLUMN IF EXISTS sport_key;
ALTER TABLE events DROP COLUMN IF EXISTS source;
ALTER TABLE cards DROP COLUMN IF EXISTS resolution_criteria;
ALTER TABLE cards DROP COLUMN IF EXISTS ai_prompt_data;
ALTER TABLE cards DROP COLUMN IF EXISTS source;
ALTER TABLE matches DROP COLUMN IF EXISTS source;
ALTER TABLE matches DROP COLUMN IF EXISTS sport_key;
ALTER TABLE matches DROP COLUMN IF EXISTS external_id;

DROP TABLE IF EXISTS sports;
