-- 0029_create_ruleset_sections.sql
-- The ruleset_section join table, matching the RulesetSection record shape
-- (model/ruleset_section.go). Table name equals record.Type() ("ruleset_section").
--   id            -> RecordId hex TEXT primary key
--   ruleset_id    -> RulesetId hex TEXT
--   section_id    -> RuleSectionId hex TEXT
--   section_index -> INTEGER (ordering of sections within the ruleset)
--   created_at    -> RFC3339 TEXT
--   updated_at    -> RFC3339 TEXT
CREATE TABLE ruleset_section (
    id            TEXT PRIMARY KEY,   -- RecordId hex string
    ruleset_id    TEXT NOT NULL,      -- RulesetId hex string
    section_id    TEXT NOT NULL,      -- RuleSectionId hex string
    section_index INTEGER NOT NULL,   -- section ordering within the ruleset
    created_at    TEXT NOT NULL,      -- RFC3339
    updated_at    TEXT NOT NULL       -- RFC3339
);
