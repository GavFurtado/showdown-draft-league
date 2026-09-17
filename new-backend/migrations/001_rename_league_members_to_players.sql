-- Migration: Rename league_members → players
-- Handles the case where GORM already created an empty `players` table.
-- Safe to run multiple times (idempotent).

BEGIN;

-- 1. Drop every FK constraint on any table that references league_members.
DO $$
DECLARE
    rec RECORD;
BEGIN
    FOR rec IN
        SELECT tc.constraint_name, tc.table_name
        FROM information_schema.table_constraints tc
        JOIN information_schema.key_column_usage kcu
          ON tc.constraint_name = kcu.constraint_name
          AND tc.table_catalog = kcu.table_catalog
        JOIN information_schema.constraint_column_usage ccu
          ON ccu.constraint_name = tc.constraint_name
          AND ccu.table_catalog = tc.table_catalog
        WHERE tc.constraint_type = 'FOREIGN KEY'
          AND ccu.table_name = 'league_members'
          AND tc.table_name != 'league_members'
    LOOP
        EXECUTE format('ALTER TABLE %I DROP CONSTRAINT %I', rec.table_name, rec.constraint_name);
    END LOOP;
END
$$;

-- 2. Drop any FK constraints that league_members itself has.
DO $$
DECLARE
    rec RECORD;
BEGIN
    FOR rec IN
        SELECT constraint_name
        FROM information_schema.table_constraints
        WHERE table_name = 'league_members'
          AND constraint_type = 'FOREIGN KEY'
    LOOP
        EXECUTE format('ALTER TABLE league_members DROP CONSTRAINT %I', rec.constraint_name);
    END LOOP;
END
$$;

-- 3. Drop FK constraints on other tables that GORM created pointing at the empty players table.
DO $$
DECLARE
    rec RECORD;
BEGIN
    FOR rec IN
        SELECT tc.constraint_name, tc.table_name
        FROM information_schema.table_constraints tc
        JOIN information_schema.key_column_usage kcu
          ON tc.constraint_name = kcu.constraint_name
          AND tc.table_catalog = kcu.table_catalog
        JOIN information_schema.constraint_column_usage ccu
          ON ccu.constraint_name = tc.constraint_name
          AND ccu.table_catalog = tc.table_catalog
        WHERE tc.constraint_type = 'FOREIGN KEY'
          AND ccu.table_name = 'players'
          AND tc.table_name != 'players'
    LOOP
        EXECUTE format('ALTER TABLE %I DROP CONSTRAINT %I', rec.table_name, rec.constraint_name);
    END LOOP;
END
$$;

-- 4. Drop the empty players table GORM already created.
DROP TABLE IF EXISTS players;

-- 5. Rename league_members → players.
ALTER TABLE league_members RENAME TO players;

COMMIT;
