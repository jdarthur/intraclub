-- 0010_create_team_assignments.sql
-- The team_assignment table, matching the TeamAssignment record shape
-- (model/team.go).
-- Table name equals record.Type() ("team_assignment").
--   id         -> RecordId hex TEXT primary key
--   team_id    -> TeamId hex TEXT
--   user_id    -> UserId hex TEXT
--   role       -> TeamRole string enum as TEXT
--   created_at -> RFC3339 TEXT
--   updated_at -> RFC3339 TEXT
--   deleted_at -> RFC3339 TEXT, nullable (*time.Time)
CREATE TABLE team_assignment (
    id         TEXT PRIMARY KEY,   -- RecordId hex string
    team_id    TEXT NOT NULL,      -- TeamId hex string
    user_id    TEXT NOT NULL,      -- UserId hex string
    role       TEXT NOT NULL,      -- TeamRole enum ("captain"|"co_captain"|"member")
    created_at TEXT NOT NULL,      -- RFC3339
    updated_at TEXT NOT NULL,      -- RFC3339
    deleted_at TEXT                -- RFC3339, nullable
);
