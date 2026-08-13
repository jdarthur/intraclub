package model

import (
	"context"
	"fmt"

	"intraclub/database"
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
	ID        database.RecordId  `json:"id"`
	UserId    database.UserId    `json:"user_id"`
	WeekId    WeekId             `json:"week_id"`
	Available AvailabilityOption `json:"available"`
}

func (a *Availability) GetOwner() database.UserId {
	return a.UserId
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

func (a *Availability) SetOwner(userId database.UserId) {
	a.UserId = userId
}

func (a *Availability) Type() string {
	return "availability"
}

func (a *Availability) GetId() database.RecordId {
	return a.ID
}

func (a *Availability) SetId(id database.RecordId) {
	a.ID = id
}

func (a *Availability) StaticallyValid() error {
	if !a.Available.Valid() {
		return fmt.Errorf("availability option %d is not valid", a.Available)
	}
	return nil
}

func (a *Availability) DynamicallyValid(ctx context.Context, db database.Provider) error {
	err := database.ExistsById(ctx, db, &User{}, a.UserId.RecordId())
	if err != nil {
		return err
	}
	return database.ExistsById(ctx, db, &Week{}, a.WeekId.RecordId())
}

func (a *Availability) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{a.UserId}
}

func (a *Availability) getTeam(ctx context.Context, db database.Provider) (*Team, error) {
	week, exists, err := database.GetOneById(ctx, db, &Week{}, a.WeekId.RecordId())
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("week with ID %s does not exist", a.WeekId.RecordId())
	}

	draft, err := database.GetExistingRecordById(ctx, db, &Draft{}, week.DraftId.RecordId())
	if err != nil {
		return nil, err
	}

	season, err := draft.GetSeason(ctx, db)
	if err != nil {
		return nil, err
	}

	if season == nil {
		return nil, fmt.Errorf("draft %s (from week %s) does not have an assigned season", draft.ID, a.WeekId)
	}

	seasonMember, err := season.IsUserIdASeasonParticipant(ctx, db, a.UserId)
	if err != nil {
		return nil, err
	}
	if !seasonMember {
		return nil, fmt.Errorf("user %s is not a participant in season %s", a.UserId.String(), season.ID)
	}

	teams, err := season.GetTeams(ctx, db)
	if err != nil {
		return nil, err
	}

	for _, team := range teams {
		isMember, err := team.IsTeamMember(ctx, db, a.UserId)
		if err != nil {
			return nil, err
		}
		if isMember {
			return team, nil
		}
	}

	return nil, fmt.Errorf("user %s was not found on any teams in season %s", a.UserId, season.ID)
}

func (a *Availability) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	team, err := a.getTeam(ctx, db)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	members, err := team.GetMembers(ctx, db)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return members
}

func GetAvailabilityForUser(ctx context.Context, db database.Provider, userId database.UserId, draftId DraftId) ([]*Availability, error) {
	weeks, err := database.GetAllWhere[*Week](ctx, db, func(_ context.Context, c *Week) bool {
		return c.DraftId == draftId
	})

	if err != nil {
		return nil, err
	}

	return database.GetAllWhere[*Availability](ctx, db, func(_ context.Context, c *Availability) bool {
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

func (a *Availability) NewRecord() database.CrudRecord {
	return new(Availability)
}

func GetAvailabilityForTeam(ctx context.Context, db database.Provider, teamId TeamId, draftId DraftId) (map[database.UserId][]*Availability, error) {
	output := make(map[database.UserId][]*Availability)
	team, err := database.GetExistingRecordById(ctx, db, &Team{}, teamId.RecordId())
	if err != nil {
		return nil, err
	}
	members, err := team.GetMembers(ctx, db)
	if err != nil {
		return nil, err
	}
	for _, member := range members {
		availability, err := GetAvailabilityForUser(ctx, db, member, draftId)
		if err != nil {
			return nil, err
		}
		output[member] = availability
	}
	return output, nil
}
