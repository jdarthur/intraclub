package model

import (
	"context"
	"fmt"
	"testing"

	"intraclub/database"
)

func newDefaultSeason(t *testing.T, db database.Provider) (s *Season, commish *User) {
	return newDefaultSeasonWithTeams(t, db, 1)
}

func newDefaultSeasonWithTeams(t *testing.T, db database.Provider, teamCount int) (s *Season, commish *User) {
	commissioner := newStoredUser(t, db)

	teams := make([]*Team, 0)
	for i := 0; i < teamCount; i++ {
		teamCaptain := newStoredUser(t, db)
		teams = append(teams, newStoredTeam(t, db, teamCaptain.ID))
	}

	return newStoredSeason(t, db, commissioner.ID, teams), commissioner
}

func newStoredSeason(t *testing.T, db database.Provider, commissioner database.UserId, teams []*Team) *Season {
	draft := newStoredDraft(t, db, commissioner)
	facility := newStoredFacility(t, db, commissioner)
	playoffStructure := newStoredPlayoffStructure(t, db)

	season := NewSeason()
	season.Name = "Test Season"
	season.StartTime = NewStartTime(8, 30)
	season.DraftId = draft.ID
	season.Facility = facility.ID
	season.PlayoffStructure = playoffStructure.ID

	v, err := database.CreateOne(context.Background(), db, season)
	if err != nil {
		t.Fatal(err)
	}

	for _, team := range teams {
		err := season.AddTeam(context.Background(), db, team.ID)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = season.AddCommissioner(context.Background(), db, commissioner)
	if err != nil {
		t.Fatal(err)
	}

	return v
}

func TestCreateSeasonAfterDraft(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	draft := doRandomDraft(t, db, 100, 4)
	facility := newStoredFacility(t, db, draft.Owner)
	season, err := draft.CreateSeason(context.Background(), db, "Test season", facility.ID, NewStartTime(8, 30))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", season)
}

func TestSeasonPostDeleteCascadesChildren(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, _ := newDefaultSeasonWithTeams(t, db, 4)

	lateUser := newStoredUser(t, db)
	_, err := database.CreateOne(context.Background(), db, &SeasonLateAddition{SeasonId: season.ID, UserId: lateUser.ID})
	if err != nil {
		t.Fatal(err)
	}

	counts := func() (int, int, int) {
		commissioners, err := database.GetAllWhere[*SeasonCommissioner](context.Background(), db, func(_ context.Context, r *SeasonCommissioner) bool {
			return r.SeasonId == season.ID
		})
		if err != nil {
			t.Fatal(err)
		}
		late, err := database.GetAllWhere[*SeasonLateAddition](context.Background(), db, func(_ context.Context, r *SeasonLateAddition) bool {
			return r.SeasonId == season.ID
		})
		if err != nil {
			t.Fatal(err)
		}
		teams, err := database.GetAllWhere[*SeasonTeam](context.Background(), db, func(_ context.Context, r *SeasonTeam) bool {
			return r.SeasonId == season.ID
		})
		if err != nil {
			t.Fatal(err)
		}
		return len(commissioners), len(late), len(teams)
	}

	commCount, lateCount, teamCount := counts()
	if commCount == 0 || teamCount == 0 {
		t.Fatalf("expected season to have commissioners and teams, got commissioners=%d teams=%d", commCount, teamCount)
	}
	if lateCount != 1 {
		t.Fatalf("expected 1 late addition, got %d", lateCount)
	}

	_, _, err = database.DeleteOneById(context.Background(), db, &Season{}, season.ID.RecordId())
	if err != nil {
		t.Fatal(err)
	}

	commCount, lateCount, teamCount = counts()
	if commCount != 0 || lateCount != 0 || teamCount != 0 {
		t.Fatalf("expected 0 after delete, got commissioners=%d late=%d teams=%d", commCount, lateCount, teamCount)
	}
}
