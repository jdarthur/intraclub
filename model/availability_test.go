package model

import (
	"fmt"
	"intraclub/common"
	"testing"
)

func newStoredAvailability(t *testing.T, db common.DatabaseProvider, u UserId, team TeamId, week WeekId) *Availability {

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

func TestAvailabilityOnlyAccessibleToTeamMembers(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	teamCaptain := newStoredUser(t, db)
	team := newStoredTeam(t, db, teamCaptain.ID)
	season := newStoredSeason(t, db, teamCaptain.ID, []*Team{team})
	week := newStoredWeek(t, db, season.ID)
	v := newStoredAvailability(t, db, teamCaptain.ID, team.ID, week.ID)

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
	teamCaptain := newStoredUser(t, db)
	team := newStoredTeam(t, db, teamCaptain.ID)
	season := newStoredSeason(t, db, teamCaptain.ID, []*Team{team})
	week := newStoredWeek(t, db, season.ID)
	v := newStoredAvailability(t, db, teamCaptain.ID, team.ID, week.ID)

	wac := common.WithAccessControl[*Availability]{Database: db, AccessControlUser: teamCaptain.ID.RecordId()}
	v, exists, err := wac.GetOneById(&Availability{}, v.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("expected user 2 not to be able to access availability")
	}
	fmt.Printf("%T %+v\n", v, v)
}
