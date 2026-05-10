package model

import (
	"context"
	"fmt"
	"intraclub/common"
	"time"
)

type TeamId common.RecordId

func (t TeamId) RecordId() common.RecordId {
	return common.RecordId(t)
}

func (t TeamId) String() string {
	return t.RecordId().String()
}

type TeamRole string

const (
	TeamRoleCaptain   TeamRole = "captain"
	TeamRoleCoCaptain TeamRole = "co_captain"
	TeamRoleMember    TeamRole = "member"
)

type TeamAssignment struct {
	ID        common.RecordId `json:"id"`
	TeamId    TeamId          `json:"team_id"`
	UserId    UserId          `json:"user_id"`
	Role      TeamRole        `json:"role"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *time.Time      `json:"deleted_at"`
}

func (a *TeamAssignment) GetOwner() common.RecordId {
	return common.InvalidRecordId
}

func (a *TeamAssignment) SetOwner(recordId common.RecordId) {
	// No owner field to set
}

func (a *TeamAssignment) EditableBy(ctx context.Context, db common.DatabaseProvider) []common.RecordId {
	// Team assignments are editable by captains/co-captains of team
	team, exists, err := common.GetOneById(ctx, db, &Team{}, a.TeamId.RecordId())
	if err != nil || !exists {
		return []common.RecordId{}
	}
	return team.EditableBy(ctx, db)
}

func (a *TeamAssignment) AccessibleTo(ctx context.Context, db common.DatabaseProvider) []common.RecordId {
	// Team assignments are accessible to team members
	team, exists, err := common.GetOneById(ctx, db, &Team{}, a.TeamId.RecordId())
	if err != nil || !exists {
		return []common.RecordId{common.EveryoneRecordId}
	}
	return team.AccessibleTo(ctx, db)
}

func (a *TeamAssignment) StaticallyValid() error {
	if a.Role != TeamRoleCaptain && a.Role != TeamRoleCoCaptain && a.Role != TeamRoleMember {
		return fmt.Errorf("invalid team role: %s", a.Role)
	}
	return nil
}

func (a *TeamAssignment) DynamicallyValid(ctx context.Context, db common.DatabaseProvider) error {
	if err := common.ExistsById(ctx, db, &Team{}, a.TeamId.RecordId()); err != nil {
		return err
	}
	if err := common.ExistsById(ctx, db, &User{}, a.UserId.RecordId()); err != nil {
		return err
	}
	return nil
}

func (a *TeamAssignment) Type() string {
	return "team_assignment"
}

func (a *TeamAssignment) GetId() common.RecordId {
	return a.ID
}

func (a *TeamAssignment) SetId(id common.RecordId) {
	a.ID = id
}

func (a *TeamAssignment) PreCreate(db common.DatabaseProvider) error {
	return nil
}

func (a *TeamAssignment) PostCreate() error {
	return nil
}

func (a *TeamAssignment) PreUpdate() error {
	return nil
}

func (a *TeamAssignment) PostUpdate() error {
	return nil
}

func (a *TeamAssignment) PreDelete(_ context.Context, db common.DatabaseProvider) error {
	return nil
}

func (a *TeamAssignment) PostDelete() error {
	return nil
}

func (a *TeamAssignment) Timestamps() (time.Time, time.Time, *time.Time) {
	return a.CreatedAt, a.UpdatedAt, a.DeletedAt
}

func (a *TeamAssignment) SetCreatedAt(createdAt time.Time) {
	a.CreatedAt = createdAt
}

func (a *TeamAssignment) SetUpdatedAt(updatedAt time.Time) {
	a.UpdatedAt = updatedAt
}

func (a *TeamAssignment) SetDeletedAt(deletedAt *time.Time) {
	a.DeletedAt = deletedAt
}

func (a *TeamAssignment) BlankRecord() common.CrudRecord {
	return new(TeamAssignment)
}

type Team struct {
	ID          TeamId              `json:"id"`
	Name        string              `json:"name"`
	Color       TeamColor           `json:"color"`
	RatingsMap  map[UserId]RatingId `json:"ratings_map"`
	tempCaptain UserId              `json:"-"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	DeletedAt   *time.Time          `json:"deleted_at"`
}

func (t *Team) GetOwner() common.RecordId {
	return common.InvalidRecordId
}

func (t *Team) SetOwner(recordId common.RecordId) {
	// don't need to do anything as Captain will not necessarily
	// be the same as the RecordId that was passed into the
	// Create request for this type. The Captain for a given
	// Team will be set after creation via the draft initialization
}

func NewTeam() *Team {
	return &Team{
		RatingsMap: make(map[UserId]RatingId),
	}
}

func NewDefaultTeam(captain UserId, name string) *Team {
	team := NewTeam()
	team.Name = name
	team.Color = TeamColor{
		Name: "Unset",
		Hex:  "000000",
	}
	team.tempCaptain = captain
	return team
}

func (t *Team) getAssignments(ctx context.Context, db common.DatabaseProvider) ([]*TeamAssignment, error) {
	filter := func(_ context.Context, a *TeamAssignment) bool {
		return a.TeamId == t.ID
	}
	return common.GetAllWhere[*TeamAssignment](ctx, db, filter)
}

func (t *Team) GetCaptain(ctx context.Context, db common.DatabaseProvider) (UserId, error) {
	assignments, err := t.getAssignments(ctx, db)
	if err != nil {
		return UserId(0), err
	}
	for _, a := range assignments {
		if a.Role == TeamRoleCaptain {
			return a.UserId, nil
		}
	}
	return UserId(0), fmt.Errorf("no captain found for team %s", t.ID)
}

func (t *Team) GetCoCaptains(ctx context.Context, db common.DatabaseProvider) ([]UserId, error) {
	assignments, err := t.getAssignments(ctx, db)
	if err != nil {
		return nil, err
	}
	coCaptains := make([]UserId, 0)
	for _, a := range assignments {
		if a.Role == TeamRoleCoCaptain {
			coCaptains = append(coCaptains, a.UserId)
		}
	}
	return coCaptains, nil
}

func (t *Team) GetMembers(ctx context.Context, db common.DatabaseProvider) ([]UserId, error) {
	assignments, err := t.getAssignments(ctx, db)
	if err != nil {
		return nil, err
	}
	members := make([]UserId, 0, len(assignments))
	for _, a := range assignments {
		members = append(members, a.UserId)
	}
	return members, nil
}

func (t *Team) IsTeamMember(ctx context.Context, db common.DatabaseProvider, u UserId) (bool, error) {
	members, err := t.GetMembers(ctx, db)
	if err != nil {
		return false, err
	}
	for _, member := range members {
		if member == u {
			return true, nil
		}
	}
	return false, nil
}

func (t *Team) EditableBy(ctx context.Context, db common.DatabaseProvider) []common.RecordId {
	// Captains and co-captains can edit
	ids := make([]common.RecordId, 0)

	// Get captain
	if captain, err := t.GetCaptain(ctx, db); err == nil {
		ids = append(ids, captain.RecordId())
	}

	// Get co-captains
	if coCaptains, err := t.GetCoCaptains(ctx, db); err == nil {
		ids = append(ids, UserIdListToRecordIdList(coCaptains)...)
	}

	return ids
}

func (t *Team) AccessibleTo(ctx context.Context, db common.DatabaseProvider) []common.RecordId {
	// All team members can access
	members, err := t.GetMembers(ctx, db)
	if err != nil {
		return []common.RecordId{common.EveryoneRecordId}
	}
	return UserIdListToRecordIdList(members)
}

func (t *Team) StaticallyValid() error {
	return nil
}

func (t *Team) DynamicallyValid(ctx context.Context, db common.DatabaseProvider) error {
	// If tempCaptain is set, skip captain validation (it will be created in PostCreate)
	if t.tempCaptain != UserId(0) {
		// Still verify tempCaptain is a valid user
		if err := common.ExistsById(ctx, db, &User{}, t.tempCaptain.RecordId()); err != nil {
			return err
		}
		return nil
	}

	// Verify that team has exactly one captain
	captain, err := t.GetCaptain(ctx, db)
	if err != nil {
		return fmt.Errorf("team has no captain: %w", err)
	}

	// Verify captain is a team member
	isMember, err := t.IsTeamMember(ctx, db, captain)
	if err != nil {
		return err
	}
	if !isMember {
		return fmt.Errorf("captain id %s not a team member", captain)
	}

	// Verify co-captains are team members
	coCaptains, err := t.GetCoCaptains(ctx, db)
	if err != nil {
		return err
	}
	for _, coCaptain := range coCaptains {
		isMember, err := t.IsTeamMember(ctx, db, coCaptain)
		if err != nil {
			return err
		}
		if !isMember {
			return fmt.Errorf("co-captain id %s not a team member", coCaptain)
		}
	}

	// Verify all members exist
	members, err := t.GetMembers(ctx, db)
	if err != nil {
		return err
	}
	for _, member := range members {
		err = common.ExistsById(ctx, db, &User{}, member.RecordId())
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Team) Type() string {
	return "team"
}

func (t *Team) GetId() common.RecordId {
	return t.ID.RecordId()
}

func (t *Team) SetId(id common.RecordId) {
	t.ID = TeamId(id)
}

func (t *Team) PreCreate(db common.DatabaseProvider) error {
	return nil
}

func (t *Team) PostCreate(ctx context.Context, db common.DatabaseProvider) error {
	// If a captain was set during creation, create the team assignment for them
	if t.tempCaptain != UserId(0) {
		assignment := &TeamAssignment{
			TeamId: t.ID,
			UserId: t.tempCaptain,
			Role:   TeamRoleCaptain,
		}
		_, err := common.CreateOne(ctx, db, assignment)
		if err != nil {
			return err
		}
		// Clear the temp captain after creating the assignment
		t.tempCaptain = UserId(0)
	}
	return nil
}

func (t *Team) PreUpdate() error {
	return nil
}

func (t *Team) PostUpdate() error {
	return nil
}

func (t *Team) PreDelete(ctx context.Context, db common.DatabaseProvider) error {
	return nil
}

func (t *Team) PostDelete() error {
	return nil
}

func (t *Team) Timestamps() (time.Time, time.Time, *time.Time) {
	return t.CreatedAt, t.UpdatedAt, t.DeletedAt
}

func (t *Team) SetCreatedAt(createdAt time.Time) {
	t.CreatedAt = createdAt
}

func (t *Team) SetUpdatedAt(updatedAt time.Time) {
	t.UpdatedAt = updatedAt
}

func (t *Team) SetDeletedAt(deletedAt *time.Time) {
	t.DeletedAt = deletedAt
}

func (t *Team) BlankRecord() common.CrudRecord {
	return new(Team)
}

func AccessibleByTeamMembers(ctx context.Context, db common.DatabaseProvider, t TeamId) []common.RecordId {
	team, exists, err := common.GetOneById(ctx, db, &Team{}, t.RecordId())
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if !exists {
		fmt.Println("Team does not exist")
		return nil
	}
	members, err := team.GetMembers(ctx, db)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return UserIdListToRecordIdList(members)
}

func EditableByTeamCaptainOrCoCaptains(ctx context.Context, db common.DatabaseProvider, t TeamId) []common.RecordId {
	team, exists, err := common.GetOneById(ctx, db, &Team{}, t.RecordId())
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if !exists {
		fmt.Println("Team does not exist")
		return nil
	}
	return team.EditableBy(ctx, db)
}
