-- 0039_create_lineups.sql
-- The lineup table, matching the Lineup record shape (model/lineup.go).
-- Table name equals record.Type() ("lineup").
--   id       -> LineupId hex TEXT primary key
--   team_id  -> TeamId hex TEXT
--   week_id  -> WeekId hex TEXT
-- One lineup per (team, week): UNIQUE(week_id, team_id) mirrors
-- Lineup.UniquenessEquivalent.
CREATE TABLE lineup (
    id      TEXT PRIMARY KEY,   -- LineupId hex string
    team_id TEXT NOT NULL,      -- TeamId hex string
    week_id TEXT NOT NULL,      -- WeekId hex string
    UNIQUE (week_id, team_id)
);
