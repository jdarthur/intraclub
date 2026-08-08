-- 0008_create_season_teams.sql
-- The season_team join table, matching the SeasonTeam record shape
-- (model/season_team.go).
-- Table name equals record.Type() ("season_team").
--   id         -> RecordId hex TEXT primary key
--   season_id  -> SeasonId hex TEXT
--   team_id    -> TeamId hex TEXT
--   created_at -> RFC3339 TEXT
--   updated_at -> RFC3339 TEXT
CREATE TABLE season_team (
    id         TEXT PRIMARY KEY,   -- RecordId hex string
    season_id  TEXT NOT NULL,      -- SeasonId hex string
    team_id    TEXT NOT NULL,      -- TeamId hex string
    created_at TEXT NOT NULL,      -- RFC3339
    updated_at TEXT NOT NULL       -- RFC3339
);
