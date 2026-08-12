-- 0051_create_organization_members.sql
-- The organization_member join table, matching the OrganizationMember
-- record shape (model/organization_member.go).
-- Table name equals record.Type() ("organization_member").
--   id              -> RecordId hex TEXT primary key
--   organization_id -> OrganizationId hex TEXT
--   user_id         -> UserId hex TEXT
--   created_at      -> RFC3339 TEXT
--   updated_at      -> RFC3339 TEXT
--   deleted_at      -> RFC3339 TEXT, nullable (*time.Time)
CREATE TABLE organization_member (
    id              TEXT PRIMARY KEY,   -- RecordId hex string
    organization_id TEXT NOT NULL,      -- OrganizationId hex string
    user_id         TEXT NOT NULL,      -- UserId hex string
    created_at      TEXT NOT NULL,      -- RFC3339
    updated_at      TEXT NOT NULL,      -- RFC3339
    deleted_at      TEXT                -- RFC3339, nullable
);
