package model

import (
	"fmt"
	"intraclub/common"
	"testing"
)

func newStoredAvailability(t *testing.T, db common.DatabaseProvider, u UserId, week WeekId) *Availability {

	availability := NewAvailability()
	availability.UserId = u
	availability.Available = AvailabilityAvailable
	availability.WeekId = week

	v, err := common.CreateOne(db, availability)
	if err != nil {
		t.Fatal(err)
	}

	return v
}

func newDefaultAvailability(t *testing.T, db common.DatabaseProvider) *Availability {
	season := newDefaultSeason(t, db)
	userId := getAnyTeamCaptain(t, db, season)
	week := newStoredWeek(t, db, season.ID)
	v := newStoredAvailability(t, db, userId, week.ID)
	return v
}

func TestAvailabilityOnlyAccessibleToTeamMembers(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	v := newDefaultAvailability(t, db)

	otherUser := newStoredUser(t, db)
	wac := common.WithAccessControl[*Availability]{Database: db, AccessControlUser: otherUser.ID.RecordId()}
	v, exists, err := wac.GetOneById(&Availability{}, v.ID)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("expected user 2 not to be able to access availability")
	}
}

func TestAvailabilityIsAccessibleToTeamMembers(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	season := newDefaultSeason(t, db)
	userId := getAnyTeamCaptain(t, db, season)
	week := newStoredWeek(t, db, season.ID)
	v := newStoredAvailability(t, db, userId, week.ID)

	wac := common.WithAccessControl[*Availability]{Database: db, AccessControlUser: userId.RecordId()}
	v, exists, err := wac.GetOneById(&Availability{}, v.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("expected user 2 not to be able to access availability")
	}
	fmt.Printf("%T %+v\n", v, v)
}

func TestAvailabilityInvalidOption(t *testing.T) {
	v := NewAvailability()
	v.Available = AvailabilityOption(999)
	err := v.StaticallyValid()
	if err == nil {
		t.Fatal("expected invalid option to fail")
	}
	fmt.Println(err)
}

func TestAvailabilityUserDoesNotExist(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	v := NewAvailability()
	v.Available = AvailabilityAvailable

	err := v.DynamicallyValid(db)
	if err == nil {
		t.Fatal("expected invalid option to fail")
	}
	fmt.Println(err)
}

func TestAvailabilityWeekDoesNotExist(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	season := newDefaultSeason(t, db)
	userId := getAnyTeamCaptain(t, db, season)
	v := NewAvailability()
	v.UserId = userId
	err := v.DynamicallyValid(db)
	if err == nil {
		t.Fatal("expected invalid option to fail")
	}
	fmt.Println(err)
}
