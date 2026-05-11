package model

import (
	"context"
	"fmt"
	"testing"

	"intraclub/database"
)

func newStoredSchedule(t *testing.T, db database.DatabaseProvider, season *Season) *Schedule {
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
	schedule := newStoredSchedule(t, db, season)

	schedule2 := NewSchedule()
	schedule2.SeasonId = season.ID
	schedule2.Matchups = schedule.Matchups
	_, err := database.CreateOne(context.Background(), db, schedule2)
	if err == nil {
		t.Fatal("Expected error creating a duplicate schedule")
	}
	fmt.Println(err)
}
