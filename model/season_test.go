package model

import (
	"intraclub/common"
	"testing"
)

func newDefaultSeason(t *testing.T, db common.DatabaseProvider) *Season {
	commissioner := newStoredUser(t, db)
	team := newStoredTeam(t, db, commissioner.ID)
	return newStoredSeason(t, db, commissioner.ID, []*Team{team})
}

func newStoredSeason(t *testing.T, db common.DatabaseProvider, commissioner UserId, teams []*Team) *Season {
	draft := newStoredDraft(t, db, commissioner)
	facility := newStoredFacility(t, db, commissioner)

	season := NewSeason()
	season.Name = "Test Season"
	season.Commissioners = []UserId{commissioner}
	season.StartTime = NewStartTime(8, 30)
	season.DraftId = draft.ID
	season.Facility = facility.ID
	for _, team := range teams {
		season.Teams = append(season.Teams, team.ID)
	}

	v, err := common.CreateOne(db, season)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
