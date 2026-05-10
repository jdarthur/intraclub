package model

import (
	"context"
	"fmt"
	"intraclub/common"
	"math/rand/v2"
	"strings"
	"testing"
	"time"
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

func assertRulesetIsDynamicallyInvalid(t *testing.T, db common.DatabaseProvider, r *Ruleset, textContains string) {
	err := r.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("expected error on DynamicallyValid, but got nil")
	}
	if !strings.Contains(err.Error(), textContains) {
		t.Fatalf("expected error to contain '%s', but got '%v'", textContains, err)
	}
	fmt.Println(err.Error())
}

func newValidRuleset(t *testing.T, owner UserId) *Ruleset {
	x := NewRuleset()
	x.Name = fmt.Sprintf("test ruleset #%d", rand.Int())
	x.Owner = owner
	return x
}

func newValidStoredRuleset(t *testing.T, db common.DatabaseProvider) *Ruleset {
	user := newStoredUser(t, db)
	x := newValidRuleset(t, user.ID)
	v, err := common.CreateOne(context.Background(), db, x)
	if err != nil {
		t.Fatalf("Error creating ruleset: %s\n", err)
	}
	return v
}

func newValidStoredRulesetWithOneSection(t *testing.T, db common.DatabaseProvider) *Ruleset {
	ruleset := newValidStoredRuleset(t, db)
	amended := addSectionRevisionToEndOfExistingRuleset(t, db, ruleset)
	return amended
}

func newValidStoredRulesetWithXSections(t *testing.T, db common.DatabaseProvider, count int) *Ruleset {
	ruleset := newValidStoredRuleset(t, db)
	for i := 0; i < count; i++ {
		ruleset = addSectionRevisionToEndOfExistingRuleset(t, db, ruleset)
	}
	return ruleset
}

func addSectionRevisionToEndOfExistingRuleset(t *testing.T, db common.DatabaseProvider, existing *Ruleset) *Ruleset {
	afterSectionId := RuleSectionId(common.InvalidRecordId)

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
	db := common.NewUnitTestDBProvider()
	ruleset := newValidStoredRuleset(t, db)

	ruleset.Name = "new name"
	err := common.UpdateOne(context.Background(), db, ruleset)
	if err == nil {
		t.Fatal("expected error updating existing ruleset")
	}
	fmt.Println(err)
}
func TestRulesetDateIsAfterSupersedingRuleset(t *testing.T) {
	// Date value for this Ruleset must be before the Date value
	// for the superseding Ruleset (date value isn't very meaningful
	// otherwise. shouldn't be a user-settable value anyway)
	db := common.NewUnitTestDBProvider()
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
	db := common.NewUnitTestDBProvider()
	user2 := newStoredUser(t, db)
	v := newValidStoredRuleset(t, db)
	_, err := v.Fork(context.Background(), db, user2.ID)
	if err == nil {
		t.Fatal("expected error forking ruleset with empty name")
	}
	fmt.Println(err)
}

func TestSupersededByRevisionMustBeOneHigher(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	r := newValidStoredRuleset(t, db)

	v2 := addSectionRevisionToEndOfExistingRuleset(t, db, r)

	r.Revision = v2.Revision
	assertRulesetIsDynamicallyInvalid(t, db, r, "must be greater than this ruleset's revision")
}

func TestSupersededByOwnerMustBeIdentical(t *testing.T) {
	// ruleset cannot be superseded by a ruleset that this not
	// owned by the same user ID

	db := common.NewUnitTestDBProvider()
	user2 := newStoredUser(t, db)
	r := newValidStoredRuleset(t, db)

	v2 := addSectionRevisionToEndOfExistingRuleset(t, db, r)

	v2.Owner = user2.ID
	assertRulesetIsDynamicallyInvalid(t, db, r, "does not match this ruleset's owner")

}
