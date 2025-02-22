package model

import (
	"errors"
	"fmt"
	"intraclub/common"
	"time"
)

type SeasonId common.RecordId

func (id SeasonId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id SeasonId) String() string {
	return id.RecordId().String()
}

type weeklyMatchup struct {
	WeekId   WeekId
	HomeTeam TeamId
	AwayTeam TeamId
}

type StartTime time.Time

func NewStartTime(hour int, minute int) StartTime {
	t := time.Date(1, 1, 1, hour, minute, 0, 0, time.UTC)
	return StartTime(t)
}

func (s StartTime) StaticallyValid() error {
	if time.Time(s).Year() > 1 {
		return fmt.Errorf("start time year should be zero (got %d)", time.Time(s).Year())
	} else if time.Time(s).Month() > 1 {
		return fmt.Errorf("start time month should be zero (got %d)", time.Time(s).Month())
	} else if time.Time(s).Day() > 1 {
		return fmt.Errorf("start time day should be zero (got %d)", time.Time(s).Day())
	} else if time.Time(s).Hour() == 0 {
		return fmt.Errorf("start time hour should not be zero (got %d)", time.Time(s).Hour())
	}
	return nil
}

func (s StartTime) String() string {
	return time.Time(s).Format("15:04 PM")
}

type Season struct {
	ID               SeasonId           // unique identifier for this Season
	Name             string             // descriptive name for this season, e.g. _Men's Intraclub 2025_
	Facility         FacilityId         // ID of the Facility at which this Season is played
	StartTime        StartTime          // time of day when the first matches kick off (e.g. _8:30 AM_)
	Commissioners    []UserId           // list of User IDs who act as commissioners in this Season
	Teams            []TeamId           // list of Team IDs participating in this Season
	Weeks            []WeekId           // list of Week IDs for this Season
	DraftId          DraftId            // Draft results for this Season
	ScheduleID       ScheduleId         // ID of the Schedule for this Season
	PlayoffStructure PlayoffStructureId // ID of the PlayoffStructure for the Season
	LateAdditions    []UserId           // IDs of any User s who were added to the Season after the season's Draft
}

func (s *Season) SetOwner(recordId common.RecordId) {
	s.Commissioners = []UserId{UserId(recordId)}
}

func NewSeason() *Season {
	return &Season{}
}

func (s *Season) EditableBy(common.DatabaseProvider) []common.RecordId {
	return UserIdListToRecordIdList(s.Commissioners)
}

func (s *Season) AccessibleTo(common.DatabaseProvider) []common.RecordId {
	return common.AccessibleToEveryone
}

func (s *Season) Type() string {
	return "season"
}

func (s *Season) GetId() common.RecordId {
	return s.ID.RecordId()
}

func (s *Season) SetId(id common.RecordId) {
	s.ID = SeasonId(id)
}

func (s *Season) StaticallyValid() error {
	if s.Name == "" {
		return errors.New("season name is empty")
	}
	if len(s.Commissioners) == 0 {
		return errors.New("season commissioners is empty")
	}
	if len(s.Teams) == 0 {
		return errors.New("season teams is empty")
	}

	return s.StartTime.StaticallyValid()
}

func (s *Season) DynamicallyValid(db common.DatabaseProvider, existing common.DatabaseValidatable) error {
	for _, commissioner := range s.Commissioners {
		if err := common.ExistsById(db, &User{}, commissioner.RecordId()); err != nil {
			return err
		}
	}

	for _, team := range s.Teams {
		if err := common.ExistsById(db, &Team{}, team.RecordId()); err != nil {
			return err
		}
	}

	for _, week := range s.Weeks {
		if err := common.ExistsById(db, &Week{}, week.RecordId()); err != nil {
			return err
		}
	}

	err := common.ExistsById(db, &Draft{}, s.DraftId.RecordId())
	if err != nil {
		return err
	}

	err = common.ExistsById(db, &Facility{}, s.Facility.RecordId())
	if err != nil {
		return err
	}

	if s.ScheduleID.RecordId() != common.InvalidRecordId {
		err = common.ExistsById(db, &Schedule{}, s.ScheduleID.RecordId())
		if err != nil {
			return err
		}
	}

	if s.PlayoffStructure.RecordId() != common.InvalidRecordId {
		err = common.ExistsById(db, &PlayoffStructure{}, s.PlayoffStructure.RecordId())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Season) GetDraft(db common.DatabaseProvider) (*Draft, error) {
	draft, exists, err := common.GetOneById(db, &Draft{}, s.DraftId.RecordId())
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("draft with ID %d does not exist", s.DraftId)
	}
	return draft, nil
}

func (s *Season) IsUserIdASeasonParticipant(db common.DatabaseProvider, u UserId) (bool, error) {

	for _, lateAdd := range s.LateAdditions {
		if lateAdd == u {
			return true, nil
		}
	}

	draft, err := s.GetDraft(db)
	if err != nil {
		return false, err
	}

	return draft.IsAvailableToSelect(u) || draft.IsSelected(u), nil
}

func (s *Season) IsUserIdACommissioner(u UserId) bool {
	for _, commissioner := range s.Commissioners {
		if commissioner == u {
			return true
		}
	}
	return false
}

func GetSeason(db common.DatabaseProvider, id SeasonId) (*Season, error) {
	season, exists, err := common.GetOneById(db, &Season{}, id.RecordId())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("season with ID %s does not exist\n", id)
	}
	return season, nil
}
