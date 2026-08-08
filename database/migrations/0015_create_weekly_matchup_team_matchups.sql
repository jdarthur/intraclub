-- 0015_create_weekly_matchup_team_matchups.sql
-- The weekly_matchup_team_matchup join table, matching the
-- WeeklyMatchupTeamMatchup record shape (model/weekly_matchup_team_matchup.go).
-- This is the normalized replacement for the former inline
-- WeeklyMatchup.Matchups []*TeamMatchup slice, preserving ordering via the
-- Position column.
-- Table name equals record.Type() ("weekly_matchup_team_matchup").
--   id                 -> WeeklyMatchupTeamMatchupId hex TEXT primary key
--   weekly_matchup_id  -> WeeklyMatchupId hex TEXT
--   home_team_id       -> TeamId hex TEXT
--   away_team_id       -> TeamId hex TEXT
--   bye                -> bool as INTEGER (0/1)
--   position           -> INTEGER
CREATE TABLE weekly_matchup_team_matchup (
    id                TEXT PRIMARY KEY,   -- WeeklyMatchupTeamMatchupId hex string
    weekly_matchup_id TEXT NOT NULL,      -- WeeklyMatchupId hex string
    home_team_id      TEXT NOT NULL,      -- TeamId hex string
    away_team_id      TEXT NOT NULL,      -- TeamId hex string
    bye               INTEGER NOT NULL,   -- bool
    position          INTEGER NOT NULL
);
