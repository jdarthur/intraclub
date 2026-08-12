package database_test

import (
	"context"
	"testing"

	"intraclub/database"
	"intraclub/model"
)

// TestOrganizationRoundTrip verifies field-by-field losslessness for the
// Organization model across Create -> GetOne/GetAll -> Update -> Delete on the
// SQLite provider, confirming the `organization` table exists post-migration.
func TestOrganizationRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	owner, err := database.CreateOne(ctx, p, &model.User{
		FirstName:   "Dana",
		LastName:    "Lee",
		PhoneNumber: "7705552222",
		Email:       "dana@example.com",
	})
	if err != nil {
		t.Fatalf("CreateOne(user): %v", err)
	}

	created, err := database.CreateOne(ctx, p, &model.Organization{
		UserId: owner.ID,
		Name:   "Martin's Landing Men",
	})
	if err != nil {
		t.Fatalf("CreateOne(organization): %v", err)
	}
	org := created
	if org.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(organization) did not assign an OrganizationId")
	}
	if org.GetOwner() != owner.ID {
		t.Fatalf("organization owner = %v, want %v", org.GetOwner(), owner.ID)
	}

	// GetOne round-trips every field.
	got, exists, err := database.GetOneById(ctx, p, &model.Organization{}, org.GetId())
	if err != nil {
		t.Fatalf("GetOneById(organization): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(organization): record not found")
	}
	if got.Name != org.Name || got.UserId != org.UserId {
		t.Fatalf("organization round-trip mismatch:\n  got  %+v\n  want %+v", got, org)
	}

	// GetAll returns the record with matching fields.
	all, err := database.GetAll[*model.Organization](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(organization): %v", err)
	}
	if len(all) != 1 || all[0].ID != org.ID {
		t.Fatalf("GetAll(organization): got %d records, want 1 matching the created org", len(all))
	}

	// Update persists changes.
	org.Name = "Martin's Landing Ladies"
	if err := database.UpdateOne(ctx, p, org); err != nil {
		t.Fatalf("UpdateOne(organization): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.Organization{}, org.GetId())
	if err != nil {
		t.Fatalf("GetOneById(organization) after update: %v", err)
	}
	if got2.Name != "Martin's Landing Ladies" {
		t.Fatalf("organization update not persisted, got %+v", got2)
	}

	// Delete removes the record.
	if _, _, err := database.DeleteOneById(ctx, p, &model.Organization{}, org.GetId()); err != nil {
		t.Fatalf("DeleteOneById(organization): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.Organization{}, org.GetId())
	if err != nil {
		t.Fatalf("GetOneById(organization) after delete: %v", err)
	}
	if exists {
		t.Fatal("organization should have been deleted")
	}
}

// TestOrganizationMemberRoundTrip verifies that the `organization_member`
// join table exists post-migration and round-trips through the SQLite provider.
func TestOrganizationMemberRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	owner, err := database.CreateOne(ctx, p, &model.User{
		FirstName:   "Erin",
		LastName:    "Ng",
		PhoneNumber: "6785553333",
		Email:       "erin@example.com",
	})
	if err != nil {
		t.Fatalf("CreateOne(owner): %v", err)
	}
	member, err := database.CreateOne(ctx, p, &model.User{
		FirstName:   "Frank",
		LastName:    "Wu",
		PhoneNumber: "6785554444",
		Email:       "frank@example.com",
	})
	if err != nil {
		t.Fatalf("CreateOne(member): %v", err)
	}
	org, err := database.CreateOne(ctx, p, &model.Organization{
		UserId: owner.ID,
		Name:   "Riverside Racquets",
	})
	if err != nil {
		t.Fatalf("CreateOne(organization): %v", err)
	}

	created, err := database.CreateOne(ctx, p, &model.OrganizationMember{
		OrganizationId: org.ID,
		UserId:         member.ID,
	})
	if err != nil {
		t.Fatalf("CreateOne(organization_member): %v", err)
	}
	m := created
	if m.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(organization_member) did not assign an ID")
	}

	got, exists, err := database.GetOneById(ctx, p, &model.OrganizationMember{}, m.GetId())
	if err != nil {
		t.Fatalf("GetOneById(organization_member): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(organization_member): record not found")
	}
	if got.OrganizationId != m.OrganizationId || got.UserId != m.UserId {
		t.Fatalf("organization_member round-trip mismatch:\n  got  %+v\n  want %+v", got, m)
	}

	// Duplicate membership must be rejected through the provider.
	if _, err := database.CreateOne(ctx, p, &model.OrganizationMember{
		OrganizationId: org.ID,
		UserId:         member.ID,
	}); err == nil {
		t.Fatal("expected error on duplicate organization membership")
	}

	// Delete removes the record.
	if _, _, err := database.DeleteOneById(ctx, p, &model.OrganizationMember{}, m.GetId()); err != nil {
		t.Fatalf("DeleteOneById(organization_member): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.OrganizationMember{}, m.GetId())
	if err != nil {
		t.Fatalf("GetOneById(organization_member) after delete: %v", err)
	}
	if exists {
		t.Fatal("organization_member should have been deleted")
	}
}
