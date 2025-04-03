package model

import (
	"fmt"
	"intraclub/common"
	"testing"
)

func newStoredWeeklyMatchup(t *testing.T, db common.DatabaseProvider) (*Season, *WeeklyMatchup) {
	season := newDefaultSeasonWithTeams(t, db, 4)
	week := newStoredWeek(t, db, season)

	w := NewWeeklyMatchup()
	w.WeekId = week.ID
	w.SeasonId = season.ID

	matchup := TeamMatchup{
		HomeTeam: season.Teams[0],
		AwayTeam: season.Teams[1],
	}
	matchup2 := TeamMatchup{
		HomeTeam: season.Teams[2],
		AwayTeam: season.Teams[3],
	}

	w.Matchups = []*TeamMatchup{&matchup, &matchup2}

	v, err := common.CreateOne(db, w)
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
	db := common.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)
	w.Matchups[0].HomeTeam = TeamId(common.InvalidRecordId)
	err := w.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Invalid home team ID should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupInvalidAwayTeamId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)
	w.Matchups[0].AwayTeam = TeamId(common.InvalidRecordId)
	err := w.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Invalid away team ID should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupInvalidSeasonId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)
	w.SeasonId = SeasonId(common.InvalidRecordId)
	err := w.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Invalid season ID should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupInvalidWeekId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)
	w.WeekId = WeekId(common.InvalidRecordId)
	err := w.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Invalid week ID should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupWeekDoesNotBelongToSeason(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	otherSeason := newDefaultSeason(t, db)
	someOtherWeek := newStoredWeek(t, db, otherSeason)

	w.WeekId = someOtherWeek.ID
	err := w.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Invalid week ID should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupHomeTeamDoesNotBelongToSeason(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	otherTeam := newStoredTeam(t, db, newStoredUser(t, db).ID)
	w.Matchups[0].HomeTeam = otherTeam.ID

	err := w.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Team from another season should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupAwayTeamDoesNotBelongToSeason(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	otherTeam := newStoredTeam(t, db, newStoredUser(t, db).ID)
	w.Matchups[0].AwayTeam = otherTeam.ID

	err := w.DynamicallyValid(db)
	if err == nil {
		t.Fatal("Team from another season should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupTeamPlayingInMultipleMatchups(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	season, w := newStoredWeeklyMatchup(t, db)

	anotherMatchup := TeamMatchup{
		HomeTeam: season.Teams[0],
		AwayTeam: season.Teams[2],
	}
	w.Matchups = append(w.Matchups, &anotherMatchup)

	err := w.StaticallyValid()
	if err == nil {
		t.Fatal("Double matchup for team 1 should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupTeamDoesNotHaveAnyMatchups(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	w.Matchups[1].Bye = true
	w.Matchups[1].AwayTeam = TeamId(common.InvalidRecordId)

	err := common.Validate(db, w)
	if err == nil {
		t.Fatal("Team 4 without matchup or bye should produce error")
	}
	fmt.Println(err)

}

func TestWeeklyMatchupTeamHasBye(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	season, w := newStoredWeeklyMatchup(t, db)

	w.Matchups[1].Bye = true
	w.Matchups[1].AwayTeam = TeamId(common.InvalidRecordId)
	anotherBye := TeamMatchup{HomeTeam: season.Teams[3], Bye: true}
	w.Matchups = append(w.Matchups, &anotherBye)

	err := common.Validate(db, w)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWeeklyMatchupHomeTeamByeButAwayTeamIsSet(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	w.Matchups[0].Bye = true
	err := common.Validate(db, w)
	if err == nil {
		t.Fatal("Bye with away team set should produce error")
	}
	fmt.Println(err)
}

func TestDuplicateWeeklyMatchup(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	w2 := copyWeeklyMatchup(w)
	_, err := common.CreateOne(db, w2)
	if err == nil {
		t.Fatal("Expected error on duplicate weekly matchup")
	}
	fmt.Println(err)
}
