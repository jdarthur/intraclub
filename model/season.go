package model

import (
	"errors"
	"fmt"
	"intraclub/common"
	"time"
)

type Season struct {
	ID        string    `json:"season_id" bson:"season_id"`
	LeagueId  string    `json:"league_id" bson:"league_id"`
	StartTime time.Time `json:"start_time" bson:"start_time"`
	Weeks     []string  `json:"weeks" bson:"weeks"`
}

func (s *Season) RecordType() string {
	return "season"
}

func (s *Season) OneRecord() common.CrudRecord {
	return new(Season)
}

type listOfSeasons []*Season

func (l listOfSeasons) Length() int {
	return len(l)
}

func (s *Season) ListOfRecords() common.ListOfCrudRecords {
	return make(listOfSeasons, 0)
}

func (s *Season) SetId(id string) {
	s.ID = id
}

func (s *Season) GetId() string {
	return s.ID
}

func (s *Season) ValidateStatic() error {
	year, month, day := s.StartTime.Date()

	if year != 0 {
		return errors.New("year must not be set in start time")
	}

	if month != 0 {
		return errors.New("month must not be set in start time")
	}

	if day != 0 {
		return errors.New("day must not be set in start time")
	}

	return nil
}

func (s *Season) ValidateDynamic(db common.DbProvider) error {

	err := common.CheckExistenceOrError(db, &League{ID: s.LeagueId})
	if err != nil {
		return err
	}

	for _, w := range s.Weeks {
		err := common.CheckExistenceOrError(db, &Week{ID: w})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Season) GetWeeks(provider common.DbProvider) ([]*Week, error) {

	weeks := make([]*Week, 0)
	for _, w := range s.Weeks {
		search := &Week{ID: w}
		week, exists, err := provider.GetOne(search)
		if !exists {
			return nil, common.RecordDoesNotExist(&Week{})
		}
		if err != nil {
			return nil, err
		}
		weeks = append(weeks, week.(*Week))
	}

	return weeks, nil
}

var oneWeek = time.Hour * 24 * 7

var TimeFormat = "2006-01-02"

// RainDelayOn takes a Season and pushes all of the weeks back to the next
// week's date. This is usually 7 days later, but when there is a holiday in
// between weeks, this logic will correctly move the week around the holiday
// by just switching the week's Date to the next Week in the list (e.g. 14 days later)
func (s *Season) RainDelayOn(provider common.DbProvider, weekId string) error {

	startWeek := -1
	weeks, err := s.GetWeeks(provider)
	if err != nil {
		return err
	}

	for i, week := range weeks {
		if weekId == week.ID {
			startWeek = i
			break
		}
	}

	if startWeek == -1 {
		return fmt.Errorf("week with ID %s was not found in season %s", weekId, s.ID)
	}

	weeksAffected := weeks[startWeek:]

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

		err = provider.Update(week)
		if err != nil {
			return err
		}
	}

	return nil
}
