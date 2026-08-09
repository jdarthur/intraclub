-- 0035_create_matches.sql
-- The match table, matching the IndividualMatch record shape
-- (model/individual_match.go).
-- Table name equals record.Type() ("match").
--   id              -> IndividualMatchId hex TEXT primary key
--   opponent        -> IndividualMatchId hex TEXT
--   structure       -> ScoringStructureId hex TEXT
--   main_value      -> INTEGER
--   secondary_value -> INTEGER
--   win_override    -> INTEGER (bool: 0/1)
--   status          -> INTEGER (IndividualMatchStatus)
-- The former inline IndividualMatch.Editors slice was normalized into the
-- match_editor child table (0036); it is not part of this table.
CREATE TABLE match (
    id              TEXT PRIMARY KEY,   -- IndividualMatchId hex string
    opponent        TEXT NOT NULL,      -- IndividualMatchId hex string
    structure       TEXT NOT NULL,      -- ScoringStructureId hex string
    main_value      INTEGER NOT NULL,
    secondary_value INTEGER NOT NULL,
    win_override    INTEGER NOT NULL,
    status          INTEGER NOT NULL
);
