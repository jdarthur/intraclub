package model

import (
	"fmt"
	"intraclub/common"
	"math/rand"
	"testing"
)

func newStoredFacility(t *testing.T, db common.DatabaseProvider, owner UserId) *Facility {
	facility := NewFacility()
	facility.UserId = owner
	facility.Name = fmt.Sprintf("Test facility %d", rand.Intn(1_000_000))
	facility.Address = fmt.Sprintf("%d Test Rd.", rand.Intn(1_000_000))
	facility.NumberOfCourts = 5
	v, err := common.CreateOne(db, facility)
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
	db := common.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	facility := newStoredFacility(t, db, user.ID)
	fmt.Printf("%+v\n", facility)

	// do CRUD via the WithAccessControl construct
	wac := common.WithAccessControl[*Facility]{Database: db, AccessControlUser: user.GetId()}

	// copy facility to a new record and update in the database
	f2 := copyFacility(facility)
	f2.Name = "New name"
	err := wac.UpdateOneById(f2)
	if err != nil {
		t.Fatal(err)
	}

	// verify that facility was updated
	v, exists, err := wac.GetOneById(facility, facility.GetId())
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
	_, exists, err = wac.DeleteOneById(facility, facility.GetId())
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("facility does not exist")
	}
}

func TestEditableBySysAdmin(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	sysAdmin := newSysAdmin(t, db)
	facility := newStoredFacility(t, db, user.ID)

	wac := common.WithAccessControl[*Facility]{Database: db, AccessControlUser: sysAdmin.GetId()}
	canEdit := wac.CanUserEdit(facility)
	if !canEdit {
		t.Fatalf("Sys admin should be able to edit facility")
	}
}

func TestNameAlreadyExists(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	facility := newStoredFacility(t, db, user.ID)

	copied := copyFacility(facility)
	copied.ID = FacilityId(common.InvalidRecordId) // generate new record ID to force a name conflict with old record
	_, err := common.CreateOne(db, copied)
	if err == nil {
		t.Fatal("expected error on duplicate name")
	}
	fmt.Println(err)
}

func TestAddressAlreadyExists(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	facility := newStoredFacility(t, db, user.ID)

	copied := copyFacility(facility)
	copied.Name = "New name"
	copied.ID = FacilityId(common.InvalidRecordId) // generate new record ID to force a name conflict with old record
	_, err := common.CreateOne(db, copied)
	if err == nil {
		t.Fatal("expected error on duplicate name")
	}
	fmt.Println(err)
}

func TestFacilityAppliedToSeasonCannotBeDeleted(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	season := newDefaultSeason(t, db)

	userId := season.Commissioners[0].RecordId()
	facilityId := season.Facility.RecordId()

	wac := common.NewWithAccessControl[*Facility](db, userId)
	_, _, err := wac.DeleteOneById(&Facility{}, facilityId)
	if err == nil {
		t.Fatal("expected error on delete")
	}
	fmt.Println(err)
}
