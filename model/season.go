package model

import (
	"fmt"
	"intraclub/common"
	"time"
)

type Season struct {
	ID        string
	StartTime time.Time
	Weeks     []*Week
}

var oneWeek = time.Hour * 24 * 7

var TimeFormat = "2006-01-02"

// RainDelayOn takes a Season and pushes all of the weeks back to the next
// week's date. This is usually 7 days later, but when there is a holiday in
// between weeks, this logic will correctly move the week around the holiday
// by just switching the week's Date to the next Week in the list (e.g. 14 days later)
func (s *Season) RainDelayOn(weekId string) error {

	startWeek := -1
	for i, week := range s.Weeks {
		if weekId == week.ID {
			startWeek = i
			break
		}
	}

	if startWeek == -1 {
		return fmt.Errorf("week with ID %s was not found in season %s", weekId, s.ID)
	}

	weeksAffected := s.Weeks[startWeek:]

	for i, week := range weeksAffected {

		// if this is the last week, we will push back a manual time period
		if i == len(weeksAffected)-1 {

			// check if the next week is a holiday maybe

			week.PushBack(1)
		} else {

			// if this isn't the last week, just use the date of the next entry in the season

			nextWeek := weeksAffected[i+1]
			week.Date = nextWeek.Date
		}
	}

	return nil
}

type Week struct {
	ID           string
	Date         time.Time // date when this week was actually played
	OriginalDate time.Time // date when this week was originally scheduled to play (e.g. before a rain day)
}

func (w *Week) RecordType() string {
	return "week"
}

func (w *Week) OneRecord() common.CrudRecord {
	return new(Week)
}

func (w *Week) ListOfRecords() interface{} {
	return make([]*Week, 0)
}

func (w *Week) SetId(id string) {
	w.ID = id
}

func (w *Week) GetId() string {
	return w.ID
}

func (w *Week) ValidateStatic() error {
	if w.Date.IsZero() {
		return fmt.Errorf("date field must not be empty")
	}

	if w.OriginalDate.IsZero() {
		return fmt.Errorf("original date field must not be empty")
	}

	return nil
}

func (w *Week) ValidateDynamic(db common.DbProvider) error {
	return nil
}

func (w *Week) PushBack(weeks int) {
	w.Date = w.Date.Add(time.Duration(weeks) * oneWeek)
}
