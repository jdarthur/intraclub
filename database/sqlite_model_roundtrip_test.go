package database_test

import (
	"context"
	"path/filepath"
	"testing"

	"intraclub/database"
	"intraclub/model"
)

// newSqliteProvider opens a fresh on-disk SQLite database (running all
// embedded migrations, including the skeleton and the #55 users/accounts
// migrations) and registers cleanup to close it.
func newSqliteProvider(t *testing.T) database.Provider {
	t.Helper()
	p, err := database.NewSqliteDbProvider(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("NewSqliteDbProvider: %v", err)
	}
	t.Cleanup(func() {
		if d, ok := p.(interface{ Disconnect() error }); ok {
			d.Disconnect()
		}
	})
	return p
}

// TestUserRoundTrip verifies field-by-field losslessness for the User model
// across Create -> GetOne/GetAll -> Update -> Delete on the SQLite provider.
func TestUserRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	created, err := database.CreateOne(ctx, p, &model.User{
		FirstName:   "Alice",
		LastName:    "Smith",
		PhoneNumber: "7705551234",
		Email:       "Alice@Example.COM",
		Verified:    true,
	})
	if err != nil {
		t.Fatalf("CreateOne(user): %v", err)
	}
	u := created
	if u.ID == database.InvalidUserId {
		t.Fatal("CreateOne(user) did not assign a UserId")
	}

	// GetOne round-trips every field.
	got, exists, err := database.GetOneById(ctx, p, &model.User{}, u.ID.RecordId())
	if err != nil {
		t.Fatalf("GetOneById(user): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(user): record not found")
	}
	if got.FirstName != u.FirstName || got.LastName != u.LastName ||
		got.PhoneNumber != u.PhoneNumber || got.Email != u.Email ||
		got.Verified != u.Verified {
		t.Fatalf("user round-trip mismatch:\n  got  %+v\n  want %+v", got, u)
	}

	// GetAll returns the record with matching fields.
	all, err := database.GetAll[*model.User](ctx, p)
	if err != nil {
		t.Fatalf("GetAll(user): %v", err)
	}
	if len(all) != 1 || all[0].ID != u.ID {
		t.Fatalf("GetAll(user): got %d records, want 1 matching the created user", len(all))
	}

	// Update persists changes.
	u.LastName = "Jones"
	u.Verified = false
	if err := database.UpdateOne(ctx, p, u); err != nil {
		t.Fatalf("UpdateOne(user): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.User{}, u.ID.RecordId())
	if err != nil {
		t.Fatalf("GetOneById(user) after update: %v", err)
	}
	if got2.LastName != "Jones" || got2.Verified {
		t.Fatalf("user update not persisted, got %+v", got2)
	}

	// Delete removes the record.
	if _, _, err := database.DeleteOneById(ctx, p, &model.User{}, u.ID.RecordId()); err != nil {
		t.Fatalf("DeleteOneById(user): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.User{}, u.ID.RecordId())
	if err != nil {
		t.Fatalf("GetOneById(user) after delete: %v", err)
	}
	if exists {
		t.Fatal("user should have been deleted")
	}
}

// TestUserRoleAssignmentRoundTrip verifies field-by-field losslessness for the
// UserRoleAssignment model, using a SystemAdministrator role (no referenced
// record needed for dynamic validation).
func TestUserRoleAssignmentRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	user, err := database.CreateOne(ctx, p, &model.User{
		FirstName:   "Bob",
		LastName:    "Jones",
		PhoneNumber: "4045551234",
		Email:       "bob@example.com",
	})
	if err != nil {
		t.Fatalf("CreateOne(user): %v", err)
	}

	created, err := database.CreateOne(ctx, p, &model.UserRoleAssignment{
		UserId:      user.ID,
		Role:        model.SystemAdministrator,
		ReferenceId: database.InvalidRecordId,
	})
	if err != nil {
		t.Fatalf("CreateOne(assignment): %v", err)
	}
	a := created
	if a.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(assignment) did not assign an ID")
	}

	// GetOne round-trips every field.
	got, exists, err := database.GetOneById(ctx, p, &model.UserRoleAssignment{}, a.GetId())
	if err != nil {
		t.Fatalf("GetOneById(assignment): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(assignment): record not found")
	}
	if got.UserId != a.UserId || got.Role != a.Role || got.ReferenceId != a.ReferenceId {
		t.Fatalf("assignment round-trip mismatch:\n  got  %+v\n  want %+v", got, a)
	}

	// Update persists a field change (reference id, unvalidated for sysadmin).
	refId := database.NewRecordId()
	a.ReferenceId = refId
	if err := database.UpdateOne(ctx, p, a); err != nil {
		t.Fatalf("UpdateOne(assignment): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.UserRoleAssignment{}, a.GetId())
	if err != nil {
		t.Fatalf("GetOneById(assignment) after update: %v", err)
	}
	if got2.ReferenceId != refId {
		t.Fatalf("assignment reference_id not persisted, got %+v", got2)
	}

	// Delete removes the record.
	if _, _, err := database.DeleteOneById(ctx, p, &model.UserRoleAssignment{}, a.GetId()); err != nil {
		t.Fatalf("DeleteOneById(assignment): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.UserRoleAssignment{}, a.GetId())
	if err != nil {
		t.Fatalf("GetOneById(assignment) after delete: %v", err)
	}
	if exists {
		t.Fatal("assignment should have been deleted")
	}
}

// TestEmailTokenRoundTrip verifies field-by-field losslessness for the
// EmailToken model.
func TestEmailTokenRoundTrip(t *testing.T) {
	ctx := context.Background()
	p := newSqliteProvider(t)

	user, err := database.CreateOne(ctx, p, &model.User{
		FirstName:   "Carol",
		LastName:    "Davis",
		PhoneNumber: "6785551234",
		Email:       "carol@example.com",
	})
	if err != nil {
		t.Fatalf("CreateOne(user): %v", err)
	}

	created, err := database.CreateOne(ctx, p, &model.EmailToken{
		UserId: user.ID,
		Token:  "a1b2c3d4e5f6",
	})
	if err != nil {
		t.Fatalf("CreateOne(token): %v", err)
	}
	tok := created
	if tok.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(token) did not assign an ID")
	}

	// GetOne round-trips every field.
	got, exists, err := database.GetOneById(ctx, p, &model.EmailToken{}, tok.GetId())
	if err != nil {
		t.Fatalf("GetOneById(token): %v", err)
	}
	if !exists {
		t.Fatal("GetOneById(token): record not found")
	}
	if got.Token != tok.Token || got.UserId != tok.UserId {
		t.Fatalf("token round-trip mismatch:\n  got  %+v\n  want %+v", got, tok)
	}

	// Update persists a token change.
	tok.Token = "f6e5d4c3b2a1"
	if err := database.UpdateOne(ctx, p, tok); err != nil {
		t.Fatalf("UpdateOne(token): %v", err)
	}
	got2, _, err := database.GetOneById(ctx, p, &model.EmailToken{}, tok.GetId())
	if err != nil {
		t.Fatalf("GetOneById(token) after update: %v", err)
	}
	if got2.Token != "f6e5d4c3b2a1" {
		t.Fatalf("token update not persisted, got %+v", got2)
	}

	// Delete removes the record.
	if _, _, err := database.DeleteOneById(ctx, p, &model.EmailToken{}, tok.GetId()); err != nil {
		t.Fatalf("DeleteOneById(token): %v", err)
	}
	_, exists, err = database.GetOneById(ctx, p, &model.EmailToken{}, tok.GetId())
	if err != nil {
		t.Fatalf("GetOneById(token) after delete: %v", err)
	}
	if exists {
		t.Fatal("token should have been deleted")
	}
}
