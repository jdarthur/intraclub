package test

import (
	"intraclub/model"
	"testing"
	"time"
)

var OriginalWeek1 = time.Date(2024, time.March, 24, 0, 0, 0, 0, time.UTC)
var OriginalWeek2 = time.Date(2024, time.March, 31, 0, 0, 0, 0, time.UTC)
var OriginalWeek3 = time.Date(2024, time.April, 6, 0, 0, 0, 0, time.UTC)
var OriginalWeek4 = time.Date(2024, time.April, 13, 0, 0, 0, 0, time.UTC)

func TestRainDelay(t *testing.T) {

	season := &model.Season{
		ID:        "",
		StartTime: time.Date(0, 0, 0, 80, 30, 0, 0, time.UTC),
		Weeks: []*model.Week{
			WeekAt(OriginalWeek1, "1"),
			WeekAt(OriginalWeek2, "2"),
			WeekAt(OriginalWeek3, "3"),
			WeekAt(OriginalWeek4, "4"),
		},
	}

	err := season.RainDelayOn("2")
	if err != nil {
		t.Fatal(err)
	}

	if len(season.Weeks) != 4 {
		t.Fatalf("Expected 4 week season, got %d after RainDelayOn", len(season.Weeks))
	}

	week1 := season.Weeks[0]
	expectWeek(week1, OriginalWeek1, OriginalWeek1, t)

	week2 := season.Weeks[1]
	expectWeek(week2, OriginalWeek2, OriginalWeek3, t)

	week3 := season.Weeks[2]
	expectWeek(week3, OriginalWeek3, OriginalWeek4, t)

	week4 := season.Weeks[3]
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

func WeekAt(t time.Time, id string) *model.Week {
	return &model.Week{
		ID:           id,
		Date:         t,
		OriginalDate: t,
	}

}
