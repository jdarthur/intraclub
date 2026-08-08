-- 0002_create_users.sql
-- The user table, matching the User record shape (model/user.go).
-- Table name equals record.Type() ("user").
--   id           -> RecordId hex TEXT primary key
--   first_name   -> TEXT
--   last_name    -> TEXT
--   phone_number -> TEXT (PhoneNumber is a string alias)
--   email        -> TEXT (EmailAddress is a string alias; unique per User.UniquenessEquivalent)
--   verified     -> bool as INTEGER (0/1)
CREATE TABLE user (
    id           TEXT PRIMARY KEY,   -- RecordId hex string
    first_name   TEXT NOT NULL,
    last_name    TEXT NOT NULL,
    phone_number TEXT NOT NULL,
    email        TEXT NOT NULL UNIQUE,
    verified     INTEGER NOT NULL DEFAULT 0   -- bool
);
