package model

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"intraclub/database"
)

func assertRuleAmendmentIsStaticallyInvalid(t *testing.T, r *RuleAmendment, containsText string) {
	err := r.StaticallyValid()
	if err == nil {
		t.Fatal("Expected StaticallyValid to return error")
	}
	if !strings.Contains(err.Error(), containsText) {
		t.Fatalf("expected StaticallyValid error `to contain text:\n   %s\ngot:\n   %s", containsText, err.Error())
	}
	fmt.Println(err)
}

func assertDynamicallyInvalid(t *testing.T, db database.Provider, r *RuleAmendment, containsText string) {
	err := r.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("Expected DynamicallyValid to return error")
	}
	if !strings.Contains(err.Error(), containsText) {
		t.Fatalf("expected DynamicallyValid error to contain text:\n   %s\ngot:\n   %s", containsText, err.Error())
	}
	fmt.Println(err)
}

func assertAmendmentIsInvalidForRuleset(t *testing.T, db database.Provider, r *Ruleset, a *RuleAmendment, containsText string) {
	_, err := r.Amend(context.Background(), db, a)
	if err == nil {
		t.Fatal("Expected Amend to return error")
	}
	if !strings.Contains(err.Error(), containsText) {
		t.Fatalf("expected Amend error to contain text:\n   %s\ngot:\n   %s", containsText, err.Error())
	}
	fmt.Println(err)
}

func TestRuleAmendmentTypeNegative(t *testing.T) {
	r := &RuleAmendment{Type: -1}
	assertRuleAmendmentIsStaticallyInvalid(t, r, "invalid rule amendment type")
}

func TestRuleAmendmentTypeInvalid(t *testing.T) {
	r := &RuleAmendment{Type: 999}
	assertRuleAmendmentIsStaticallyInvalid(t, r, "invalid rule amendment type")
}

func TestRuleAmendmentTargetSectionIdIsNonEmptyOnAddSection(t *testing.T) {
	r := &RuleAmendment{
		Type:          RuleAmendmentTypeAddSection,
		TargetSection: RuleSectionId(database.NewRecordId()),
	}
	assertRuleAmendmentIsStaticallyInvalid(t, r, "'target section'")
}

func TestRuleAmendmentTargetSectionIdDoesNotExist(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	r := &RuleAmendment{
		Type:          RuleAmendmentTypeRemoveSection,
		TargetSection: RuleSectionId(database.NewRecordId()),
	}
	assertDynamicallyInvalid(t, db, r, "does not exist")

}

func TestRuleAmendmentTargetSectionIdIsNotInRuleset(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)
	ruleset2 := newValidStoredRulesetWithOneSection(t, db)

	sections, err := ruleset2.GetSections(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	a := &RuleAmendment{
		Type:          RuleAmendmentTypeModifySection,
		NewSection:    RuleSection{Markdown: "test"},
		TargetSection: sections[0],
	}
	assertAmendmentIsInvalidForRuleset(t, db, ruleset, a, "was not found in ruleset")
}

func TestNewContentsEmpty(t *testing.T) {
	r := RuleAmendment{
		Type:       RuleAmendmentTypeAddSection,
		NewSection: RuleSection{Markdown: ""},
		After:      RuleSectionId(database.NewRecordId()),
	}
	err := r.StaticallyValid()
	if err == nil {
		t.Fatal("Expected StaticallyValid to return error")
	}
	fmt.Println(err)
}

func TestNewContentsIsNotEmptyOnRemoveSection(t *testing.T) {
	r := &RuleAmendment{
		Type:          RuleAmendmentTypeRemoveSection,
		NewSection:    RuleSection{Markdown: "test"},
		TargetSection: RuleSectionId(database.NewRecordId()),
	}
	assertRuleAmendmentIsStaticallyInvalid(t, r, "'new section' contents must be empty")
}

func TestNewContentsIsNotEmptyOnReorderSection(t *testing.T) {
	r := &RuleAmendment{
		Type:       RuleAmendmentTypeReorderSection,
		NewSection: RuleSection{Markdown: "test"},
		After:      RuleSectionId(database.NewRecordId()),
	}
	assertRuleAmendmentIsStaticallyInvalid(t, r, "'new section' contents must be empty")
}

func TestNewContentsMarkdownIsEmptyOnAddSection(t *testing.T) {
	r := &RuleAmendment{
		Type:  RuleAmendmentTypeAddSection,
		After: RuleSectionId(database.NewRecordId()),
	}
	assertRuleAmendmentIsStaticallyInvalid(t, r, "'new section' contents must not be empty")
}

func TestNewContentsMarkdownIsEmptyOnModifySection(t *testing.T) {
	r := &RuleAmendment{
		Type:          RuleAmendmentTypeModifySection,
		TargetSection: RuleSectionId(database.NewRecordId()),
	}
	assertRuleAmendmentIsStaticallyInvalid(t, r, "'new section' contents must not be empty")
}

func TestNewContentsIsNotModifiedOnModifySection(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)

	sections, err := ruleset.GetSections(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	existing, err := database.GetExistingRecordById(context.Background(), db, &RuleSection{}, sections[0].RecordId())
	if err != nil {
		t.Fatal(err)
	}

	a := &RuleAmendment{
		Type:          RuleAmendmentTypeModifySection,
		TargetSection: sections[0],
		NewSection:    RuleSection{Markdown: existing.Markdown},
	}
	assertAmendmentIsInvalidForRuleset(t, db, ruleset, a, "is not updated")
}

func TestParentRuleIdIsSetOnAddSectionContents(t *testing.T) {
	r := &RuleAmendment{
		Type:       RuleAmendmentTypeAddSection,
		NewSection: RuleSection{Markdown: "test", Parent: RulesetId(database.NewRecordId())},
	}
	assertRuleAmendmentIsStaticallyInvalid(t, r, "'new section' parent must not be set")
}

func TestSectionIdIsSetInAddSectionContents(t *testing.T) {
	r := &RuleAmendment{
		Type:       RuleAmendmentTypeAddSection,
		NewSection: RuleSection{Markdown: "test", ID: RuleSectionId(database.NewRecordId())},
	}
	assertRuleAmendmentIsStaticallyInvalid(t, r, "'new section' ID must not be set")
}

func TestTargetSectionIdIsNotInRulesetOnModifySection(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)
	a := &RuleAmendment{
		Type:          RuleAmendmentTypeModifySection,
		NewSection:    RuleSection{Markdown: "new contents"},
		TargetSection: RuleSectionId(database.NewRecordId()),
	}
	assertAmendmentIsInvalidForRuleset(t, db, ruleset, a, "does not exist")
}

func TestAfterSectionIdIsNotInRulesetOnReorderSection(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)
	ruleset2 := newValidStoredRulesetWithOneSection(t, db)

	sections, err := ruleset.GetSections(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	sections2, err := ruleset2.GetSections(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	a := &RuleAmendment{
		Type:          RuleAmendmentTypeReorderSection,
		TargetSection: sections[0],
		After:         sections2[0],
	}
	assertAmendmentIsInvalidForRuleset(t, db, ruleset, a, "was not found in ruleset")
}

func TestAfterSectionIdIsNotInRulesetOnAddSection(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)
	ruleset2 := newValidStoredRulesetWithOneSection(t, db)

	sections2, err := ruleset2.GetSections(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	a := &RuleAmendment{
		Type:       RuleAmendmentTypeAddSection,
		NewSection: RuleSection{Markdown: "test1"},
		After:      sections2[0],
	}
	assertAmendmentIsInvalidForRuleset(t, db, ruleset, a, "was not found in ruleset")
}

func TestAfterSectionIdDoesNotExistOnReorderSection(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)

	sections, err := ruleset.GetSections(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	a := &RuleAmendment{
		Type:          RuleAmendmentTypeReorderSection,
		TargetSection: sections[0],
		After:         RuleSectionId(database.NewRecordId()),
	}
	assertAmendmentIsInvalidForRuleset(t, db, ruleset, a, "does not exist")
}

func TestAfterSectionIdIsSameAsTargetSectionId(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)

	sections, err := ruleset.GetSections(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	a := &RuleAmendment{
		Type:          RuleAmendmentTypeReorderSection,
		TargetSection: sections[0],
		After:         sections[0],
	}
	assertAmendmentIsInvalidForRuleset(t, db, ruleset, a, "must be different than after section")
}

func TestAfterSectionIdIsSetOnRemoveSection(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)

	sections, err := ruleset.GetSections(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	a := &RuleAmendment{
		Type:          RuleAmendmentTypeRemoveSection,
		TargetSection: sections[0],
		After:         sections[0],
	}
	assertAmendmentIsInvalidForRuleset(t, db, ruleset, a, "'after section ID' must be empty")
}

func TestAfterSectionIdIsSetOnModifySection(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)

	sections, err := ruleset.GetSections(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	a := &RuleAmendment{
		Type:          RuleAmendmentTypeModifySection,
		TargetSection: sections[0],
		NewSection:    RuleSection{Markdown: "new contents"},
		After:         sections[0],
	}
	assertAmendmentIsInvalidForRuleset(t, db, ruleset, a, "'after section ID' must be empty")
}

func TestAfterSectionIdIsUnchangedOnReorderSection(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithXSections(t, db, 2)

	sections, err := ruleset.GetSections(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	a := &RuleAmendment{
		Type:          RuleAmendmentTypeReorderSection,
		TargetSection: sections[1],
		After:         sections[0],
	}
	assertAmendmentIsInvalidForRuleset(t, db, ruleset, a, "does not reorder")
}
