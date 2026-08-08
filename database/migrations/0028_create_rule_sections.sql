-- 0028_create_rule_sections.sql
-- The rule_section table, matching the RuleSection record shape
-- (model/ruleset.go). Table name equals record.Type() ("rule_section").
-- The primary key is the record's own hex RuleSectionId (mapped via GetId /
-- SetId). RuleSection.ID carries the JSON tag "section_id", so the reflection
-- mapper also emits a section_id column holding the same value (see
-- docs/schema-conventions.md); both are present so insert/scan round-trips.
--   id          -> RuleSectionId hex TEXT primary key
--   section_id  -> RuleSectionId hex TEXT (JSON-tagged alias of the id)
--   parent      -> RulesetId hex TEXT
--   title       -> TEXT
--   markdown    -> TEXT
--   owner       -> UserId hex TEXT
CREATE TABLE rule_section (
    id         TEXT PRIMARY KEY,   -- RuleSectionId hex string
    section_id TEXT NOT NULL,      -- RuleSectionId hex string (json "section_id")
    parent     TEXT NOT NULL,      -- RulesetId hex string
    title      TEXT NOT NULL,
    markdown   TEXT NOT NULL,
    owner      TEXT NOT NULL       -- UserId hex string
);
