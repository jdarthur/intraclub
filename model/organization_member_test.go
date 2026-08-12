package model

import (
	"context"
	"fmt"
	"testing"

	"intraclub/database"
)

func newStoredOrganizationMember(t *testing.T, db database.Provider, org *Organization, user *User) *OrganizationMember {
	m := &OrganizationMember{
		OrganizationId: org.ID,
		UserId:         user.ID,
	}
	v, err := database.CreateOne(context.Background(), db, m)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestOrganizationMemberCrud(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	org := newStoredOrganization(t, db, user.ID)

	member := newStoredOrganizationMember(t, db, org, user)
	if member.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(organization_member) did not assign an ID")
	}
	if member.OrganizationId != org.ID || member.UserId != user.ID {
		t.Fatalf("member mismatch: %+v", member)
	}

	// round-trip via GetOne
	got, exists, err := database.GetOneById(context.Background(), db, &OrganizationMember{}, member.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("organization_member not found")
	}
	if got.OrganizationId != member.OrganizationId || got.UserId != member.UserId {
		t.Fatalf("organization_member round-trip mismatch:\n  got  %+v\n  want %+v", got, member)
	}
}

func TestOrganizationMemberDuplicateGuard(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	org := newStoredOrganization(t, db, user.ID)

	newStoredOrganizationMember(t, db, org, user)

	// Attempt to add the same (organization, user) pair again.
	dup := &OrganizationMember{
		OrganizationId: org.ID,
		UserId:         user.ID,
	}
	_, err := database.CreateOne(context.Background(), db, dup)
	if err == nil {
		t.Fatal("expected error on duplicate organization membership")
	}
	fmt.Println(err)
}

func TestOrganizationMemberRequiresExistingRecords(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	org := newStoredOrganization(t, db, user.ID)

	// Referencing a non-existent user should fail dynamic validation.
	bad := &OrganizationMember{
		OrganizationId: org.ID,
		UserId:         database.InvalidUserId,
	}
	if err := bad.DynamicallyValid(context.Background(), db); err == nil {
		t.Fatal("expected error when referenced user does not exist")
	}

	// Referencing a non-existent organization should fail dynamic validation.
	bad2 := &OrganizationMember{
		OrganizationId: OrganizationId(database.InvalidRecordId),
		UserId:         user.ID,
	}
	if err := bad2.DynamicallyValid(context.Background(), db); err == nil {
		t.Fatal("expected error when referenced organization does not exist")
	}
}
