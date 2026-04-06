package model

import (
	"fmt"
	"testing"

	"intraclub/common"
)

func newDefaultSeason(t *testing.T, db common.DatabaseProvider) (s *Season, commish *User) {
	return newDefaultSeasonWithTeams(t, db, 1)
}

func newDefaultSeasonWithTeams(t *testing.T, db common.DatabaseProvider, teamCount int) (s *Season, commish *User) {
	commissioner := newStoredUser(t, db)

	teams := make([]*Team, 0)
	for i := 0; i < teamCount; i++ {
		teamCaptain := newStoredUser(t, db)
		teams = append(teams, newStoredTeam(t, db, teamCaptain.ID))
	}

	return newStoredSeason(t, db, commissioner.ID, teams), commissioner
}

func newStoredSeason(t *testing.T, db common.DatabaseProvider, commissioner UserId, teams []*Team) *Season {
	draft := newStoredDraft(t, db, commissioner)
	facility := newStoredFacility(t, db, commissioner)
	playoffStructure := newStoredPlayoffStructure(t, db)

	season := NewSeason()
	season.Name = "Test Season"
	season.StartTime = NewStartTime(8, 30)
	season.DraftId = draft.ID
	season.Facility = facility.ID
	season.PlayoffStructure = playoffStructure.ID

	v, err := common.CreateOne(db, season)
	if err != nil {
		t.Fatal(err)
	}

	for _, team := range teams {
		err := season.AddTeam(db, team.ID)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = season.AddCommissioner(db, commissioner)
	if err != nil {
		t.Fatal(err)
	}

	return v
}

func TestCreateSeasonAfterDraft(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft := doRandomDraft(t, db, 100, 4)
	facility := newStoredFacility(t, db, draft.Owner)
	season, err := draft.CreateSeason(db, "Test season", facility.ID, NewStartTime(8, 30))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", season)
}
