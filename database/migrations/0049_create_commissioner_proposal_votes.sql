-- 0049_create_commissioner_proposal_votes.sql
-- The commissioner_proposal_vote table, matching the CommissionerProposalVote
-- record shape (model/commissioner_proposal.go).
-- Table name equals record.Type() ("commissioner_proposal_vote").
--   id          -> RecordId hex TEXT primary key
--   proposal_id -> RecordId hex TEXT
--   user_id     -> UserId hex TEXT
--   vote        -> INTEGER (bool: 0/1)
-- One vote per (proposal, user): UNIQUE(proposal_id, user_id) mirrors
-- CommissionerProposalVote.UniquenessEquivalent (a voter may only cast one
-- vote per proposal).
CREATE TABLE commissioner_proposal_vote (
    id          TEXT PRIMARY KEY,   -- RecordId hex string
    proposal_id TEXT NOT NULL,      -- RecordId hex string
    user_id     TEXT NOT NULL,      -- UserId hex string
    vote        INTEGER NOT NULL,
    UNIQUE (proposal_id, user_id)
);
