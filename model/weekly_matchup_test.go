package model

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"intraclub/database"
)

func newStoredWeeklyMatchup(t *testing.T, db database.Provider) (*Season, *WeeklyMatchup) {
	t.Helper()
	season, _ := newDefaultSeasonWithTeams(t, db, 4)
	week := newStoredWeek(t, db, season)

	teams, err := season.GetTeams(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	w := NewWeeklyMatchup()
	w.WeekId = week.ID
	w.SeasonId = season.ID

	v, err := database.CreateOne(context.Background(), db, w)
	if err != nil {
		t.Fatal(err)
	}

	// Create matchup records via SetMatchups
	matchup := TeamMatchup{
		HomeTeam: teams[0].ID,
		AwayTeam: teams[1].ID,
	}
	matchup2 := TeamMatchup{
		HomeTeam: teams[2].ID,
		AwayTeam: teams[3].ID,
	}

	err = v.SetMatchups(context.Background(), db, []*TeamMatchup{&matchup, &matchup2})
	if err != nil {
		t.Fatal(err)
	}
	return season, v
}


func TestWeeklyMatchupInvalidHomeTeamId(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	// Update the matchup to have an invalid home team (bypass validation to set bad data)
	matchups, err := w.GetMatchups(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	matchups[0].HomeTeam = TeamId(database.InvalidRecordId)
	err = w.setMatchupsRaw(context.Background(), db, matchups)
	if err != nil {
		t.Fatal(err)
	}

	err = w.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("Invalid home team ID should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupInvalidAwayTeamId(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	matchups, err := w.GetMatchups(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	matchups[0].AwayTeam = TeamId(database.InvalidRecordId)
	err = w.setMatchupsRaw(context.Background(), db, matchups)
	if err != nil {
		t.Fatal(err)
	}

	err = w.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("Invalid away team ID should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupInvalidSeasonId(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)
	w.SeasonId = SeasonId(database.InvalidRecordId)
	err := w.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("Invalid season ID should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupInvalidWeekId(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)
	w.WeekId = WeekId(database.InvalidRecordId)
	err := w.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("Invalid week ID should produce error")
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
		t.Fatal("Week from another season should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupHomeTeamDoesNotBelongToSeason(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	otherTeam := newStoredTeam(t, db, newStoredUser(t, db).ID)
	matchups, err := w.GetMatchups(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	matchups[0].HomeTeam = otherTeam.ID
	err = w.SetMatchups(context.Background(), db, matchups)
	if err != nil {
		t.Fatal(err)
	}

	err = w.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("Team from another season should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupAwayTeamDoesNotBelongToSeason(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	otherTeam := newStoredTeam(t, db, newStoredUser(t, db).ID)
	matchups, err := w.GetMatchups(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	matchups[0].AwayTeam = otherTeam.ID
	err = w.SetMatchups(context.Background(), db, matchups)
	if err != nil {
		t.Fatal(err)
	}

	err = w.DynamicallyValid(context.Background(), db)
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

	matchups, err := w.GetMatchups(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	anotherMatchup := TeamMatchup{
		HomeTeam: teams[0].ID,
		AwayTeam: teams[2].ID,
	}
	matchups = append(matchups, &anotherMatchup)
	// Bypass validation so the double-booked team is persisted and caught by
	// the double-booking check in DynamicallyValid rather than the uniqueness constraint.
	err = w.setMatchupsRaw(context.Background(), db, matchups)
	if err != nil {
		t.Fatal(err)
	}

	err = w.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("Double matchup for team 1 should produce error")
	}
	fmt.Println(err)
}

func TestWeeklyMatchupTeamDoesNotHaveAnyMatchups(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	matchups, err := w.GetMatchups(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	matchups[1].Bye = true
	matchups[1].AwayTeam = TeamId(database.InvalidRecordId)
	err = w.SetMatchups(context.Background(), db, matchups)
	if err != nil {
		t.Fatal(err)
	}

	err = database.Validate(context.Background(), db, w)
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

	v, err := database.CreateOne(context.Background(), db, w)
	if err != nil {
		t.Fatal(err)
	}

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

	err = v.SetMatchups(context.Background(), db, []*TeamMatchup{&matchup, &bye1, &bye2})
	if err != nil {
		t.Fatal(err)
	}

	err = database.Validate(context.Background(), db, v)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWeeklyMatchupHomeTeamByeButAwayTeamIsSet(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	matchups, err := w.GetMatchups(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	matchups[0].Bye = true
	err = w.SetMatchups(context.Background(), db, matchups)
	if err != nil {
		t.Fatal(err)
	}

	err = database.Validate(context.Background(), db, w)
	if err == nil {
		t.Fatal("Bye with away team set should produce error")
	}
	fmt.Println(err)
}

func TestDuplicateWeeklyMatchup(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	w2 := NewWeeklyMatchup()
	w2.WeekId = w.WeekId
	w2.SeasonId = w.SeasonId
	_, err := database.CreateOne(context.Background(), db, w2)
	if err == nil {
		t.Fatal("Expected error on duplicate weekly matchup")
	}
	fmt.Println(err)
}

// TestWeeklyMatchupRoundTrip verifies that matchups can be set and retrieved correctly.
func TestWeeklyMatchupRoundTrip(t *testing.T) {
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

	v, err := database.CreateOne(context.Background(), db, w)
	if err != nil {
		t.Fatal(err)
	}

	matchup := TeamMatchup{
		HomeTeam: teams[0].ID,
		AwayTeam: teams[1].ID,
	}
	matchup2 := TeamMatchup{
		HomeTeam: teams[2].ID,
		AwayTeam: teams[3].ID,
	}

	err = v.SetMatchups(context.Background(), db, []*TeamMatchup{&matchup, &matchup2})
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve and verify
	retrieved, err := v.GetMatchups(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(retrieved) != 2 {
		t.Fatalf("Expected 2 matchups, got %d", len(retrieved))
	}
	if retrieved[0].HomeTeam != teams[0].ID {
		t.Fatalf("Expected home team %s, got %s", teams[0].ID, retrieved[0].HomeTeam)
	}
	if retrieved[0].AwayTeam != teams[1].ID {
		t.Fatalf("Expected away team %s, got %s", teams[1].ID, retrieved[0].AwayTeam)
	}
	if retrieved[1].HomeTeam != teams[2].ID {
		t.Fatalf("Expected home team %s, got %s", teams[2].ID, retrieved[1].HomeTeam)
	}
	if retrieved[1].AwayTeam != teams[3].ID {
		t.Fatalf("Expected away team %s, got %s", teams[3].ID, retrieved[1].AwayTeam)
	}
}

// TestWeeklyMatchupCascadeDelete verifies that deleting a WeeklyMatchup
// also deletes all associated WeeklyMatchupTeamMatchup records.
func TestWeeklyMatchupCascadeDelete(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	// Verify matchups exist
	matchups, err := w.GetMatchups(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(matchups) != 2 {
		t.Fatalf("Expected 2 matchups, got %d", len(matchups))
	}

	// Delete the weekly matchup
	_, _, err = database.DeleteOneById(context.Background(), db, &WeeklyMatchup{}, w.ID.RecordId())
	if err != nil {
		t.Fatal(err)
	}

	// Verify matchup records are gone
	allRecords, err := database.GetAllWhere[*WeeklyMatchupTeamMatchup](context.Background(), db, func(_ context.Context, wmtm *WeeklyMatchupTeamMatchup) bool {
		return wmtm.WeeklyMatchupId == w.ID
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(allRecords) != 0 {
		t.Fatalf("Expected 0 matchup records after cascade delete, got %d", len(allRecords))
	}
}

// TestWeeklyMatchupTeamMatchupRecord validates the normalized record type.
func TestWeeklyMatchupTeamMatchupRecord(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	_, w := newStoredWeeklyMatchup(t, db)

	// Query the underlying records directly
	records, err := database.GetAllWhere[*WeeklyMatchupTeamMatchup](context.Background(), db, func(_ context.Context, wmtm *WeeklyMatchupTeamMatchup) bool {
		return wmtm.WeeklyMatchupId == w.ID
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("Expected 2 WeeklyMatchupTeamMatchup records, got %d", len(records))
	}

	// Verify Position field preserves ordering
	slices.SortFunc(records, func(a, b *WeeklyMatchupTeamMatchup) int {
		return a.Position - b.Position
	})
	if records[0].Position != 0 {
		t.Fatalf("Expected position 0, got %d", records[0].Position)
	}
	if records[1].Position != 1 {
		t.Fatalf("Expected position 1, got %d", records[1].Position)
	}

	// Verify FK
	if records[0].WeeklyMatchupId != w.ID {
		t.Fatalf("Expected weekly_matchup_id %s, got %s", w.ID, records[0].WeeklyMatchupId)
	}
}
