-- 0031_create_scoring_structure_secondaries.sql
-- The scoring_structure_secondary join table, matching the
-- ScoringStructureSecondary record shape (model/scoring_structure_secondary.go).
-- This is the normalized replacement for the former inline
-- ScoringStructure.SecondaryScoringStructures slice, preserving order via
-- SecondaryIndex. Table name equals record.Type() ("scoring_structure_secondary").
--   id                               -> RecordId hex TEXT primary key
--   scoring_structure_id             -> ScoringStructureId hex TEXT
--   secondary_scoring_structure_id   -> ScoringStructureId hex TEXT
--   secondary_index                  -> INTEGER (position within the secondary list)
-- One secondary per (structure, index): UNIQUE(scoring_structure_id,
-- secondary_index) mirrors ScoringStructureSecondary.UniquenessEquivalent.
CREATE TABLE scoring_structure_secondary (
    id                             TEXT PRIMARY KEY,   -- RecordId hex string
    scoring_structure_id           TEXT NOT NULL,      -- ScoringStructureId hex string
    secondary_scoring_structure_id TEXT NOT NULL,      -- ScoringStructureId hex string
    secondary_index                INTEGER NOT NULL,   -- position within the secondary list
    UNIQUE (scoring_structure_id, secondary_index)
);
