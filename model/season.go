package model

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"intraclub/database"
)

type SeasonId database.RecordId

func (id SeasonId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id SeasonId) String() string {
	return id.RecordId().String()
}

func (id SeasonId) MarshalJSON() ([]byte, error) {
	return id.RecordId().MarshalJSON()
}

func (id *SeasonId) UnmarshalJSON(bytes []byte) error {
	rid := id.RecordId()
	if err := (*database.RecordId)(&rid).UnmarshalJSON(bytes); err != nil {
		return err
	}
	*id = SeasonId(rid)
	return nil
}

type StartTime time.Time

func NewStartTime(hour int, minute int) StartTime {
	t := time.Date(1, 1, 1, hour, minute, 0, 0, time.UTC)
	return StartTime(t)
}

// MarshalJSON renders the daily kickoff time as a 24-hour "HH:MM" string (e.g.
// "08:30"), matching the shape the create-season endpoint accepts.
func (s StartTime) MarshalJSON() ([]byte, error) {
	return []byte("\"" + time.Time(s).Format("15:04") + "\""), nil
}

func (s *StartTime) UnmarshalJSON(bytes []byte) error {
	raw := strings.Trim(string(bytes), "\"")
	if raw == "" {
		return nil
	}
	t, err := time.Parse("15:04", raw)
	if err != nil {
		return err
	}
	*s = NewStartTime(t.Hour(), t.Minute())
	return nil
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
	ID               SeasonId           `json:"id"`            // unique identifier for this Season
	Name             string             `json:"name"`          // descriptive name for this season, e.g. _Men's Intraclub 2025_
	Facility         FacilityId         `json:"facility"`      // ID of the Facility at which this Season is played
	StartTime        StartTime          `json:"start_time"`    // time of day when the first matches kick off (e.g. _8:30 AM_)
	DraftId          DraftId            `json:"draft_id"`      // Draft results for this Season
	ScheduleID       ScheduleId         `json:"schedule_id"`   // ID of the Schedule for this Season
	PlayoffStructure PlayoffStructureId `json:"playoff_structure"` // ID of the PlayoffStructure for the Season
	Owner            database.UserId    `json:"owner"`         // commissioner who owns this season
}

func (s *Season) GetOwner() database.UserId {
	return s.Owner
}

func (s *Season) UniquenessEquivalent(other *Season) error {
	if s.DraftId == other.DraftId {
		return fmt.Errorf("duplicate season for draft ID %s", s.DraftId)
	}
	return nil
}

func (s *Season) SetOwner(userId database.UserId) {
	s.Owner = userId
}

// NewSeason creates a new empty Season record.
// Note: The season must be persisted and commissioners/teams must be added
// before the season is usable.
func NewSeason() *Season {
	return &Season{}
}

// StaticallyValid checks the basic validity of the Season record,
// ensuring the name is set and start time is valid.

func (s *Season) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	commissioners, err := s.GetCommissioners(ctx, db)
	if err != nil {
		return nil
	}
	return commissioners
}

func (s *Season) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (s *Season) Type() string {
	return "season"
}

func (s *Season) GetId() database.RecordId {
	return s.ID.RecordId()
}

func (s *Season) SetId(id database.RecordId) {
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
func (s *Season) DynamicallyValid(ctx context.Context, db database.Provider) error {
	if s.DraftId.RecordId() != database.InvalidRecordId {
		err := database.ExistsById(ctx, db, &Draft{}, s.DraftId.RecordId())
		if err != nil {
			return err
		}
	}

	if s.Facility.RecordId() != database.InvalidRecordId {
		err := database.ExistsById(ctx, db, &Facility{}, s.Facility.RecordId())
		if err != nil {
			return err
		}
	}

	if s.ScheduleID.RecordId() != database.InvalidRecordId {
		err := database.ExistsById(ctx, db, &Schedule{}, s.ScheduleID.RecordId())
		if err != nil {
			return err
		}
	}

	if s.PlayoffStructure.RecordId() != database.InvalidRecordId {
		err := database.ExistsById(ctx, db, &PlayoffStructure{}, s.PlayoffStructure.RecordId())
		if err != nil {
			return err
		}
	}

	return nil
}

// GetDraft retrieves the Draft associated with this season.
func (s *Season) GetDraft(ctx context.Context, db database.Provider) (*Draft, error) {
	draft, exists, err := database.GetOneById(ctx, db, &Draft{}, s.DraftId.RecordId())
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
func (s *Season) IsUserIdASeasonParticipant(ctx context.Context, db database.Provider, u database.UserId) (bool, error) {
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
func (s *Season) IsUserIdACommissionerViaDB(ctx context.Context, db database.Provider, u database.UserId) bool {
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
func (s *Season) GetCommissioners(ctx context.Context, db database.Provider) ([]database.UserId, error) {
	commissioners, err := database.GetAllWhere[*SeasonCommissioner](ctx, db, func(_ context.Context, c *SeasonCommissioner) bool {
		return c.SeasonId == s.ID
	})
	if err != nil {
		return nil, err
	}
	result := make([]database.UserId, 0, len(commissioners))
	for _, c := range commissioners {
		result = append(result, c.UserId)
	}
	return result, nil
}

// GetTeams retrieves all Team records for this season by querying the SeasonTeam join table
// and fetching each team from the database.
func (s *Season) GetTeams(ctx context.Context, db database.Provider) ([]*Team, error) {
	seasonTeams, err := database.GetAllWhere[*SeasonTeam](ctx, db, func(_ context.Context, st *SeasonTeam) bool {
		return st.SeasonId == s.ID
	})
	if err != nil {
		return nil, err
	}
	result := make([]*Team, 0, len(seasonTeams))
	for _, st := range seasonTeams {
		team, err := database.GetExistingRecordById(ctx, db, &Team{}, st.TeamId.RecordId())
		if err != nil {
			return nil, err
		}
		result = append(result, team)
	}
	return result, nil
}

// GetLateAdditions retrieves all late addition User IDs for this season by querying
// the SeasonLateAddition join table.
func (s *Season) GetLateAdditions(ctx context.Context, db database.Provider) ([]database.UserId, error) {
	lateAdditions, err := database.GetAllWhere[*SeasonLateAddition](ctx, db, func(_ context.Context, sla *SeasonLateAddition) bool {
		return sla.SeasonId == s.ID
	})
	if err != nil {
		return nil, err
	}
	result := make([]database.UserId, 0, len(lateAdditions))
	for _, l := range lateAdditions {
		result = append(result, l.UserId)
	}
	return result, nil
}

// AddCommissioner creates a new SeasonCommissioner record to add a commissioner to this season.
func (s *Season) AddCommissioner(ctx context.Context, db database.Provider, userId database.UserId) error {
	commissioner := &SeasonCommissioner{
		SeasonId: s.ID,
		UserId:   userId,
	}
	_, err := database.CreateOne(ctx, db, commissioner)
	return err
}

// RemoveCommissioner removes all SeasonCommissioner records for a given user from this season.
func (s *Season) RemoveCommissioner(ctx context.Context, db database.Provider, userId database.UserId) error {
	commissioners, err := database.GetAllWhere[*SeasonCommissioner](ctx, db, func(_ context.Context, c *SeasonCommissioner) bool {
		return c.SeasonId == s.ID && c.UserId == userId
	})
	if err != nil {
		return err
	}
	for _, c := range commissioners {
		_, _, err = database.DeleteOneById(ctx, db, &SeasonCommissioner{}, c.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddTeam creates a new SeasonTeam record to add a team to this season.
func (s *Season) AddTeam(ctx context.Context, db database.Provider, teamId TeamId) error {
	seasonTeam := &SeasonTeam{
		SeasonId: s.ID,
		TeamId:   teamId,
	}
	_, err := database.CreateOne(ctx, db, seasonTeam)
	return err
}

// RemoveTeam removes all SeasonTeam records for a given team from this season.
func (s *Season) RemoveTeam(ctx context.Context, db database.Provider, teamId TeamId) error {
	seasonTeams, err := database.GetAllWhere[*SeasonTeam](ctx, db, func(_ context.Context, st *SeasonTeam) bool {
		return st.SeasonId == s.ID && st.TeamId == teamId
	})
	if err != nil {
		return err
	}
	for _, st := range seasonTeams {
		_, _, err = database.DeleteOneById(ctx, db, &SeasonTeam{}, st.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddLateAddition creates a new SeasonLateAddition record to add a late addition user to this season.
func (s *Season) AddLateAddition(ctx context.Context, db database.Provider, userId database.UserId) error {
	lateAddition := &SeasonLateAddition{
		SeasonId: s.ID,
		UserId:   userId,
	}
	_, err := database.CreateOne(ctx, db, lateAddition)
	return err
}

// RemoveLateAddition removes all SeasonLateAddition records for a given user from this season.
func (s *Season) RemoveLateAddition(ctx context.Context, db database.Provider, userId database.UserId) error {
	lateAdditions, err := database.GetAllWhere[*SeasonLateAddition](ctx, db, func(_ context.Context, sla *SeasonLateAddition) bool {
		return sla.SeasonId == s.ID && sla.UserId == userId
	})
	if err != nil {
		return err
	}
	for _, l := range lateAdditions {
		_, _, err = database.DeleteOneById(ctx, db, &SeasonLateAddition{}, l.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddTeams creates multiple SeasonTeam records to add teams to this season.
func (s *Season) AddTeams(ctx context.Context, db database.Provider, teamIds []TeamId) error {
	for _, teamId := range teamIds {
		err := s.AddTeam(ctx, db, teamId)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddCommissioners creates multiple SeasonCommissioner records to add commissioners to this season.
func (s *Season) AddCommissioners(ctx context.Context, db database.Provider, userIds []database.UserId) error {
	for _, userId := range userIds {
		err := s.AddCommissioner(ctx, db, userId)
		if err != nil {
			return err
		}
	}
	return nil
}

// AddLateAdditions creates multiple SeasonLateAddition records to add late addition users to this season.
func (s *Season) AddLateAdditions(ctx context.Context, db database.Provider, userIds []database.UserId) error {
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
func EditableBySeason(ctx context.Context, db database.Provider, seasonId SeasonId) []database.UserId {
	season, err := database.GetExistingRecordById(ctx, db, &Season{}, seasonId.RecordId())
	if err != nil {
		fmt.Println(err) // shouldn't get here, but print an error if so for debugging
		return nil
	}
	return season.EditableBy(ctx, db)
}

// GetTeamCaptains returns the list of captain User IDs for all teams in this season.
func (s *Season) GetTeamCaptains(ctx context.Context, db database.Provider) ([]database.UserId, error) {
	teams, err := s.GetTeams(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("Failed to get teams for season: %s\n", err.Error())
	}
	output := make([]database.UserId, 0)
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
func (s *Season) IsTeamAssignedToSeason(ctx context.Context, db database.Provider, teamId TeamId) bool {
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

// PostDelete cascades deletion to this season's season_commissioner,
// season_late_addition, and season_team join rows. Without this, deleting a
// season would orphan those rows (see #97).
func (s *Season) PostDelete(ctx context.Context, db database.Provider) error {
	commissioners, err := database.GetAllWhere[*SeasonCommissioner](ctx, db, func(_ context.Context, c *SeasonCommissioner) bool {
		return c.SeasonId == s.ID
	})
	if err != nil {
		return err
	}
	for _, c := range commissioners {
		if _, _, err := database.DeleteOneById(ctx, db, c, c.ID); err != nil {
			return err
		}
	}

	lateAdditions, err := database.GetAllWhere[*SeasonLateAddition](ctx, db, func(_ context.Context, c *SeasonLateAddition) bool {
		return c.SeasonId == s.ID
	})
	if err != nil {
		return err
	}
	for _, c := range lateAdditions {
		if _, _, err := database.DeleteOneById(ctx, db, c, c.ID); err != nil {
			return err
		}
	}

	teams, err := database.GetAllWhere[*SeasonTeam](ctx, db, func(_ context.Context, c *SeasonTeam) bool {
		return c.SeasonId == s.ID
	})
	if err != nil {
		return err
	}
	for _, c := range teams {
		if _, _, err := database.DeleteOneById(ctx, db, c, c.ID); err != nil {
			return err
		}
	}

	return nil
}

func (s *Season) NewRecord() database.CrudRecord {
	return new(Season)
}
