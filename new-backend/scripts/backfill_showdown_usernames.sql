-- One-time migration aid: showdown_username became nullable (users now set it
-- via onboarding after login). Existing accounts created by Discord OAuth got
-- an empty string, which would otherwise block new signups via the unique index.
--
-- Run once against the dev database:
--   psql "$DATABASE_URL" -f scripts/backfill_showdown_usernames.sql
UPDATE users SET showdown_username = NULL WHERE showdown_username = '';
