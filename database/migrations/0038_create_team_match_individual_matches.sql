-- 0038_create_team_match_individual_matches.sql
-- The team_match_individual_match join table, matching the
-- TeamMatchIndividualMatch record shape (model/team_match.go). This is the
-- normalized replacement for the former inline
-- TeamMatch.IndividualMatches map (lineup pairing -> individual match).
-- Table name equals record.Type() ("team_match_individual_match").
--   id                   -> RecordId hex TEXT primary key
--   team_match_id        -> TeamMatchId hex TEXT
--   lineup_pairing_id    -> LineupPairingId hex TEXT
--   individual_match_id  -> IndividualMatchId hex TEXT
-- One match per (team match, lineup pairing): UNIQUE(team_match_id,
-- lineup_pairing_id) mirrors TeamMatchIndividualMatch.UniquenessEquivalent.
CREATE TABLE team_match_individual_match (
    id                   TEXT PRIMARY KEY,   -- RecordId hex string
    team_match_id        TEXT NOT NULL,      -- TeamMatchId hex string
    lineup_pairing_id    TEXT NOT NULL,      -- LineupPairingId hex string
    individual_match_id  TEXT NOT NULL,      -- IndividualMatchId hex string
    UNIQUE (team_match_id, lineup_pairing_id)
);
