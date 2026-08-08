-- 0018_create_draft_available_players.sql
-- The draft_available_player join table, matching the DraftAvailablePlayer
-- record shape (model/draft_available_player.go).
-- Table name equals record.Type() ("draft_available_player").
--   id         -> RecordId hex TEXT primary key
--   draft_id   -> DraftId hex TEXT
--   player_id  -> UserId hex TEXT
--   created_at -> RFC3339 TEXT
--   updated_at -> RFC3339 TEXT
CREATE TABLE draft_available_player (
    id         TEXT PRIMARY KEY,   -- RecordId hex string
    draft_id   TEXT NOT NULL,      -- DraftId hex string
    player_id  TEXT NOT NULL,      -- UserId hex string
    created_at TEXT NOT NULL,      -- RFC3339
    updated_at TEXT NOT NULL       -- RFC3339
);
