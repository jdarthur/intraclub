-- 0050_create_organizations.sql
-- The organization table, matching the Organization record shape
-- (model/organization.go).
-- Table name equals record.Type() ("organization").
--   id         -> OrganizationId hex TEXT primary key
--   user_id    -> owner UserId hex TEXT
--   name       -> TEXT
--   created_at -> RFC3339 TEXT
--   updated_at -> RFC3339 TEXT
--   deleted_at -> RFC3339 TEXT, nullable (*time.Time)
CREATE TABLE organization (
    id         TEXT PRIMARY KEY,   -- OrganizationId hex string
    user_id    TEXT NOT NULL,      -- owner UserId hex string
    name       TEXT NOT NULL,      -- unique organization name
    created_at TEXT NOT NULL,      -- RFC3339
    updated_at TEXT NOT NULL,      -- RFC3339
    deleted_at TEXT                -- RFC3339, nullable
);
