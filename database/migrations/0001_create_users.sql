-- 0001_create_users.sql
-- First migration: the users table, matching the User record shape.
-- ID is stored as a 16-hex TEXT (RecordId.String()) to avoid signed-64
-- overflow on IDs > MaxInt64. bool is stored as INTEGER (0/1).
CREATE TABLE users (
    id           TEXT PRIMARY KEY,   -- RecordId hex string
    first_name   TEXT NOT NULL,
    last_name    TEXT NOT NULL,
    phone_number TEXT NOT NULL,
    email        TEXT NOT NULL UNIQUE,
    verified     INTEGER NOT NULL DEFAULT 0   -- bool
);
