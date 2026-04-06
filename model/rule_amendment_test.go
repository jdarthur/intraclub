package model

import (
	"fmt"
	"intraclub/common"
	"strings"
	"testing"
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

func assertDynamicallyInvalid(t *testing.T, db common.DatabaseProvider, r *RuleAmendment, containsText string) {
	err := r.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Expected DynamicallyValid to return error")
	}
	if !strings.Contains(err.Error(), containsText) {
		t.Fatalf("expected DynamicallyValid error to contain text:\n   %s\ngot:\n   %s", containsText, err.Error())
	}
	fmt.Println(err)
}

func assertAmendmentIsInvalidForRuleset(t *testing.T, db common.DatabaseProvider, r *Ruleset, a *RuleAmendment, containsText string) {
	_, err := r.Amend(db, a)
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
		TargetSection: RuleSectionId(common.NewRecordId()),
	}
	assertRuleAmendmentIsStaticallyInvalid(t, r, "'target section'")
}

func TestRuleAmendmentTargetSectionIdDoesNotExist(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	r := &RuleAmendment{
		Type:          RuleAmendmentTypeRemoveSection,
		TargetSection: RuleSectionId(common.NewRecordId()),
	}
	assertDynamicallyInvalid(t, db, r, "does not exist")

}

func TestRuleAmendmentTargetSectionIdIsNotInRuleset(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)
	ruleset2 := newValidStoredRulesetWithOneSection(t, db)

	sections, err := ruleset2.GetSections(db)
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
		After:      RuleSectionId(common.NewRecordId()),
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
		TargetSection: RuleSectionId(common.NewRecordId()),
	}
	assertRuleAmendmentIsStaticallyInvalid(t, r, "'new section' contents must be empty")
}

func TestNewContentsIsNotEmptyOnReorderSection(t *testing.T) {
	r := &RuleAmendment{
		Type:       RuleAmendmentTypeReorderSection,
		NewSection: RuleSection{Markdown: "test"},
		After:      RuleSectionId(common.NewRecordId()),
	}
	assertRuleAmendmentIsStaticallyInvalid(t, r, "'new section' contents must be empty")
}

func TestNewContentsMarkdownIsEmptyOnAddSection(t *testing.T) {
	r := &RuleAmendment{
		Type:  RuleAmendmentTypeAddSection,
		After: RuleSectionId(common.NewRecordId()),
	}
	assertRuleAmendmentIsStaticallyInvalid(t, r, "'new section' contents must not be empty")
}

func TestNewContentsMarkdownIsEmptyOnModifySection(t *testing.T) {
	r := &RuleAmendment{
		Type:          RuleAmendmentTypeModifySection,
		TargetSection: RuleSectionId(common.NewRecordId()),
	}
	assertRuleAmendmentIsStaticallyInvalid(t, r, "'new section' contents must not be empty")
}

func TestNewContentsIsNotModifiedOnModifySection(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)

	sections, err := ruleset.GetSections(db)
	if err != nil {
		t.Fatal(err)
	}

	existing, err := common.GetExistingRecordById(db, &RuleSection{}, sections[0].RecordId())
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
		NewSection: RuleSection{Markdown: "test", Parent: RulesetId(common.NewRecordId())},
	}
	assertRuleAmendmentIsStaticallyInvalid(t, r, "'new section' parent must not be set")
}

func TestSectionIdIsSetInAddSectionContents(t *testing.T) {
	r := &RuleAmendment{
		Type:       RuleAmendmentTypeAddSection,
		NewSection: RuleSection{Markdown: "test", ID: RuleSectionId(common.NewRecordId())},
	}
	assertRuleAmendmentIsStaticallyInvalid(t, r, "'new section' ID must not be set")
}

func TestTargetSectionIdIsNotInRulesetOnModifySection(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)
	a := &RuleAmendment{
		Type:          RuleAmendmentTypeModifySection,
		NewSection:    RuleSection{Markdown: "new contents"},
		TargetSection: RuleSectionId(common.NewRecordId()),
	}
	assertAmendmentIsInvalidForRuleset(t, db, ruleset, a, "does not exist")
}

func TestAfterSectionIdIsNotInRulesetOnReorderSection(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)
	ruleset2 := newValidStoredRulesetWithOneSection(t, db)

	sections, err := ruleset.GetSections(db)
	if err != nil {
		t.Fatal(err)
	}
	sections2, err := ruleset2.GetSections(db)
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
	db := common.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)
	ruleset2 := newValidStoredRulesetWithOneSection(t, db)

	sections2, err := ruleset2.GetSections(db)
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
	db := common.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)

	sections, err := ruleset.GetSections(db)
	if err != nil {
		t.Fatal(err)
	}

	a := &RuleAmendment{
		Type:          RuleAmendmentTypeReorderSection,
		TargetSection: sections[0],
		After:         RuleSectionId(common.NewRecordId()),
	}
	assertAmendmentIsInvalidForRuleset(t, db, ruleset, a, "does not exist")
}

func TestAfterSectionIdIsSameAsTargetSectionId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)

	sections, err := ruleset.GetSections(db)
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
	db := common.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)

	sections, err := ruleset.GetSections(db)
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
	db := common.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithOneSection(t, db)

	sections, err := ruleset.GetSections(db)
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
	db := common.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithXSections(t, db, 2)

	sections, err := ruleset.GetSections(db)
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

