package model

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"intraclub/database"
)

func newStoredAvailability(t *testing.T, db database.Provider, u database.UserId, week WeekId) *Availability {

	availability := NewAvailability()
	availability.UserId = u
	availability.Available = AvailabilityAvailable
	availability.WeekId = week

	v, err := database.CreateOne(context.Background(), db, availability)
	require.NoError(t, err)

	return v
}

func newDefaultAvailability(t *testing.T, db database.Provider) *Availability {
	season, _ := newDefaultSeason(t, db)
	userId := getAnyTeamCaptain(t, db, season)
	week := newStoredWeek(t, db, season)
	v := newStoredAvailability(t, db, userId, week.ID)
	return v
}

func TestAvailabilityOnlyAccessibleToTeamMembers(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	v := newDefaultAvailability(t, db)

	otherUser := newStoredUser(t, db)
	wac := database.WithAccessControl[*Availability]{Database: db, AccessControlUser: otherUser.ID}
	_, exists, err := wac.GetOneById(context.Background(), &Availability{}, v.ID)
	require.NoError(t, err)
	assert.False(t, exists, "expected user 2 not to be able to access availability")
}

func TestAvailabilityIsAccessibleToTeamMembers(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, _ := newDefaultSeason(t, db)
	userId := getAnyTeamCaptain(t, db, season)
	week := newStoredWeek(t, db, season)
	v := newStoredAvailability(t, db, userId, week.ID)

	wac := database.WithAccessControl[*Availability]{Database: db, AccessControlUser: userId}
	v, exists, err := wac.GetOneById(context.Background(), &Availability{}, v.ID)
	require.NoError(t, err)
	require.True(t, exists, "expected user to be able to access availability")
	fmt.Printf("%T %+v\n", v, v)
}

func TestAvailabilityInvalidOption(t *testing.T) {
	v := NewAvailability()
	v.Available = AvailabilityOption(999)
	assert.Error(t, v.StaticallyValid(), "expected invalid option to fail")
	fmt.Println(v.StaticallyValid())
}

func TestAvailabilityUserDoesNotExist(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	v := NewAvailability()
	v.Available = AvailabilityAvailable

	assert.Error(t, v.DynamicallyValid(context.Background(), db), "expected invalid option to fail")
	fmt.Println(v.DynamicallyValid(context.Background(), db))
}

func TestAvailabilityWeekDoesNotExist(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, _ := newDefaultSeason(t, db)
	userId := getAnyTeamCaptain(t, db, season)
	v := NewAvailability()
	v.UserId = userId
	assert.Error(t, v.DynamicallyValid(context.Background(), db), "expected invalid option to fail")
	fmt.Println(v.DynamicallyValid(context.Background(), db))
}

func createAvailabilityForAllCaptains(t *testing.T, db database.Provider, season *Season, weeks []*Week) []*Availability {
	// get all teams associated with this season
	teams, err := season.GetTeams(context.Background(), db)
	require.NoError(t, err)

	// save all created availability records to a list
	output := make([]*Availability, 0)

	// for each team, create an availability for its captain
	// for every week in the list, then add the availability
	// record to the output list
	for _, team := range teams {
		captain, err := team.GetCaptain(context.Background(), db)
		require.NoError(t, err)
		for _, week := range weeks {
			output = append(output, newStoredAvailability(t, db, captain, week.ID))
		}
	}
	return output
}

func TestGetAvailabilityForUserOnlyGetsOneUser(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	weekCount := 4

	season, weeks := newDefaultSeasonWithWeeks(t, db, weekCount)
	a := createAvailabilityForAllCaptains(t, db, season, weeks)

	userId := a[0].UserId
	availability, err := GetAvailabilityForUser(context.Background(), db, userId, season.DraftId)
	require.NoError(t, err)

	assert.Len(t, availability, weekCount, "expected %d weeks, got %d", weekCount, len(availability))
	for _, a := range availability {
		assert.Equal(t, userId, a.UserId, "expected user %d, got %d", userId, a.UserId)
	}
}

func TestGetAvailabilityForUserOnlyGetsOneSeason(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	weekCount := 4

	// create a default season with 4 teams
	season, weeks := newDefaultSeasonWithWeeks(t, db, weekCount)
	a := createAvailabilityForAllCaptains(t, db, season, weeks)

	// get the teams for this season so we can make a new Season
	teams, err := season.GetTeams(context.Background(), db)
	require.NoError(t, err)
	otherSeason, otherWeeks := newDefaultSeasonWithWeeksAndTeams(t, db, teams, weekCount)
	_ = createAvailabilityForAllCaptains(t, db, otherSeason, otherWeeks)

	userId := a[0].UserId
	availability, err := GetAvailabilityForUser(context.Background(), db, userId, season.DraftId)
	require.NoError(t, err)

	assert.Len(t, availability, weekCount, "expected %d weeks, got %d", weekCount, len(availability))
	for _, a2 := range availability {
		assert.Equal(t, userId, a2.UserId, "expected user %d, got %d", userId, a2.UserId)

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
	db := database.NewUnitTestDBProvider()
	availability1 := newDefaultAvailability(t, db)
	availability2 := NewAvailability()
	availability2.UserId = availability1.UserId
	availability2.WeekId = availability1.WeekId
	availability2.Available = AvailabilityAvailable

	_, err := database.CreateOne(context.Background(), db, availability2)
	assert.Error(t, err, "expected duplicate availability to fail")
	fmt.Println(err)
}
