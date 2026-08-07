package model

import (
	"context"
	"fmt"
	"testing"

	"intraclub/database"
)

func newStoredSchedule(t *testing.T, db database.Provider, season *Season) *Schedule {
	schedule := NewSchedule()
	schedule.SeasonId = season.ID

	v, err := database.CreateOne(context.Background(), db, schedule)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestSeasonUpdatedOnScheduleCreate(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, _ := newDefaultSeason(t, db)

	schedule := newStoredSchedule(t, db, season)

	season, err := database.GetExistingRecordById(context.Background(), db, &Season{}, schedule.SeasonId.RecordId())
	if err != nil {
		t.Fatal(err)
	}

	if season.ScheduleID != schedule.ID {
		t.Fatal("Expected season to have the new schedule ID saved")
	}
}

func TestOneSchedulePerSeason(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, _ := newDefaultSeason(t, db)
	newStoredSchedule(t, db, season)

	schedule2 := NewSchedule()
	schedule2.SeasonId = season.ID
	_, err := database.CreateOne(context.Background(), db, schedule2)
	if err == nil {
		t.Fatal("Expected error creating a duplicate schedule")
	}
	fmt.Println(err)
}

// newStoredWeeklyMatchups creates count stored WeeklyMatchup records for the
// given season (one per week) and returns them in order.
func newStoredWeeklyMatchups(t *testing.T, db database.Provider, season *Season, count int) []*WeeklyMatchup {
	t.Helper()
	matchups := make([]*WeeklyMatchup, 0, count)
	for i := 0; i < count; i++ {
		week := newStoredWeek(t, db, season)
		w := NewWeeklyMatchup()
		w.WeekId = week.ID
		w.SeasonId = season.ID
		v, err := database.CreateOne(context.Background(), db, w)
		if err != nil {
			t.Fatal(err)
		}
		matchups = append(matchups, v)
	}
	return matchups
}

// newStoredScheduleWithMatchups creates a stored Schedule for the season and
// assigns the given weekly matchups to it via SetMatchups.
func newStoredScheduleWithMatchups(t *testing.T, db database.Provider, season *Season, matchups []*WeeklyMatchup) *Schedule {
	t.Helper()
	schedule := newStoredSchedule(t, db, season)
	ids := make([]WeeklyMatchupId, 0, len(matchups))
	for _, m := range matchups {
		ids = append(ids, m.ID)
	}
	err := schedule.SetMatchups(context.Background(), db, ids)
	if err != nil {
		t.Fatal(err)
	}
	return schedule
}

func TestScheduleMatchupRoundTrip(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, _ := newDefaultSeasonWithTeams(t, db, 4)
	weeklyMatchups := newStoredWeeklyMatchups(t, db, season, 3)

	schedule := newStoredScheduleWithMatchups(t, db, season, weeklyMatchups)

	got, err := schedule.GetMatchups(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 matchups, got %d", len(got))
	}
	for i, m := range weeklyMatchups {
		if got[i].ID != m.ID {
			t.Fatalf("expected matchup %d to be %s, got %s", i, m.ID, got[i].ID)
		}
	}
}

func TestScheduleMatchupPreservesOrderByPosition(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, _ := newDefaultSeasonWithTeams(t, db, 4)
	weeklyMatchups := newStoredWeeklyMatchups(t, db, season, 3)

	schedule := newStoredSchedule(t, db, season)

	// Assign matchups out of order to verify GetMatchups sorts by Position.
	ids := []WeeklyMatchupId{weeklyMatchups[2].ID, weeklyMatchups[0].ID, weeklyMatchups[1].ID}
	err := schedule.SetMatchups(context.Background(), db, ids)
	if err != nil {
		t.Fatal(err)
	}

	got, err := schedule.GetMatchups(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	// Positions were assigned in SetMatchups order (0, 1, 2) regardless of the
	// order the weekly matchups were created, so GetMatchups must return them
	// in that assignment order.
	expected := []WeeklyMatchupId{weeklyMatchups[2].ID, weeklyMatchups[0].ID, weeklyMatchups[1].ID}
	if len(got) != len(expected) {
		t.Fatalf("expected %d matchups, got %d", len(expected), len(got))
	}
	for i, id := range expected {
		if got[i].ID != id {
			t.Fatalf("expected matchup at position %d to be %s, got %s", i, id, got[i].ID)
		}
	}
}

func TestScheduleSetMatchupsReplacesExisting(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, _ := newDefaultSeasonWithTeams(t, db, 4)
	weeklyMatchups := newStoredWeeklyMatchups(t, db, season, 3)

	schedule := newStoredScheduleWithMatchups(t, db, season, weeklyMatchups[:2])

	// Replace with only the third matchup.
	err := schedule.SetMatchups(context.Background(), db, []WeeklyMatchupId{weeklyMatchups[2].ID})
	if err != nil {
		t.Fatal(err)
	}

	got, err := schedule.GetMatchups(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 matchup after replacement, got %d", len(got))
	}
	if got[0].ID != weeklyMatchups[2].ID {
		t.Fatalf("expected matchup to be %s, got %s", weeklyMatchups[2].ID, got[0].ID)
	}
}

func TestScheduleMatchupUniqueness(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, _ := newDefaultSeasonWithTeams(t, db, 4)
	weeklyMatchups := newStoredWeeklyMatchups(t, db, season, 1)
	schedule := newStoredSchedule(t, db, season)

	sm := NewScheduleMatchup(schedule.ID, weeklyMatchups[0].ID, 0)
	_, err := database.CreateOne(context.Background(), db, sm)
	if err != nil {
		t.Fatal(err)
	}

	// Assigning the same weekly matchup to the same schedule again must fail.
	sm2 := NewScheduleMatchup(schedule.ID, weeklyMatchups[0].ID, 1)
	_, err = database.CreateOne(context.Background(), db, sm2)
	if err == nil {
		t.Fatal("expected error assigning duplicate weekly matchup to schedule")
	}
	fmt.Println(err)
}

func TestScheduleMatchupDynamicallyValid(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, _ := newDefaultSeasonWithTeams(t, db, 4)
	weeklyMatchups := newStoredWeeklyMatchups(t, db, season, 1)

	// Invalid schedule ID.
	sm := NewScheduleMatchup(ScheduleId(database.InvalidRecordId), weeklyMatchups[0].ID, 0)
	err := sm.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("expected error for invalid schedule ID")
	}

	// Invalid weekly matchup ID.
	schedule := newStoredSchedule(t, db, season)
	sm2 := NewScheduleMatchup(schedule.ID, WeeklyMatchupId(database.InvalidRecordId), 0)
	err = sm2.DynamicallyValid(context.Background(), db)
	if err == nil {
		t.Fatal("expected error for invalid weekly matchup ID")
	}
}

func TestScheduleIsComplete(t *testing.T) {
	db := database.NewUnitTestDBProvider()

	// Incomplete schedule: season has two weeks but only one weekly matchup assigned.
	season1, _ := newDefaultSeasonWithTeams(t, db, 4)
	weeklyMatchups1 := newStoredWeeklyMatchups(t, db, season1, 2)
	schedule1 := newStoredScheduleWithMatchups(t, db, season1, weeklyMatchups1[:1])
	complete, err := schedule1.IsScheduleComplete(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if complete {
		t.Fatal("expected schedule to be incomplete")
	}

	// Complete schedule: season has two weeks and two weekly matchups assigned.
	season2, _ := newDefaultSeasonWithTeams(t, db, 4)
	weeklyMatchups2 := newStoredWeeklyMatchups(t, db, season2, 2)
	schedule2 := newStoredScheduleWithMatchups(t, db, season2, weeklyMatchups2)
	complete, err = schedule2.IsScheduleComplete(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if !complete {
		t.Fatal("expected schedule to be complete")
	}
}

func TestSchedulePreDeleteCascadesMatchups(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	season, _ := newDefaultSeasonWithTeams(t, db, 4)
	weeklyMatchups := newStoredWeeklyMatchups(t, db, season, 2)
	schedule := newStoredScheduleWithMatchups(t, db, season, weeklyMatchups)

	rows, err := database.GetAllWhere[*ScheduleMatchup](context.Background(), db, func(_ context.Context, sm *ScheduleMatchup) bool {
		return sm.ScheduleId == schedule.ID
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 ScheduleMatchup rows before delete, got %d", len(rows))
	}

	_, _, err = database.DeleteOneById(context.Background(), db, &Schedule{}, schedule.ID.RecordId())
	if err != nil {
		t.Fatal(err)
	}

	rows, err = database.GetAllWhere[*ScheduleMatchup](context.Background(), db, func(_ context.Context, sm *ScheduleMatchup) bool {
		return sm.ScheduleId == schedule.ID
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 ScheduleMatchup rows after delete, got %d", len(rows))
	}
}
