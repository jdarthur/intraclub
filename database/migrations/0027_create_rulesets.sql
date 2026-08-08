-- 0027_create_rulesets.sql
-- The ruleset table, matching the Ruleset record shape (model/ruleset.go).
-- Table name equals record.Type() ("ruleset").
--   id           -> RulesetId hex TEXT primary key
--   name         -> TEXT
--   revision     -> INTEGER
--   superseded_by -> RulesetId hex TEXT (0 when not archived)
--   date         -> RFC3339 TEXT
--   owner        -> UserId hex TEXT
CREATE TABLE ruleset (
    id            TEXT PRIMARY KEY,   -- RulesetId hex string
    name          TEXT NOT NULL,
    revision      INTEGER NOT NULL,
    superseded_by TEXT NOT NULL,      -- RulesetId hex string
    date          TEXT NOT NULL,      -- RFC3339
    owner         TEXT NOT NULL       -- UserId hex string
);
