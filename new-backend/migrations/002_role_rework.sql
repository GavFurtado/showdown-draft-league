-- Migration: Roles rework
-- Renames base role from 'MEMBER' to 'REGULAR' and adds derived spectator roles.
-- Spectator roles (SPECTATOR / SPECTATOR_MODERATOR / SPECTATOR_OWNER) are
-- introduced with the new spectators table (GORM AutoMigrate) — no backfill needed.
-- Safe to run multiple times (idempotent).

BEGIN;

-- Base standing renamed MEMBER -> REGULAR on existing players.
UPDATE players SET role = 'REGULAR' WHERE role = 'MEMBER';
UPDATE spectators SET role = 'SPECTATOR' WHERE role = 'MEMBER';

COMMIT;