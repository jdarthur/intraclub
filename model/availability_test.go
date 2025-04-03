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
	week := newStoredWeek(t, db, season)
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
	week := newStoredWeek(t, db, season)
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

func createAvailabilityForAllCaptains(t *testing.T, db common.DatabaseProvider, season *Season, weeks []*Week) []*Availability {
	// get all teams associated with this season
	teams, err := season.GetTeams(db)
	if err != nil {
		t.Fatal(err)
	}

	// save all created availability records to a list
	output := make([]*Availability, 0)

	// for each team, create an availability for its captain
	// for every week in the list, then add the availability
	// record to the output list
	for _, team := range teams {
		captain := team.Captain
		for _, week := range weeks {
			output = append(output, newStoredAvailability(t, db, captain, week.ID))
		}
	}
	return output
}

func TestGetAvailabilityForUserOnlyGetsOneUser(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	WeekCount := 4

	season, weeks := newDefaultSeasonWithWeeks(t, db, WeekCount)
	a := createAvailabilityForAllCaptains(t, db, season, weeks)

	userId := a[0].UserId
	availability, err := GetAvailabilityForUser(db, userId, season.DraftId)
	if err != nil {
		t.Fatal(err)
	}

	if len(availability) != WeekCount {
		t.Errorf("expected %d weeks, got %d", WeekCount, len(availability))
	}
	for _, a := range availability {
		if a.UserId != userId {
			t.Errorf("expected user %d, got %d", userId, a.UserId)
		}
	}
}

func TestGetAvailabilityForUserOnlyGetsOneSeason(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	WeekCount := 4

	// create a default season with 4 teams
	season, weeks := newDefaultSeasonWithWeeks(t, db, WeekCount)
	a := createAvailabilityForAllCaptains(t, db, season, weeks)

	// get the teams for this season so we can make a new Season
	teams, err := season.GetTeams(db)
	if err != nil {
		t.Fatal(err)
	}
	otherSeason, otherWeeks := newDefaultSeasonWithWeeksAndTeams(t, db, teams, WeekCount)
	_ = createAvailabilityForAllCaptains(t, db, otherSeason, otherWeeks)

	userId := a[0].UserId
	availability, err := GetAvailabilityForUser(db, userId, season.DraftId)
	if err != nil {
		t.Fatal(err)
	}

	if len(availability) != WeekCount {
		t.Errorf("expected %d weeks, got %d", WeekCount, len(availability))
	}
	for _, a2 := range availability {
		if a2.UserId != userId {
			t.Errorf("expected user %d, got %d", userId, a2.UserId)
		}

		// check if availability for each week is found
		found := false
		for _, week := range weeks {
			if a2.WeekId == week.ID {
				found = true
			}
		}
		if !found {
			t.Errorf("availability %s has week ID not found in target weeks\n", a2.ID)
			t.Errorf("Week ID: %s", a2.WeekId)
			t.Errorf("Expected weeks:\n")
			for _, week := range weeks {
				t.Errorf("\t%s\n", week.ID)
			}
			t.FailNow()
		}
	}
}

func TestMultipleAvailabilityForSingleWeekAndUserId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	availability1 := newDefaultAvailability(t, db)
	availability2 := NewAvailability()
	availability2.UserId = availability1.UserId
	availability2.WeekId = availability1.WeekId
	availability2.Available = AvailabilityAvailable

	_, err := common.CreateOne(db, availability2)
	if err == nil {
		t.Fatal("expected duplicate availability to fail")
	}
	fmt.Println(err)
}
