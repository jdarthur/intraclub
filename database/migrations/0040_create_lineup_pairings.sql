-- 0040_create_lineup_pairings.sql
-- The lineup_pairing table, matching the LineupPairing record shape
-- (model/lineup_pairing.go).
-- Table name equals record.Type() ("lineup_pairing").
--   id                -> LineupPairingId hex TEXT primary key
--   lineup_id         -> LineupId hex TEXT
--   team_id           -> TeamId hex TEXT
--   player1           -> UserId hex TEXT
--   player2           -> UserId hex TEXT
--   format_line_index -> INTEGER
CREATE TABLE lineup_pairing (
    id                TEXT PRIMARY KEY,   -- LineupPairingId hex string
    lineup_id         TEXT NOT NULL,      -- LineupId hex string
    team_id           TEXT NOT NULL,      -- TeamId hex string
    player1           TEXT NOT NULL,      -- UserId hex string
    player2           TEXT NOT NULL,      -- UserId hex string
    format_line_index INTEGER NOT NULL
);
