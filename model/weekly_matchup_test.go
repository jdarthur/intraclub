package model

import (
	"context"
	"fmt"
	"testing"

	"intraclub/database"
)

func newStoredWeeklyMatchup(t *testing.T, db database.DatabaseProvider) (*Season, *WeeklyMatchup) {
	season, _ := newDefaultSeasonWithTeams(t, db, 4)
	week := newStoredWeek(t, db, season)

	teams, err := season.GetTeams(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	w := NewWeeklyMatchup()
	w.WeekId = week.ID
	w.SeasonId = season.ID

	matchup := TeamMatchup{
		HomeTeam: teams[0].ID,
		AwayTeam: teams[1].ID,
	}
	matchup2 := TeamMatchup{
		HomeTeam: teams[2].ID,
		AwayTeam: teams[3].ID,
	}

	w.Matchups = []*TeamMatchup{&matchup, &matchup2}

	v, err := database.CreateOne(context.Background(), db, w)
	if err != nil {
		t.Fatal(err)
	}
	return season, v
}

func copyWeeklyMatchup(w *WeeklyMatchup) *WeeklyMatchup {
	return &WeeklyMatchup{
		WeekId:   w.WeekId,
		SeasonId: w.SeasonId,
		Matchups: w.Matchups,
	}
}

func TestWeeklyMatchupInvalidHomeTeamId(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)
	w.Matchups[0].HomeTeam = TeamId(database.InvalidRecordId)
	err := w.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("InvalidScoreCountingType home team ID should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupInvalidAwayTeamId(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)
	w.Matchups[0].AwayTeam = TeamId(database.InvalidRecordId)
	err := w.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("InvalidScoreCountingType away team ID should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupInvalidSeasonId(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)
	w.SeasonId = SeasonId(database.InvalidRecordId)
	err := w.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("InvalidScoreCountingType season ID should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupInvalidWeekId(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)
	w.WeekId = WeekId(database.InvalidRecordId)
	err := w.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("InvalidScoreCountingType week ID should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupWeekDoesNotBelongToSeason(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	otherSeason, _ := newDefaultSeason(t, db)
	someOtherWeek := newStoredWeek(t, db, otherSeason)

	w.WeekId = someOtherWeek.ID
	err := w.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("InvalidScoreCountingType week ID should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupHomeTeamDoesNotBelongToSeason(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	otherTeam := newStoredTeam(t, db, newStoredUser(t, db).ID)
	w.Matchups[0].HomeTeam = otherTeam.ID

	err := w.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("Team from another season should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupAwayTeamDoesNotBelongToSeason(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	otherTeam := newStoredTeam(t, db, newStoredUser(t, db).ID)
	w.Matchups[0].AwayTeam = otherTeam.ID

	err := w.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("Team from another season should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupTeamPlayingInMultipleMatchups(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, w := newStoredWeeklyMatchup(t, db)

	teams, err := season.GetTeams(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	anotherMatchup := TeamMatchup{
		HomeTeam: teams[0].ID,
		AwayTeam: teams[2].ID,
	}
	w.Matchups = append(w.Matchups, &anotherMatchup)

	err = w.StaticallyValid()
	if err == nil {
		t.Fatal("Double matchup for team 1 should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupTeamDoesNotHaveAnyMatchups(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	w.Matchups[1].Bye = true
	w.Matchups[1].AwayTeam = TeamId(database.InvalidRecordId)

	err := database.Validate(context.Background(), db, w)
	if err == nil {
		t.Fatal("Team 4 without matchup or bye should produce error")
	}
	fmt.Println(err)

}

func TestWeeklyMatchupTeamHasBye(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, _ := newDefaultSeasonWithTeams(t, db, 4)
	week := newStoredWeek(t, db, season)

	teams, err := season.GetTeams(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	w := NewWeeklyMatchup()
	w.WeekId = week.ID
	w.SeasonId = season.ID

	matchup := TeamMatchup{
		HomeTeam: teams[0].ID,
		AwayTeam: teams[1].ID,
	}
	bye1 := TeamMatchup{
		HomeTeam: teams[2].ID,
		Bye:      true,
	}
	bye2 := TeamMatchup{
		HomeTeam: teams[3].ID,
		Bye:      true,
	}

	w.Matchups = []*TeamMatchup{&matchup, &bye1, &bye2}

	err = database.Validate(context.Background(), db, w)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWeeklyMatchupHomeTeamByeButAwayTeamIsSet(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	w.Matchups[0].Bye = true
	err := database.Validate(context.Background(), db, w)
	if err == nil {
		t.Fatal("Bye with away team set should produce error")
	}
	fmt.Println(err)
}

func TestDuplicateWeeklyMatchup(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	w2 := copyWeeklyMatchup(w)
	_, err := database.CreateOne(context.Background(), db, w2)
	if err == nil {
		t.Fatal("Expected error on duplicate weekly matchup")
	}
	fmt.Println(err)
}
