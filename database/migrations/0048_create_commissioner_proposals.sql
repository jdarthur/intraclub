-- 0048_create_commissioner_proposals.sql
-- The commissioner_proposal table, matching the CommissionerProposal record
-- shape (model/commissioner_proposal.go).
-- Table name equals record.Type() ("commissioner_proposal").
--   id                -> RecordId hex TEXT primary key
--   description       -> TEXT
--   season_id         -> SeasonId hex TEXT
--   must_be_unanimous -> INTEGER (bool: 0/1)
CREATE TABLE commissioner_proposal (
    id                TEXT PRIMARY KEY,   -- RecordId hex string
    description       TEXT NOT NULL,
    season_id         TEXT NOT NULL,      -- SeasonId hex string
    must_be_unanimous INTEGER NOT NULL
);
