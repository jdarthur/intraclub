package database

import (
	"context"
	"testing"
)

// Regression coverage for the re-entrant-provider deadlock (the in-memory
// provider held its non-reentrant mutex while running a where predicate, so a
// record whose AccessibleTo/EditableBy re-enters the provider deadlocked — e.g.
// Team → TeamAssignment). See unit_test_db_provider.go GetAllWhere.

// reentrantGroup models a record whose AccessibleTo resolves its members by
// reading the reentrantMembership join table, i.e. it re-enters the provider
// from inside an access-control query. This mirrors model.Team (whose
// AccessibleTo calls GetMembers → GetAllWhere[TeamAssignment]).
type reentrantGroup struct {
	ID   RecordId `json:"id"`
	Name string   `json:"name"`
}

func (g *reentrantGroup) GetOwner() UserId            { return InvalidUserId }
func (g *reentrantGroup) SetOwner(UserId)             {}
func (g *reentrantGroup) Type() string                { return "reentrant_group" }
func (g *reentrantGroup) GetId() RecordId             { return g.ID }
func (g *reentrantGroup) SetId(id RecordId)           { g.ID = id }
func (g *reentrantGroup) NewRecord() CrudRecord       { return new(reentrantGroup) }
func (g *reentrantGroup) StaticallyValid() error      { return nil }
func (g *reentrantGroup) DynamicallyValid(context.Context, Provider) error { return nil }
func (g *reentrantGroup) EditableBy(context.Context, Provider) []UserId     { return nil }

// AccessibleTo re-enters the provider by reading this group's memberships.
func (g *reentrantGroup) AccessibleTo(ctx context.Context, db Provider) []UserId {
	members, err := GetAllWhere[*reentrantMembership](ctx, db, func(_ context.Context, m *reentrantMembership) bool {
		return m.GroupId == g.ID
	})
	if err != nil {
		return nil
	}
	ids := make([]UserId, 0, len(members))
	for _, m := range members {
		ids = append(ids, m.UserId)
	}
	return ids
}

// reentrantMembership is a join row linking a user to a group.
type reentrantMembership struct {
	ID      RecordId `json:"id"`
	GroupId RecordId `json:"group_id"`
	UserId  UserId   `json:"user_id"`
}

func (m *reentrantMembership) GetOwner() UserId              { return InvalidUserId }
func (m *reentrantMembership) SetOwner(UserId)               {}
func (m *reentrantMembership) Type() string                  { return "reentrant_membership" }
func (m *reentrantMembership) GetId() RecordId               { return m.ID }
func (m *reentrantMembership) SetId(id RecordId)             { m.ID = id }
func (m *reentrantMembership) NewRecord() CrudRecord         { return new(reentrantMembership) }
func (m *reentrantMembership) StaticallyValid() error        { return nil }
func (m *reentrantMembership) DynamicallyValid(context.Context, Provider) error { return nil }
func (m *reentrantMembership) EditableBy(context.Context, Provider) []UserId     { return nil }
func (m *reentrantMembership) AccessibleTo(context.Context, Provider) []UserId   { return AccessibleToEveryone }

func newStoredReentrantGroup(t *testing.T, db Provider) *reentrantGroup {
	t.Helper()
	g := &reentrantGroup{}
	v, err := CreateOne(context.Background(), db, g)
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	return v
}

func addReentrantMembership(t *testing.T, db Provider, groupId RecordId, userId UserId) {
	t.Helper()
	m := &reentrantMembership{GroupId: groupId, UserId: userId}
	if _, err := CreateOne(context.Background(), db, m); err != nil {
		t.Fatalf("create membership: %v", err)
	}
}

// testGetAllWithReentrantAccessibleTo exercises WithAccessControl.GetAll over a
// record whose AccessibleTo re-enters the provider. On the in-memory provider
// this previously deadlocked because the outer GetAllWhere held the mutex while
// the access-control filter issued a nested read.
func testGetAllWithReentrantAccessibleTo(t *testing.T, db Provider) {
	memberA := UserId(NewRecordId())
	memberB := UserId(NewRecordId())
	outsider := UserId(NewRecordId())

	// Group 1 is shared with A and B; group 2 belongs to nobody the caller knows.
	g1 := newStoredReentrantGroup(t, db)
	addReentrantMembership(t, db, g1.ID, memberA)
	addReentrantMembership(t, db, g1.ID, memberB)
	g2 := newStoredReentrantGroup(t, db)

	// A can see only group 1.
	wac := WithAccessControl[*reentrantGroup]{Database: db, AccessControlUser: memberA}
	results, err := wac.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll with re-entrant AccessibleTo: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 group for member A, got %d", len(results))
	}
	if results[0].ID != g1.ID {
		t.Fatalf("expected group 1, got %s", results[0].ID)
	}

	// An outsider with no memberships sees no groups.
	wacOutsider := WithAccessControl[*reentrantGroup]{Database: db, AccessControlUser: outsider}
	results, err = wacOutsider.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll (outsider): %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 groups for outsider, got %d", len(results))
	}

	// g2 has no members, so it is not visible to anyone via GetAll.
	if _, exists, err := wac.GetOneById(context.Background(), g2, g2.ID); err != nil {
		t.Fatalf("GetOneById group 2: %v", err)
	} else if exists {
		t.Fatalf("group 2 should not be visible to member A")
	}
}

// TestReentrantAccessContract runs the re-entrant access-control regression
// across every provider kind (memory, SQLite in-memory, SQLite on-disk).
func TestReentrantAccessContract(t *testing.T) {
	runContractTests(t, []contractTest{
		{name: "TestGetAllWithReentrantAccessibleTo", fn: testGetAllWithReentrantAccessibleTo},
	})
}
