-- 0003_create_user_role_assignments.sql
-- The user_role_assignment table, matching the UserRoleAssignment record
-- shape (model/role.go).
-- Table name equals record.Type() ("user_role_assignment").
--   id           -> RecordId hex TEXT primary key
--   user_id      -> UserId hex TEXT
--   role         -> Role enum as INTEGER
--   reference_id -> RecordId hex TEXT (the referenced record for the role,
--                   e.g. a Team ID for a TeamMember role; InvalidRecordId
--                   encodes as all-zero hex)
-- One assignment per (user, role): UNIQUE(user_id, role).
CREATE TABLE user_role_assignment (
    id           TEXT PRIMARY KEY,   -- RecordId hex string
    user_id      TEXT NOT NULL,      -- UserId hex string
    role         INTEGER NOT NULL,   -- Role enum
    reference_id TEXT NOT NULL,      -- RecordId hex string
    UNIQUE (user_id, role)
);
