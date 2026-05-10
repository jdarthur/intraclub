package model

import (
	"context"
	"errors"
	"fmt"
	"time"

	"intraclub/common"
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
	DraftId          DraftId            // Draft results for this Season
	ScheduleID       ScheduleId         // ID of the Schedule for this Season
	PlayoffStructure PlayoffStructureId // ID of the PlayoffStructure for the Season
	Owner            UserId             // commissioner who owns this season
}

func (s *Season) GetOwner() common.RecordId {
	return s.Owner.RecordId()
}

func (s *Season) UniquenessEquivalent(other *Season) error {
	if s.DraftId == other.DraftId {
		return fmt.Errorf("duplicate season for draft ID %s", s.DraftId)
	}
	return nil
}

func (s *Season) SetOwner(recordId common.RecordId) {
	s.Owner = UserId(recordId)
}

// NewSeason creates a new empty Season record.
// Note: The season must be persisted and commissioners/teams must be added
// before the season is usable.
func NewSeason() *Season {
	return &Season{}
}

// StaticallyValid checks the basic validity of the Season record,
// ensuring the name is set and start time is valid.

func (s *Season) EditableBy(ctx context.Context, db common.DatabaseProvider) []common.RecordId {
	commissioners, err := s.GetCommissioners(ctx, db)
	if err != nil {
		return nil
	}
	return UserIdListToRecordIdList(commissioners)
}

func (s *Season) AccessibleTo(ctx context.Context, db common.DatabaseProvider) []common.RecordId {
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

// StaticallyValid checks the basic validity of the Season record,
// ensuring the name is set and start time is valid.
func (s *Season) StaticallyValid() error {
	if s.Name == "" {
		return errors.New("season name is empty")
	}

	return s.StartTime.StaticallyValid()
}

// DynamicallyValid validates that all referenced entities (Draft, Facility, Schedule, PlayoffStructure)
// exist in the database. Note: This does not validate that commissioners or teams exist as they
// are stored in separate join tables and should be validated through those models.

// DynamicallyValid validates that all referenced entities (Draft, Facility, Schedule, PlayoffStructure)
// exist in the database. Note: This does not validate that commissioners or teams exist as they
// are stored in separate join tables and should be validated through those models.
func (s *Season) DynamicallyValid(ctx context.Context, db common.DatabaseProvider) error {
	if s.DraftId.RecordId() != common.InvalidRecordId {
		err := common.ExistsById(ctx, db, &Draft{}, s.DraftId.RecordId())
		if err != nil {
			return err
		}
	}

	if s.Facility.RecordId() != common.InvalidRecordId {
		err := common.ExistsById(ctx, db, &Facility{}, s.Facility.RecordId())
		if err != nil {
			return err
		}
	}

	if s.ScheduleID.RecordId() != common.InvalidRecordId {
		err := common.ExistsById(ctx, db, &Schedule{}, s.ScheduleID.RecordId())
		if err != nil {
			return err
		}
	}

	if s.PlayoffStructure.RecordId() != common.InvalidRecordId {
		err := common.ExistsById(ctx, db, &PlayoffStructure{}, s.PlayoffStructure.RecordId())
		if err != nil {
			return err
		}
	}

	return nil
}

// GetDraft retrieves the Draft associated with this season.
func (s *Season) GetDraft(ctx context.Context, db common.DatabaseProvider) (*Draft, error) {
	draft, exists, err := common.GetOneById(ctx, db, &Draft{}, s.DraftId.RecordId())
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("draft with ID %d does not exist", s.DraftId)
	}
	return draft, nil
}

// IsUserIdASeasonParticipant checks if a user is a participant in the season,
// either as a commissioner, late addition, or as a member of a team in the season.

// IsUserIdASeasonParticipant checks if a user is a participant in the season,
// either as a commissioner, late addition, or as a member of a team in the season.
func (s *Season) IsUserIdASeasonParticipant(ctx context.Context, db common.DatabaseProvider, u UserId) (bool, error) {
	commissioners, err := s.GetCommissioners(ctx, db)
	if err != nil {
		return false, err
	}
	for _, commissioner := range commissioners {
		if commissioner == u {
			return true, nil
		}
	}

	lateAdditions, err := s.GetLateAdditions(ctx, db)
	if err != nil {
		return false, err
	}
	for _, lateAdd := range lateAdditions {
		if lateAdd == u {
			return true, nil
		}
	}

	teams, err := s.GetTeams(ctx, db)
	if err != nil {
		return false, err
	}

	for _, team := range teams {
		isMember, err := team.IsTeamMember(ctx, db, u)
		if err != nil {
			return false, err
		}
		if isMember {
			return true, nil
		}
	}
	return false, nil
}

// IsUserIdACommissionerViaDB checks if a user ID is one of the commissioners for this season.
// Returns false if an error occurs during the database query.
func (s *Season) IsUserIdACommissionerViaDB(ctx context.Context, db common.DatabaseProvider, u UserId) bool {
	commissioners, err := s.GetCommissioners(ctx, db)
	if err != nil {
		return false
	}
	for _, commissioner := range commissioners {
		if commissioner == u {
			return true
		}
	}
	return false
}

// GetCommissioners retrieves all commissioner User IDs for this season by querying
// the SeasonCommissioner join table.
func (s *Season) GetCommissioners(ctx context.Context, db common.DatabaseProvider) ([]UserId, error) {
	commissioners, err := common.GetAllWhere[*SeasonCommissioner](ctx, db, func(_ context.Context, c *SeasonCommissioner) bool {
		return c.SeasonId == s.ID
	})
	if err != nil {
		return nil, err
	}
	result := make([]UserId, 0, len(commissioners))
	for _, c := range commissioners {
		result = append(result, c.UserId)
	}
	return result, nil
}

// GetTeams retrieves all Team records for this season by querying the SeasonTeam join table
// and fetching each team from the database.
func (s *Season) GetTeams(ctx context.Context, db common.DatabaseProvider) ([]*Team, error) {
	seasonTeams, err := common.GetAllWhere[*SeasonTeam](ctx, db, func(_ context.Context, st *SeasonTeam) bool {
		return st.SeasonId == s.ID
	})
	if err != nil {
		return nil, err
	}
	result := make([]*Team, 0, len(seasonTeams))
	for _, st := range seasonTeams {
		team, err := common.GetExistingRecordById(ctx, db, &Team{}, st.TeamId.RecordId())
		if err != nil {
			return nil, err
		}
		result = append(result, team)
	}
	return result, nil
}

// GetLateAdditions retrieves all late addition User IDs for this season by querying
// the SeasonLateAddition join table.
func (s *Season) GetLateAdditions(ctx context.Context, db common.DatabaseProvider) ([]UserId, error) {
	lateAdditions, err := common.GetAllWhere[*SeasonLateAddition](ctx, db, func(_ context.Context, sla *SeasonLateAddition) bool {
		return sla.SeasonId == s.ID
	})
	if err != nil {
		return nil, err
	}
	result := make([]UserId, 0, len(lateAdditions))
	for _, l := range lateAdditions {
		result = append(result, l.UserId)
	}
	return result, nil
}

// AddCommissioner creates a new SeasonCommissioner record to add a commissioner to this season.
func (s *Season) AddCommissioner(ctx context.Context, db common.DatabaseProvider, userId UserId) error {
	commissioner := &SeasonCommissioner{
		SeasonId: s.ID,
		UserId:   userId,
	}
	_, err := common.CreateOne(ctx, db, commissioner)
	return err
}

// RemoveCommissioner removes all SeasonCommissioner records for a given user from this season.
func (s *Season) RemoveCommissioner(ctx context.Context, db common.DatabaseProvider, userId UserId) error {
	commissioners, err := common.GetAllWhere[*SeasonCommissioner](ctx, db, func(_ context.Context, c *SeasonCommissioner) bool {
		return c.SeasonId == s.ID && c.UserId == userId
	})
	if err != nil {
		return err
	}
	for _, c := range commissioners {
		_, _, err = common.DeleteOneById(ctx, db, &SeasonCommissioner{}, c.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddTeam creates a new SeasonTeam record to add a team to this season.
func (s *Season) AddTeam(ctx context.Context, db common.DatabaseProvider, teamId TeamId) error {
	seasonTeam := &SeasonTeam{
		SeasonId: s.ID,
		TeamId:   teamId,
	}
	_, err := common.CreateOne(ctx, db, seasonTeam)
	return err
}

// RemoveTeam removes all SeasonTeam records for a given team from this season.
func (s *Season) RemoveTeam(ctx context.Context, db common.DatabaseProvider, teamId TeamId) error {
	seasonTeams, err := common.GetAllWhere[*SeasonTeam](ctx, db, func(_ context.Context, st *SeasonTeam) bool {
		return st.SeasonId == s.ID && st.TeamId == teamId
	})
	if err != nil {
		return err
	}
	for _, st := range seasonTeams {
		_, _, err = common.DeleteOneById(ctx, db, &SeasonTeam{}, st.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddLateAddition creates a new SeasonLateAddition record to add a late addition user to this season.
func (s *Season) AddLateAddition(ctx context.Context, db common.DatabaseProvider, userId UserId) error {
	lateAddition := &SeasonLateAddition{
		SeasonId: s.ID,
		UserId:   userId,
	}
	_, err := common.CreateOne(ctx, db, lateAddition)
	return err
}

// RemoveLateAddition removes all SeasonLateAddition records for a given user from this season.
func (s *Season) RemoveLateAddition(ctx context.Context, db common.DatabaseProvider, userId UserId) error {
	lateAdditions, err := common.GetAllWhere[*SeasonLateAddition](ctx, db, func(_ context.Context, sla *SeasonLateAddition) bool {
		return sla.SeasonId == s.ID && sla.UserId == userId
	})
	if err != nil {
		return err
	}
	for _, l := range lateAdditions {
		_, _, err = common.DeleteOneById(ctx, db, &SeasonLateAddition{}, l.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddTeams creates multiple SeasonTeam records to add teams to this season.
func (s *Season) AddTeams(ctx context.Context, db common.DatabaseProvider, teamIds []TeamId) error {
	for _, teamId := range teamIds {
		err := s.AddTeam(ctx, db, teamId)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddCommissioners creates multiple SeasonCommissioner records to add commissioners to this season.
func (s *Season) AddCommissioners(ctx context.Context, db common.DatabaseProvider, userIds []UserId) error {
	for _, userId := range userIds {
		err := s.AddCommissioner(ctx, db, userId)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddLateAdditions creates multiple SeasonLateAddition records to add late addition users to this season.
func (s *Season) AddLateAdditions(ctx context.Context, db common.DatabaseProvider, userIds []UserId) error {
	for _, userId := range userIds {
		err := s.AddLateAddition(ctx, db, userId)
		if err != nil {
			return err
		}
	}
	return nil
}

// EditableBySeason returns a list of common.RecordId values who can edit a particular
// record based on those who can edit the record's associated Season. This function is
// used as a reusable way to compose the common.CrudRecord.EditableBy() list for record
// types which are downstream of a Season and editable by the commissioners, e.g. a Schedule
// or Week belonging to a Season
func EditableBySeason(ctx context.Context, db common.DatabaseProvider, seasonId SeasonId) []common.RecordId {
	season, err := common.GetExistingRecordById(ctx, db, &Season{}, seasonId.RecordId())
	if err != nil {
		fmt.Println(err) // shouldn't get here, but print an error if so for debugging
		return nil
	}
	return season.EditableBy(ctx, db)
}

// GetTeamCaptains returns the list of captain User IDs for all teams in this season.
func (s *Season) GetTeamCaptains(ctx context.Context, db common.DatabaseProvider) ([]UserId, error) {
	teams, err := s.GetTeams(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("Failed to get teams for season: %s\n", err.Error())
	}
	output := make([]UserId, 0)
	for _, team := range teams {
		captain, err := team.GetCaptain(ctx, db)
		if err != nil {
			return nil, err
		}
		output = append(output, captain)
	}
	return output, nil
}

// IsTeamAssignedToSeason checks if a team is assigned to this season by querying
// the SeasonTeam join table.
func (s *Season) IsTeamAssignedToSeason(ctx context.Context, db common.DatabaseProvider, teamId TeamId) bool {
	teams, err := s.GetTeams(ctx, db)
	if err != nil {
		return false
	}
	for _, t := range teams {
		if t.ID == teamId {
			return true
		}
	}
	return false
}

func (s *Season) BlankRecord() common.CrudRecord {
	return new(Season)
}
