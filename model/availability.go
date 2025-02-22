package model

import (
	"fmt"
	"intraclub/common"
)

type AvailabilityOption int

const (
	AvailabilityUnset AvailabilityOption = iota
	AvailabilityAvailable
	AvailabilityMaybe
	AvailabilityNotAvailable
	AvailabilityInvalid
)

func (opt AvailabilityOption) String() string {
	switch opt {
	case AvailabilityUnset:
		return "unset"
	case AvailabilityAvailable:
		return "available"
	case AvailabilityMaybe:
		return "maybe"
	case AvailabilityNotAvailable:
		return "not-available"
	default:
		return "invalid"
	}
}

func (opt AvailabilityOption) Valid() bool {
	return opt < AvailabilityInvalid
}

type Availability struct {
	ID        common.RecordId
	UserId    UserId
	WeekId    WeekId
	Available AvailabilityOption
}

func (a *Availability) SetOwner(recordId common.RecordId) {
	a.UserId = UserId(recordId)
}

func NewAvailability() *Availability {
	return &Availability{}
}

func (a *Availability) Type() string {
	return "availability"
}

func (a *Availability) GetId() common.RecordId {
	return a.ID
}

func (a *Availability) SetId(id common.RecordId) {
	a.ID = id
}

func (a *Availability) StaticallyValid() error {
	if !a.Available.Valid() {
		return fmt.Errorf("availability option %d is not valid", a.Available)
	}
	return nil
}

func (a *Availability) DynamicallyValid(db common.DatabaseProvider, existing common.DatabaseValidatable) error {
	err := common.ExistsById(db, &User{}, a.UserId.RecordId())
	if err != nil {
		return err
	}
	return common.ExistsById(db, &Week{}, a.WeekId.RecordId())
}

func (a *Availability) EditableBy(db common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{a.UserId.RecordId()}
}

func (a *Availability) getTeam(db common.DatabaseProvider) (*Team, error) {
	week, exists, err := common.GetOneById(db, &Week{}, a.WeekId.RecordId())
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("week with ID %s does not exist", a.WeekId.RecordId())
	}

	season, err := GetSeason(db, week.SeasonId)
	if err != nil {
		return nil, err
	}

	seasonMember, err := season.IsUserIdASeasonParticipant(db, a.UserId)
	if err != nil {
		return nil, err
	}
	if !seasonMember {
		return nil, fmt.Errorf("user %s is not a participant in season %s", a.UserId.String(), season.ID)
	}

	for _, teamId := range season.Teams {
		team, exists, err := common.GetOneById(db, &Team{}, teamId.RecordId())
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("team with ID %s does not exist", teamId.RecordId())
		}
		if team.IsTeamMember(a.UserId) {
			return team, nil
		}
	}

	return nil, fmt.Errorf("user %s was not found on any teams in season %s", a.UserId, season.ID)
}

func (a *Availability) AccessibleTo(db common.DatabaseProvider) []common.RecordId {
	team, err := a.getTeam(db)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return UserIdListToRecordIdList(team.Members)
}
