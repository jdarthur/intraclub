-- 0009_create_teams.sql
-- The team table, matching the Team record shape (model/team.go).
-- Table name equals record.Type() ("team").
--   id           -> TeamId hex TEXT primary key
--   name         -> TEXT
--   color_name   -> TeamColor.Name (nested value-struct flattened; model/colors.go)
--   color_hex    -> TeamColor.Hex
--   created_at   -> RFC3339 TEXT
--   updated_at   -> RFC3339 TEXT
--   deleted_at   -> RFC3339 TEXT, nullable (*time.Time)
-- Team.RatingsMap was normalized into the team_rating join table (0011); the
-- inline map is not part of this table.
CREATE TABLE team (
    id          TEXT PRIMARY KEY,   -- TeamId hex string
    name        TEXT NOT NULL,
    color_name  TEXT NOT NULL,      -- TeamColor.Name
    color_hex   TEXT NOT NULL,      -- TeamColor.Hex
    created_at  TEXT NOT NULL,      -- RFC3339
    updated_at  TEXT NOT NULL,      -- RFC3339
    deleted_at  TEXT                -- RFC3339, nullable
);
