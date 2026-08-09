-- 0047_create_comment_reactions.sql
-- The comment_reaction child table, matching the CommentReaction record shape
-- (model/comment.go). This is the normalized replacement for the former inline
-- Comment.Reactions slice.
-- Table name equals record.Type() ("comment_reaction").
--   id             -> RecordId hex TEXT primary key
--   comment_id     -> CommentId hex TEXT
--   user_id        -> UserId hex TEXT
--   reaction_type  -> INTEGER (reactionType)
-- One reaction per (comment, user, type): UNIQUE(comment_id, user_id,
-- reaction_type) mirrors the no-duplicate-reaction invariant of
-- ReactionList.StaticallyValid.
CREATE TABLE comment_reaction (
    id             TEXT PRIMARY KEY,   -- RecordId hex string
    comment_id     TEXT NOT NULL,      -- CommentId hex string
    user_id        TEXT NOT NULL,      -- UserId hex string
    reaction_type  INTEGER NOT NULL,   -- reactionType
    UNIQUE (comment_id, user_id, reaction_type)
);
