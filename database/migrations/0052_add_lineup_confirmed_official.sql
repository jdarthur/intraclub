-- 0052_add_lineup_confirmed_official.sql
-- Add the captain-confirmed and commissioner-official flags to the lineup
-- table. The Lineup model (model/lineup.go) tracks these so a weekly lineup
-- must be confirmed by the team captain/co-captains before the season
-- commissioner marks it official. Stored as INTEGER (0/1), matching the
-- provider's bool mapping.
ALTER TABLE lineup ADD COLUMN confirmed INTEGER NOT NULL DEFAULT 0;
ALTER TABLE lineup ADD COLUMN official INTEGER NOT NULL DEFAULT 0;
