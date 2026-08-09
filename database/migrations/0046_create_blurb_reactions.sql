-- 0046_create_blurb_reactions.sql
-- The blurb_reaction child table, matching the BlurbReaction record shape
-- (model/blurb.go). This is the normalized replacement for the former inline
-- Blurb.Reactions slice.
-- Table name equals record.Type() ("blurb_reaction").
--   id             -> RecordId hex TEXT primary key
--   blurb_id       -> BlurbId hex TEXT
--   user_id        -> UserId hex TEXT
--   reaction_type  -> INTEGER (reactionType)
-- One reaction per (blurb, user, type): UNIQUE(blurb_id, user_id,
-- reaction_type) mirrors the no-duplicate-reaction invariant of
-- ReactionList.StaticallyValid.
CREATE TABLE blurb_reaction (
    id             TEXT PRIMARY KEY,   -- RecordId hex string
    blurb_id       TEXT NOT NULL,      -- BlurbId hex string
    user_id        TEXT NOT NULL,      -- UserId hex string
    reaction_type  INTEGER NOT NULL,   -- reactionType
    UNIQUE (blurb_id, user_id, reaction_type)
);
