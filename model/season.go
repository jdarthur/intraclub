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
	DraftId          DraftId            // Draft results for this Season
	ScheduleID       ScheduleId         // ID of the Schedule for this Season
	PlayoffStructure PlayoffStructureId // ID of the PlayoffStructure for the Season
	LateAdditions    []UserId           // IDs of any User s who were added to the Season after the season's Draft
}

func (s *Season) GetOwner() common.RecordId {
	return s.Commissioners[0].RecordId()
}

func (s *Season) UniquenessEquivalent(other *Season) error {
	if s.DraftId == other.DraftId {
		return fmt.Errorf("duplicate season for draft ID %s", s.DraftId)
	}
	return nil
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

func (s *Season) DynamicallyValid(db common.DatabaseProvider) error {
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
	for _, commissioner := range s.Commissioners {
		if commissioner == u {
			return true, nil
		}
	}

	for _, lateAdd := range s.LateAdditions {
		if lateAdd == u {
			return true, nil
		}
	}

	teams, err := s.GetTeams(db)
	if err != nil {
		return false, err
	}

	for _, team := range teams {
		if team.IsTeamMember(u) {
			return true, nil
		}
	}
	return false, nil
}

func (s *Season) IsUserIdACommissioner(u UserId) bool {
	for _, commissioner := range s.Commissioners {
		if commissioner == u {
			return true
		}
	}
	return false
}

func (s *Season) GetTeams(db common.DatabaseProvider) ([]*Team, error) {
	return common.GetAllWhere(db, &Team{}, func(c *Team) bool {
		for _, team := range s.Teams {
			if c.ID == team {
				return true
			}
		}
		return false
	})
}

func (s *Season) IsTeamAssignedToSeason(teamId TeamId) bool {
	for _, t := range s.Teams {
		if t == teamId {
			return true
		}
	}
	return false
}

func (s *Season) GetTeamCaptains(db common.DatabaseProvider) ([]UserId, error) {
	teams, err := s.GetTeams(db)
	if err != nil {
		return nil, fmt.Errorf("Failed to get teams for season: %s\n", err.Error())
	}
	output := make([]UserId, 0)
	for _, team := range teams {
		output = append(output, team.Captain)
	}
	return output, nil
}

// EditableBySeason returns a list of common.RecordId values who can edit a particular
// record based on those who can edit the record's associated Season. This function is
// used as a reusable way to compose the common.CrudRecord.EditableBy() list for record
// types which are downstream of a Season and editable by the commissioners, e.g. a Schedule
// or Week belonging to a Season
func EditableBySeason(db common.DatabaseProvider, seasonId SeasonId) []common.RecordId {
	season, err := common.GetExistingRecordById(db, &Season{}, seasonId.RecordId())
	if err != nil {
		fmt.Println(err) // shouldn't get here, but print an error if so for debugging
		return nil
	}
	return season.EditableBy(db)
}
