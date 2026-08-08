-- 0004_create_email_verification_tokens.sql
-- The email_verification_token table, matching the EmailToken record shape
-- (model/email_token.go).
-- Table name equals record.Type() ("email_verification_token").
--   id      -> RecordId hex TEXT primary key
--   token   -> TEXT (unique per EmailToken.UniquenessEquivalent)
--   user_id -> UserId hex TEXT (unique: one verification token per user)
CREATE TABLE email_verification_token (
    id      TEXT PRIMARY KEY,   -- RecordId hex string
    token   TEXT NOT NULL UNIQUE,
    user_id TEXT NOT NULL UNIQUE   -- UserId hex string
);
