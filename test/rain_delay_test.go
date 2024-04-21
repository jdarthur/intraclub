package test

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
	"intraclub/model"
	"testing"
	"time"
)

var OriginalWeek1 = time.Date(2024, time.March, 24, 0, 0, 0, 0, time.UTC)
var OriginalWeek2 = time.Date(2024, time.March, 31, 0, 0, 0, 0, time.UTC)
var OriginalWeek3 = time.Date(2024, time.April, 6, 0, 0, 0, 0, time.UTC)
var OriginalWeek4 = time.Date(2024, time.April, 13, 0, 0, 0, 0, time.UTC)

func TestRainDelay(t *testing.T) {

	weeks := []*model.Week{
		WeekAt(OriginalWeek1),
		WeekAt(OriginalWeek2),
		WeekAt(OriginalWeek3),
		WeekAt(OriginalWeek4),
	}

	weekIds := make([]string, 0)

	for _, week := range weeks {
		w, err := common.Create(common.GlobalDbProvider, week)
		if err != nil {
			panic(err)
		}

		weekIds = append(weekIds, w.GetId().Hex())
	}

	season := &model.Season{
		ID:        primitive.NewObjectID(),
		StartTime: time.Date(0, 0, 0, 80, 30, 0, 0, time.UTC),
		Weeks:     weekIds,
	}

	err := season.RainDelayOn(common.GlobalDbProvider, weekIds[1])
	if err != nil {
		t.Fatal(err)
	}

	if len(season.Weeks) != 4 {
		t.Fatalf("Expected 4 week season, got %d after RainDelayOn", len(season.Weeks))
	}

	weeks, err = season.GetWeeks(common.GlobalDbProvider)
	if err != nil {
		t.Fatal(err)
	}

	week1 := weeks[0]
	expectWeek(week1, OriginalWeek1, OriginalWeek1, t)

	week2 := weeks[1]
	expectWeek(week2, OriginalWeek2, OriginalWeek3, t)

	week3 := weeks[2]
	expectWeek(week3, OriginalWeek3, OriginalWeek4, t)

	week4 := weeks[3]
	expectWeek(week4, OriginalWeek4, time.Date(2024, time.April, 20, 0, 0, 0, 0, time.UTC), t)

}

func expectWeek(week *model.Week, origDate, newDate time.Time, t *testing.T) {
	if week.OriginalDate != origDate {
		t.Errorf("Expected week %s original date to be %s (got %s)", week.ID, origDate.Format(model.TimeFormat), week.OriginalDate.Format(model.TimeFormat))
	}

	if week.Date != newDate {
		t.Errorf("Expected week %s date to be %s (got %s)", week.ID, newDate.Format(model.TimeFormat), week.Date.Format(model.TimeFormat))
	}
}

func WeekAt(t time.Time) *model.Week {
	return &model.Week{
		Date:         t,
		OriginalDate: t,
	}
}
