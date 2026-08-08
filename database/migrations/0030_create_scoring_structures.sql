-- 0030_create_scoring_structures.sql
-- The scoring_structure table, matching the ScoringStructure record shape
-- (model/scoring_structure.go). Table name equals record.Type() ("scoring_structure").
--   id                              -> ScoringStructureId hex TEXT primary key
--   owner                           -> UserId hex TEXT
--   name                            -> TEXT
--   win_condition_counting_type     -> INTEGER (ScoreCountingType enum)
--   win_condition_win_threshold     -> INTEGER (nested WinCondition flattened)
--   win_condition_must_win_by       -> INTEGER
--   win_condition_instant_win_threshold -> INTEGER
--
-- ScoringStructure.SecondaryScoringStructures was normalized into the
-- scoring_structure_secondary join table (0031); the inline slice is not part
-- of this table.
CREATE TABLE scoring_structure (
    id                                 TEXT PRIMARY KEY,   -- ScoringStructureId hex string
    owner                              TEXT NOT NULL,      -- UserId hex string
    name                               TEXT NOT NULL,
    win_condition_counting_type        INTEGER NOT NULL,  -- ScoreCountingType enum
    win_condition_win_threshold        INTEGER NOT NULL,  -- WinCondition.WinThreshold
    win_condition_must_win_by          INTEGER NOT NULL,  -- WinCondition.MustWinBy
    win_condition_instant_win_threshold INTEGER NOT NULL  -- WinCondition.InstantWinThreshold
);
