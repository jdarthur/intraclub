package model

import (
	"fmt"
	"intraclub/common"
	"testing"
)

func newStoredSchedule(t *testing.T, db common.DatabaseProvider, season *Season) *Schedule {
	schedule := NewSchedule()
	schedule.SeasonId = season.ID

	v, err := common.CreateOne(db, schedule)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestSeasonUpdatedOnScheduleCreate(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	season := newDefaultSeason(t, db)

	schedule := newStoredSchedule(t, db, season)

	season, err := common.GetExistingRecordById(db, &Season{}, schedule.SeasonId.RecordId())
	if err != nil {
		t.Fatal(err)
	}

	if season.ScheduleID != schedule.ID {
		t.Fatal("Expected season to have the new schedule ID saved")
	}
}

func TestOneSchedulePerSeason(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	season := newDefaultSeason(t, db)
	schedule := newStoredSchedule(t, db, season)

	schedule2 := NewSchedule()
	schedule2.SeasonId = season.ID
	schedule2.Matchups = schedule.Matchups
	_, err := common.CreateOne(db, schedule2)
	if err == nil {
		t.Fatal("Expected error creating a duplicate schedule")
	}
	fmt.Println(err)
}
