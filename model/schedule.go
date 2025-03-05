package model

import (
	"fmt"
	"intraclub/common"
)

type TeamMatchup struct {
	HomeTeam common.RecordId
	AwayTeam common.RecordId
}

type WeeklyMatchup struct {
	WeekId   common.RecordId
	Matchups []TeamMatchup
}

func (w WeeklyMatchup) StaticallyValid() error {
	return nil
}

func (w WeeklyMatchup) DynamicallyValid(db common.DatabaseProvider) error {
	err := common.ExistsById(db, &Week{}, w.WeekId)
	if err != nil {
		return err
	}

	for _, matchup := range w.Matchups {
		err = common.ExistsById(db, &Team{}, matchup.HomeTeam)
		if err != nil {
			return err
		}
		err = common.ExistsById(db, &Team{}, matchup.AwayTeam)
		if err != nil {
			return err
		}
	}

	return nil
}

type ScheduleId common.RecordId

func (id ScheduleId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id ScheduleId) String() string {
	return id.RecordId().String()
}

type Schedule struct {
	ID       ScheduleId
	SeasonId SeasonId
	Matchups []WeeklyMatchup
}

func (s *Schedule) SetOwner(recordId common.RecordId) {
	// don't need to do anything here as the ownership of the
	// Schedule record type is automatically inferred &
	// enforced by the associated Season assigned to it
}

func NewSchedule() *Schedule {
	return &Schedule{}
}

func (s *Schedule) EditableBy(db common.DatabaseProvider) []common.RecordId {
	season, err := GetSeason(db, s.SeasonId)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return season.EditableBy(db)
}

func (s *Schedule) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func (s *Schedule) Type() string {
	return "schedule"
}

func (s *Schedule) GetId() common.RecordId {
	return s.ID.RecordId()
}

func (s *Schedule) SetId(id common.RecordId) {
	s.ID = ScheduleId(id)
}

func (s *Schedule) StaticallyValid() error {
	return nil
}

func (s *Schedule) DynamicallyValid(db common.DatabaseProvider) error {
	err := common.ExistsById(db, &Season{}, s.SeasonId.RecordId())
	if err != nil {
		return err
	}

	for _, m := range s.Matchups {
		err = m.DynamicallyValid(db)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Schedule) PostCreate(db common.DatabaseProvider) error {
	season, err := GetSeason(db, s.SeasonId)
	if err != nil {
		return err
	}
	season.ScheduleID = s.ID
	return common.UpdateOne(db, season)
}
