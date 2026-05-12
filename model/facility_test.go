package model

import (
	"context"
	"fmt"
	"testing"

	"intraclub/database"
)

func newStoredFacility(t *testing.T, db database.Provider, owner database.UserId) *Facility {
	facility := NewFacility()
	facility.UserId = owner
	facility.Name = fmt.Sprintf("Test facility %s", database.NewRecordId())
	facility.Address = fmt.Sprintf("%s Test Rd.", database.NewRecordId())
	facility.NumberOfCourts = 5
	v, err := database.CreateOne(context.Background(), db, facility)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func copyFacility(facility *Facility) *Facility {
	f := NewFacility()
	f.ID = facility.ID
	f.UserId = facility.UserId
	f.Name = facility.Name
	f.Address = facility.Address
	f.NumberOfCourts = facility.NumberOfCourts
	return f
}

func TestFacilityCrud(t *testing.T) {
	// create a database, user, and facility
	db := database.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	facility := newStoredFacility(t, db, user.ID)
	fmt.Printf("%+v\n", facility)

	// do CRUD via the WithAccessControl construct
	wac := database.WithAccessControl[*Facility]{Database: db, AccessControlUser: user.ID}

	// copy facility to a new record and update in the database
	f2 := copyFacility(facility)
	f2.Name = "New name"
	err := wac.UpdateOneById(context.Background(), f2)
	if err != nil {
		t.Fatal(err)
	}

	// verify that facility was updated
	v, exists, err := wac.GetOneById(context.Background(), facility, facility.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("facility does not exist")
	}
	if v.Name != f2.Name {
		t.Fatal("facility name does not match")
	}

	// delete facility
	_, exists, err = wac.DeleteOneById(context.Background(), facility, facility.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("facility does not exist")
	}
}

func TestEditableBySysAdmin(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	sysAdmin := newSysAdmin(t, db)
	facility := newStoredFacility(t, db, user.ID)

	wac := database.WithAccessControl[*Facility]{Database: db, AccessControlUser: sysAdmin.ID}
	canEdit := wac.CanUserEdit(facility)
	if !canEdit {
		t.Fatalf("Sys admin should be able to edit facility")
	}
}

func TestNameAlreadyExists(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	facility := newStoredFacility(t, db, user.ID)

	copied := copyFacility(facility)
	copied.ID = FacilityId(database.InvalidRecordId) // generate new record ID to force a name conflict with old record
	_, err := database.CreateOne(context.Background(), db, copied)
	if err == nil {
		t.Fatal("expected error on duplicate name")
	}
	fmt.Println(err)
}

func TestAddressAlreadyExists(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	facility := newStoredFacility(t, db, user.ID)

	copied := copyFacility(facility)
	copied.Name = "New name"
	copied.ID = FacilityId(database.InvalidRecordId) // generate new record ID to force a name conflict with old record
	_, err := database.CreateOne(context.Background(), db, copied)
	if err == nil {
		t.Fatal("expected error on duplicate address")
	}
	fmt.Println(err)
}

func TestFacilityAppliedToSeasonCannotBeDeleted(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, commish := newDefaultSeason(t, db)

	facilityId := season.Facility.RecordId()

	wac := database.NewWithAccessControl[*Facility](context.Background(), db, commish.ID)
	_, _, err := wac.DeleteOneById(context.Background(), &Facility{}, facilityId)
	if err == nil {
		t.Fatal("expected error on delete")
	}
	fmt.Println(err)
}
