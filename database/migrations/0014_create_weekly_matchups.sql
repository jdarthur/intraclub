-- 0014_create_weekly_matchups.sql
-- The weekly_matchup table, matching the WeeklyMatchup record shape
-- (model/weekly_matchup.go).
-- Table name equals record.Type() ("weekly_matchup").
--   id        -> WeeklyMatchupId hex TEXT primary key
--   week_id   -> WeekId hex TEXT
--   season_id -> SeasonId hex TEXT
-- One weekly matchup per (season, week): UNIQUE(season_id, week_id) mirrors
-- WeeklyMatchup.UniquenessEquivalent.
-- WeeklyMatchup.Matchups ([]*TeamMatchup) is normalized into the
-- weekly_matchup_team_matchup join table (0015), not stored inline.
CREATE TABLE weekly_matchup (
    id        TEXT PRIMARY KEY,   -- WeeklyMatchupId hex string
    week_id   TEXT NOT NULL,      -- WeekId hex string
    season_id TEXT NOT NULL,      -- SeasonId hex string
    UNIQUE (season_id, week_id)
);
