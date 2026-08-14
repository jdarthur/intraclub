package model

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	"intraclub/database"
)

func assertRulesetIsStaticallyInvalid(t *testing.T, r *Ruleset, textContains string) {
	err := r.StaticallyValid()
	if err == nil {
		t.Fatal("expected error on StaticallyValid, but got nil")
	}
	if !strings.Contains(err.Error(), textContains) {
		t.Fatalf("expected error to contain '%s', but got '%v'", textContains, err)
	}
	fmt.Println(err.Error())
}

func assertRulesetIsDynamicallyInvalid(t *testing.T, db database.Provider, r *Ruleset, textContains string) {
	err := r.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("expected error on DynamicallyValid, but got nil")
	}
	if !strings.Contains(err.Error(), textContains) {
		t.Fatalf("expected error to contain '%s', but got '%v'", textContains, err)
	}
	fmt.Println(err.Error())
}

func newValidRuleset(t *testing.T, owner database.UserId) *Ruleset {
	x := NewRuleset()
	x.Name = fmt.Sprintf("test ruleset #%d", rand.Int())
	x.Owner = owner
	return x
}

func newValidStoredRuleset(t *testing.T, db database.Provider) *Ruleset {
	user := newStoredUser(t, db)
	x := newValidRuleset(t, user.ID)
	v, err := database.CreateOne(context.Background(), db, x)
	if err != nil {
		t.Fatalf("Error creating ruleset: %s\n", err)
	}
	return v
}

func newValidStoredRulesetWithOneSection(t *testing.T, db database.Provider) *Ruleset {
	ruleset := newValidStoredRuleset(t, db)
	amended := addSectionRevisionToEndOfExistingRuleset(t, db, ruleset)
	return amended
}

func newValidStoredRulesetWithXSections(t *testing.T, db database.Provider, count int) *Ruleset {
	ruleset := newValidStoredRuleset(t, db)
	for i := 0; i < count; i++ {
		ruleset = addSectionRevisionToEndOfExistingRuleset(t, db, ruleset)
	}
	return ruleset
}

func addSectionRevisionToEndOfExistingRuleset(t *testing.T, db database.Provider, existing *Ruleset) *Ruleset {
	afterSectionId := RuleSectionId(database.InvalidRecordId)

	sections, err := existing.GetSections(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	if len(sections) != 0 {
		afterSectionId = sections[len(sections)-1]
	}

	amendment := &RuleAmendment{
		Type: RuleAmendmentTypeAddSection,
		NewSection: RuleSection{
			Markdown: fmt.Sprintf("test contents #%d", rand.Int()),
		},
		After: afterSectionId,
	}

	v, err := existing.Amend(context.Background(), db, amendment)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestRulesetNameIsEmpty(t *testing.T) {
	ruleset := NewRuleset()
	assertRulesetIsStaticallyInvalid(t, ruleset, "name must not be empty")
}

func TestRulesetRevisionIsNegative(t *testing.T) {
	ruleset := NewRuleset()
	ruleset.Name = "some value"
	ruleset.Revision = -1
	assertRulesetIsStaticallyInvalid(t, ruleset, "cannot be negative")
}

func TestSectionsNonEmptyWhenRevisionIsZero(t *testing.T) {
	ruleset := NewRuleset()
	ruleset.Name = "test"
	ruleset.Revision = 0

	// Cannot test this anymore as Sections field was removed
	// This test is now obsolete since sections are stored in a join table
}

func TestRulesetCannotBeUpdatedWithoutAmendment(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ruleset := newValidStoredRuleset(t, db)

	ruleset.Name = "new name"
	err := database.UpdateOne(context.Background(), db, ruleset)
	if err == nil {
		t.Fatal("expected error updating existing ruleset")
	}
	fmt.Println(err)
}

func TestRulesetEditNameCreatesSupersedingRevision(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ruleset := newValidStoredRuleset(t, db)

	newRevision, err := ruleset.EditName(context.Background(), db, "Amended Name")
	if err != nil {
		t.Fatalf("unexpected error amending ruleset name: %s", err)
	}

	// the new revision supersedes the old and keeps the owner
	if newRevision.Name != "Amended Name" {
		t.Fatalf("expected new revision name %q, got %q", "Amended Name", newRevision.Name)
	}
	if newRevision.Revision != ruleset.Revision+1 {
		t.Fatalf("expected revision %d, got %d", ruleset.Revision+1, newRevision.Revision)
	}
	if newRevision.Owner != ruleset.Owner {
		t.Fatalf("expected owner %s to be preserved, got %s", ruleset.Owner, newRevision.Owner)
	}

	// the original ruleset is now marked as superseded
	reloaded, err := database.GetExistingRecordById(context.Background(), db, &Ruleset{}, ruleset.ID.RecordId())
	if err != nil {
		t.Fatalf("unexpected error reloading original ruleset: %s", err)
	}
	if reloaded.SupersededBy != newRevision.ID {
		t.Fatalf("expected original ruleset to be superseded by %s, got %s", newRevision.ID, reloaded.SupersededBy)
	}

	// the original ruleset cannot be directly updated even after amending
	err = database.UpdateOne(context.Background(), db, ruleset)
	if err == nil {
		t.Fatal("expected error updating existing ruleset after amend")
	}
}

func TestRulesetEditNameRequiresDifferentNonEmptyName(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ruleset := newValidStoredRuleset(t, db)

	if _, err := ruleset.EditName(context.Background(), db, ""); err == nil {
		t.Fatal("expected error amending ruleset with empty name")
	}
	if _, err := ruleset.EditName(context.Background(), db, ruleset.Name); err == nil {
		t.Fatal("expected error amending ruleset with unchanged name")
	}
}

func TestRulesetDateIsAfterSupersedingRuleset(t *testing.T) {
	// Date value for this Ruleset must be before the Date value
	// for the superseding Ruleset (date value isn't very meaningful
	// otherwise. shouldn't be a user-settable value anyway)
	db := database.NewUnitTestDBProvider()
	r := newValidStoredRuleset(t, db)

	_ = addSectionRevisionToEndOfExistingRuleset(t, db, r)

	r.Date = time.Now()
	assertRulesetIsDynamicallyInvalid(t, db, r, "must be after this ruleset's date")
}

func TestRulesetDuplicateSections(t *testing.T) {
	// Cannot test this anymore as Sections field was removed
	// This test is now obsolete since sections are stored in a join table
}

func TestRulesetForkEmpty(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	user2 := newStoredUser(t, db)
	v := newValidStoredRuleset(t, db)
	_, err := v.Fork(context.Background(), db, user2.ID)
	if err == nil {
		t.Fatal("expected error forking ruleset with empty name")
	}
	fmt.Println(err)
}

func TestSupersededByRevisionMustBeOneHigher(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	r := newValidStoredRuleset(t, db)

	v2 := addSectionRevisionToEndOfExistingRuleset(t, db, r)

	r.Revision = v2.Revision
	assertRulesetIsDynamicallyInvalid(t, db, r, "must be greater than this ruleset's revision")
}

func TestSupersededByOwnerMustBeIdentical(t *testing.T) {
	// ruleset cannot be superseded by a ruleset that this not
	// owned by the same user ID

	db := database.NewUnitTestDBProvider()
	user2 := newStoredUser(t, db)
	r := newValidStoredRuleset(t, db)

	v2 := addSectionRevisionToEndOfExistingRuleset(t, db, r)

	v2.Owner = user2.ID
	assertRulesetIsDynamicallyInvalid(t, db, r, "does not match this ruleset's owner")

}

func TestRuleSectionIdJSONRoundTrip(t *testing.T) {
	// RuleSectionId / RulesetId must be parsed back from JSON bodies (used by
	// the RuleAmendment request). A value-receiver UnmarshalJSON silently
	// discards the parsed value, so this guards the pointer-receiver fix.
	rid, err := database.RecordIdFromString("1234567890abcdef")
	if err != nil {
		t.Fatal(err)
	}
	id := RuleSectionId(rid)
	raw, err := id.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var parsed RuleSectionId
	if err := parsed.UnmarshalJSON(raw); err != nil {
		t.Fatal(err)
	}
	if parsed != id {
		t.Fatalf("expected %s after round trip, got %s", id, parsed)
	}
}

func TestRulesetIdJSONRoundTrip(t *testing.T) {
	rid, err := database.RecordIdFromString("fedcba9876543210")
	if err != nil {
		t.Fatal(err)
	}
	id := RulesetId(rid)
	raw, err := id.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var parsed RulesetId
	if err := parsed.UnmarshalJSON(raw); err != nil {
		t.Fatal(err)
	}
	if parsed != id {
		t.Fatalf("expected %s after round trip, got %s", id, parsed)
	}
}

func TestRulesetReorderSectionToFront(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithXSections(t, db, 3)

	sections, err := ruleset.GetSections(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}
	first, second := sections[0], sections[1]

	// Move the second section to the front (empty After).
	amended, err := ruleset.Amend(context.Background(), db, &RuleAmendment{
		Type:          RuleAmendmentTypeReorderSection,
		TargetSection: second,
		After:         RuleSectionId(database.InvalidRecordId),
	})
	if err != nil {
		t.Fatalf("unexpected error reordering section to front: %s", err)
	}

	got, err := amended.GetSections(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 sections after reorder, got %d", len(got))
	}
	if got[0] != second || got[1] != first {
		t.Fatalf("expected [%s %s ...] after reorder, got [%s %s ...]", second, first, got[0], got[1])
	}
}

func TestRulesetReorderSectionAfter(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	ruleset := newValidStoredRulesetWithXSections(t, db, 3)

	sections, err := ruleset.GetSections(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	first, third := sections[0], sections[2]

	// Move the first section after the third (to the end).
	amended, err := ruleset.Amend(context.Background(), db, &RuleAmendment{
		Type:          RuleAmendmentTypeReorderSection,
		TargetSection: first,
		After:         third,
	})
	if err != nil {
		t.Fatalf("unexpected error reordering section: %s", err)
	}

	got, err := amended.GetSections(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 sections after reorder, got %d", len(got))
	}
	if got[2] != first {
		t.Fatalf("expected first section to move to end, got last %s", got[2])
	}
}

func TestRulesetPostDeleteCascadesChildren(t *testing.T) {	db := database.NewUnitTestDBProvider()
	ruleset := newValidStoredRuleset(t, db)

	section := &RuleSection{Parent: ruleset.ID, Title: "title", Markdown: "contents", Owner: ruleset.Owner}
	created, err := database.CreateOne(context.Background(), db, section)
	if err != nil {
		t.Fatal(err)
	}
	err = ruleset.AddSection(context.Background(), db, created.ID, 0)
	if err != nil {
		t.Fatal(err)
	}

	counts := func() (int, int) {
		sections, err := database.GetAllWhere[*RuleSection](context.Background(), db, func(_ context.Context, s *RuleSection) bool {
			return s.Parent == ruleset.ID
		})
		if err != nil {
			t.Fatal(err)
		}
		relations, err := database.GetAllWhere[*RulesetSection](context.Background(), db, func(_ context.Context, s *RulesetSection) bool {
			return s.RulesetId == ruleset.ID
		})
		if err != nil {
			t.Fatal(err)
		}
		return len(sections), len(relations)
	}

	sections, relations := counts()
	if sections != 1 || relations != 1 {
		t.Fatalf("expected 1 section and 1 relation, got sections=%d relations=%d", sections, relations)
	}

	_, _, err = database.DeleteOneById(context.Background(), db, &Ruleset{}, ruleset.ID.RecordId())
	if err != nil {
		t.Fatal(err)
	}

	sections, relations = counts()
	if sections != 0 || relations != 0 {
		t.Fatalf("expected 0 after delete, got sections=%d relations=%d", sections, relations)
	}
}
