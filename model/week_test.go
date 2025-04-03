package model

import (
	"fmt"
	"intraclub/common"
	"testing"
	"time"
)

func newStoredWeek(t *testing.T, db common.DatabaseProvider, season *Season) *Week {
	week := NewWeek()
	week.DraftId = season.DraftId
	week.Date = time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC)
	v, err := common.CreateOne(db, week)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func newStoredWeekAt(t *testing.T, db common.DatabaseProvider, season *Season, date time.Time) *Week {
	week := NewWeek()
	week.DraftId = season.DraftId
	week.Date = date
	v, err := common.CreateOne(db, week)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func newDefaultSeasonWithWeeks(t *testing.T, db common.DatabaseProvider, weekCount int) (*Season, []*Week) {
	season := newDefaultSeasonWithTeams(t, db, 4)

	weeks := make([]*Week, 0)
	for i := 0; i < weekCount; i++ {
		weeks = append(weeks, newStoredWeek(t, db, season))
	}
	return season, weeks
}

func newDefaultSeasonWithWeeksAndTeams(t *testing.T, db common.DatabaseProvider, teams []*Team, weekCount int) (*Season, []*Week) {
	commissioner := teams[0].Captain
	season := newStoredSeason(t, db, commissioner, teams)

	weeks := make([]*Week, 0)
	for i := 0; i < weekCount; i++ {
		weeks = append(weeks, newStoredWeek(t, db, season))
	}
	return season, weeks
}

func TestInvalidDraftId(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	week := NewWeek()
	week.Date = time.Now()
	_, err := common.CreateOne(db, week)
	if err == nil {
		t.Fatal("expected error on invalid draft id")
	}
	fmt.Println(err)
}

func TestWeeksSorted(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	draft, season := newCompletedDraft(t, db)
	week1 := newStoredWeekAt(t, db, season, time.Now().AddDate(0, 0, 1))
	week2 := newStoredWeekAt(t, db, season, time.Now().AddDate(0, 0, 5))
	week3 := newStoredWeekAt(t, db, season, time.Now().AddDate(0, 0, 3))

	weeks, err := GetWeeksForDraft(db, draft.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(weeks) != 3 {
		t.Fatalf("expected 3 weeks, got %d", len(weeks))
	}
	if weeks[0].ID != week1.ID {
		t.Fatal("expected week1 at index 0")
	}
	if weeks[1].ID != week3.ID {
		t.Fatal("expected week3 at index 1")
	}
	if weeks[2].ID != week2.ID {
		t.Fatal("expected week2 at index 2")
	}
}
