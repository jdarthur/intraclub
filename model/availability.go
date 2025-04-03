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

func (a *Availability) UniquenessEquivalent(other *Availability) error {
	if a.UserId == other.UserId && a.WeekId == other.WeekId {
		return fmt.Errorf("duplicate record for user ID & week ID")
	}
	return nil
}

func NewAvailability() *Availability {
	return &Availability{}
}

func (a *Availability) SetOwner(recordId common.RecordId) {
	a.UserId = UserId(recordId)
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

func (a *Availability) DynamicallyValid(db common.DatabaseProvider) error {
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

	draft, err := common.GetExistingRecordById(db, &Draft{}, week.DraftId.RecordId())
	if err != nil {
		return nil, err
	}

	season, err := draft.GetSeason(db)
	if err != nil {
		return nil, err
	}

	if season == nil {
		return nil, fmt.Errorf("draft %s (from week %s) does not have an assigned season", draft.ID, a.WeekId)
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

func GetAvailabilityForUser(db common.DatabaseProvider, userId UserId, draftId DraftId) ([]*Availability, error) {
	weeks, err := common.GetAllWhere(db, &Week{}, func(c *Week) bool {
		return c.DraftId == draftId
	})

	if err != nil {
		return nil, err
	}

	return common.GetAllWhere(db, &Availability{}, func(c *Availability) bool {
		// only return Availability associated with this User
		if c.UserId != userId {
			return false
		}

		// only return Availability associated with the Weeks of this Season
		for _, week := range weeks {
			if c.WeekId == week.ID {
				return true
			}
		}
		return false
	})
}

func GetAvailabilityForTeam(db common.DatabaseProvider, teamId TeamId, draftId DraftId) (map[UserId][]*Availability, error) {
	output := make(map[UserId][]*Availability)
	team, err := common.GetExistingRecordById(db, &Team{}, teamId.RecordId())
	if err != nil {
		return nil, err
	}
	for _, member := range team.Members {
		availability, err := GetAvailabilityForUser(db, member, draftId)
		if err != nil {
			return nil, err
		}
		output[member] = availability
	}
	return output, nil
}
