-- 0045_create_comments.sql
-- The comment table, matching the Comment record shape (model/comment.go).
-- Table name equals record.Type() ("comment").
--   id         -> CommentId hex TEXT primary key
--   blurb      -> BlurbId hex TEXT
--   reply_to   -> CommentId hex TEXT
--   user_id    -> UserId hex TEXT (Comment.Owner)
--   content    -> TEXT
--   edited_at  -> RFC3339 TEXT
--   created_at -> RFC3339 TEXT
-- The former inline Comment.Reactions slice was normalized into the
-- comment_reaction child table (0047); it is not part of this table.
CREATE TABLE comment (
    id         TEXT PRIMARY KEY,   -- CommentId hex string
    blurb      TEXT NOT NULL,      -- BlurbId hex string
    reply_to   TEXT NOT NULL,      -- CommentId hex string
    user_id    TEXT NOT NULL,      -- UserId hex string
    content    TEXT NOT NULL,
    edited_at  TEXT NOT NULL,      -- RFC3339
    created_at TEXT NOT NULL       -- RFC3339
);
