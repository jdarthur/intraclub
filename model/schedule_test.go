package model

import (
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

	season, err := GetSeason(db, schedule.SeasonId)
	if err != nil {
		t.Fatal(err)
	}

	if season.ScheduleID != schedule.ID {
		t.Fatal("Expected season to have the new schedule ID saved")
	}

}
