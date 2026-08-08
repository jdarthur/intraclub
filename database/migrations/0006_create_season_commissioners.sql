-- 0006_create_season_commissioners.sql
-- The season_commissioner join table, matching the SeasonCommissioner record
-- shape (model/season_commissioner.go).
-- Table name equals record.Type() ("season_commissioner").
--   id         -> RecordId hex TEXT primary key
--   season_id  -> SeasonId hex TEXT
--   user_id    -> UserId hex TEXT
--   created_at -> RFC3339 TEXT
--   updated_at -> RFC3339 TEXT
CREATE TABLE season_commissioner (
    id         TEXT PRIMARY KEY,   -- RecordId hex string
    season_id  TEXT NOT NULL,      -- SeasonId hex string
    user_id    TEXT NOT NULL,      -- UserId hex string
    created_at TEXT NOT NULL,      -- RFC3339
    updated_at TEXT NOT NULL       -- RFC3339
);
