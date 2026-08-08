-- 0021_create_draft_picks.sql
-- The draft_pick join table, matching the DraftPick record shape
-- (model/draft_pick.go).
-- Table name equals record.Type() ("draft_pick").
--   id         -> RecordId hex TEXT primary key
--   draft_id   -> DraftId hex TEXT
--   team_id    -> TeamId hex TEXT
--   user_id    -> UserId hex TEXT
--   round      -> INTEGER
--   pick       -> INTEGER
--   rating     -> RatingId hex TEXT
--   created_at -> RFC3339 TEXT
--   updated_at -> RFC3339 TEXT
CREATE TABLE draft_pick (
    id         TEXT PRIMARY KEY,   -- RecordId hex string
    draft_id   TEXT NOT NULL,      -- DraftId hex string
    team_id    TEXT NOT NULL,      -- TeamId hex string
    user_id    TEXT NOT NULL,      -- UserId hex string
    round      INTEGER NOT NULL,
    pick       INTEGER NOT NULL,
    rating     TEXT NOT NULL,      -- RatingId hex string
    created_at TEXT NOT NULL,      -- RFC3339
    updated_at TEXT NOT NULL       -- RFC3339
);
