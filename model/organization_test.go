package model

import (
	"context"
	"fmt"
	"testing"

	"intraclub/database"
)

func newStoredOrganization(t *testing.T, db database.Provider, owner database.UserId) *Organization {
	org := NewOrganization()
	org.UserId = owner
	org.Name = fmt.Sprintf("Test organization %s", database.NewRecordId())
	v, err := database.CreateOne(context.Background(), db, org)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func copyOrganization(org *Organization) *Organization {
	o := NewOrganization()
	o.ID = org.ID
	o.UserId = org.UserId
	o.Name = org.Name
	return o
}

func TestOrganizationCrud(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	org := newStoredOrganization(t, db, user.ID)
	if org.GetId() == database.InvalidRecordId {
		t.Fatal("CreateOne(organization) did not assign an ID")
	}
	if org.GetOwner() != user.ID {
		t.Fatalf("organization owner = %v, want %v", org.GetOwner(), user.ID)
	}

	// do CRUD via the WithAccessControl construct
	wac := database.WithAccessControl[*Organization]{Database: db, AccessControlUser: user.ID}

	// copy organization to a new record and update in the database
	o2 := copyOrganization(org)
	o2.Name = "New name"
	err := wac.UpdateOneById(context.Background(), o2)
	if err != nil {
		t.Fatal(err)
	}

	// verify that organization was updated
	v, exists, err := wac.GetOneById(context.Background(), org, org.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("organization does not exist")
	}
	if v.Name != o2.Name {
		t.Fatal("organization name does not match")
	}

	// delete organization
	_, exists, err = wac.DeleteOneById(context.Background(), org, org.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("organization does not exist")
	}
}

func TestOrganizationEditableByOwnerAndSysAdmin(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	sysAdmin := newSysAdmin(t, db)
	org := newStoredOrganization(t, db, user.ID)

	ownerWac := database.WithAccessControl[*Organization]{Database: db, AccessControlUser: user.ID}
	if !ownerWac.CanUserEdit(org) {
		t.Fatal("owner should be able to edit organization")
	}

	sysWac := database.WithAccessControl[*Organization]{Database: db, AccessControlUser: sysAdmin.ID}
	if !sysWac.CanUserEdit(org) {
		t.Fatal("sys admin should be able to edit organization")
	}
}

func TestOrganizationNameAlreadyExists(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	org := newStoredOrganization(t, db, user.ID)

	copied := copyOrganization(org)
	copied.ID = OrganizationId(database.InvalidRecordId) // force a name conflict with old record
	_, err := database.CreateOne(context.Background(), db, copied)
	if err == nil {
		t.Fatal("expected error on duplicate name")
	}
}

func TestOrganizationEmptyNameInvalid(t *testing.T) {
	org := NewOrganization()
	org.UserId = database.InvalidUserId
	org.Name = "   "
	if err := org.StaticallyValid(); err == nil {
		t.Fatal("expected error for blank organization name")
	}
}
