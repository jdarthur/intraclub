-- 0037_create_team_matches.sql
-- The team_match table, matching the TeamMatch record shape
-- (model/team_match.go).
-- Table name equals record.Type() ("team_match").
--   id         -> TeamMatchId hex TEXT primary key
--   week_id    -> WeekId hex TEXT
--   home_team  -> TeamId hex TEXT
--   away_team  -> TeamId hex TEXT
--   lineup     -> LineupId hex TEXT
CREATE TABLE team_match (
    id        TEXT PRIMARY KEY,   -- TeamMatchId hex string
    week_id   TEXT NOT NULL,      -- WeekId hex string
    home_team TEXT NOT NULL,      -- TeamId hex string
    away_team TEXT NOT NULL,      -- TeamId hex string
    lineup    TEXT NOT NULL       -- LineupId hex string
);
