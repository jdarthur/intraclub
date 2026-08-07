package model

import (
	"context"
	"fmt"
	"time"

	"intraclub/database"
)

type TeamId database.RecordId

func (t TeamId) RecordId() database.RecordId {
	return database.RecordId(t)
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
	ID        database.RecordId `json:"id"`
	TeamId    TeamId            `json:"team_id"`
	UserId    database.UserId   `json:"user_id"`
	Role      TeamRole          `json:"role"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	DeletedAt *time.Time        `json:"deleted_at"`
}

func (a *TeamAssignment) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (a *TeamAssignment) SetOwner(userId database.UserId) {
	// No owner field to set
}

func (a *TeamAssignment) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	// Team assignments are editable by captains/co-captains of team
	team, exists, err := database.GetOneById(ctx, db, &Team{}, a.TeamId.RecordId())
	if err != nil || !exists {
		return []database.UserId{}
	}
	return team.EditableBy(ctx, db)
}

func (a *TeamAssignment) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	// Team assignments are accessible to team members
	team, exists, err := database.GetOneById(ctx, db, &Team{}, a.TeamId.RecordId())
	if err != nil || !exists {
		return []database.UserId{database.EveryoneUserId}
	}
	return team.AccessibleTo(ctx, db)
}

func (a *TeamAssignment) StaticallyValid() error {
	if a.Role != TeamRoleCaptain && a.Role != TeamRoleCoCaptain && a.Role != TeamRoleMember {
		return fmt.Errorf("invalid team role: %s", a.Role)
	}
	return nil
}

func (a *TeamAssignment) DynamicallyValid(ctx context.Context, db database.Provider) error {
	if err := database.ExistsById(ctx, db, &Team{}, a.TeamId.RecordId()); err != nil {
		return err
	}
	if err := database.ExistsById(ctx, db, &User{}, a.UserId.RecordId()); err != nil {
		return err
	}
	return nil
}

func (a *TeamAssignment) Type() string {
	return "team_assignment"
}

func (a *TeamAssignment) GetId() database.RecordId {
	return a.ID
}

func (a *TeamAssignment) SetId(id database.RecordId) {
	a.ID = id
}

func (a *TeamAssignment) PreCreate(db database.Provider) error {
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

func (a *TeamAssignment) PreDelete(_ context.Context, db database.Provider) error {
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

func (a *TeamAssignment) NewRecord() database.CrudRecord {
	return new(TeamAssignment)
}

// TeamRating is a join table record that assigns a RatingId to a UserId
// on a particular Team. This replaces the former denormalized
// `Team.RatingsMap` inline map (user -> rating), enabling the relationship
// to be queried/indexed individually and stored in its own collection/table.
//
// API/JSON shape decision: `Team.RatingsMap` (the `ratings_map` JSON field)
// has been removed from `Team`. Team records are not currently exposed via a
// REST CRUD route (see main.go), so removing the field is not a breaking wire
// change; in-process reads can reassemble the relationship rows into the old
// map shape via `Team.GetRatingsMap`. `TeamRating` is stored in its own
// collection (`team_rating`) with FKs to team/user/rating and a natural unique
// constraint on (TeamId, UserId). When the #36 SQLite provider lands, this
// becomes a `team_ratings` table with a migration/backfill.
type TeamRating struct {
	ID       database.RecordId `json:"id"`
	TeamId   TeamId            `json:"team_id"`
	UserId   database.UserId   `json:"user_id"`
	RatingId RatingId          `json:"rating_id"`
}

func (r *TeamRating) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (r *TeamRating) SetOwner(userId database.UserId) {}

func (r *TeamRating) Type() string {
	return "team_rating"
}

func (r *TeamRating) GetId() database.RecordId {
	return r.ID
}

func (r *TeamRating) SetId(id database.RecordId) {
	r.ID = id
}

func (r *TeamRating) StaticallyValid() error {
	return nil
}

func (r *TeamRating) DynamicallyValid(ctx context.Context, db database.Provider) error {
	if err := database.ExistsById(ctx, db, &Team{}, r.TeamId.RecordId()); err != nil {
		return err
	}

	if err := database.ExistsById(ctx, db, &User{}, r.UserId.RecordId()); err != nil {
		return err
	}

	_, err := database.GetExistingRecordById(ctx, db, &Rating{}, r.RatingId.RecordId())
	if err != nil {
		return err
	}

	return nil
}

func (r *TeamRating) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return database.AccessibleToEveryone
}

func (r *TeamRating) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{database.SysAdminUserId}
}

func (r *TeamRating) NewRecord() database.CrudRecord {
	return new(TeamRating)
}

// UniquenessEquivalent enforces the natural unique constraint on (TeamId, UserId):
// a user may only have one assigned rating per team.
func (r *TeamRating) UniquenessEquivalent(other *TeamRating) error {
	if r.TeamId == other.TeamId && r.UserId == other.UserId {
		return fmt.Errorf("user %s already has a rating assigned on team %s", r.UserId, r.TeamId)
	}
	return nil
}

type Team struct {
	ID          TeamId          `json:"id"`
	Name        string          `json:"name"`
	Color       TeamColor       `json:"color"`
	tempCaptain database.UserId `json:"-"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   *time.Time      `json:"deleted_at"`
}

func (t *Team) GetOwner() database.UserId {
	return database.InvalidUserId
}

func (t *Team) SetOwner(userId database.UserId) {
	// don't need to do anything as Captain will not necessarily
	// be the same as the UserId that was passed into the
	// Create request for this type. The Captain for a given
	// Team will be set after creation via the draft initialization
}

func NewTeam() *Team {
	return &Team{}
}

func NewDefaultTeam(captain database.UserId, name string) *Team {
	team := NewTeam()
	team.Name = name
	team.Color = TeamColor{
		Name: "Unset",
		Hex:  "000000",
	}
	team.tempCaptain = captain
	return team
}

func (t *Team) getAssignments(ctx context.Context, db database.Provider) ([]*TeamAssignment, error) {
	filter := func(_ context.Context, a *TeamAssignment) bool {
		return a.TeamId == t.ID
	}
	return database.GetAllWhere[*TeamAssignment](ctx, db, filter)
}

// getRatings returns all TeamRating relationship rows for this team.
func (t *Team) getRatings(ctx context.Context, db database.Provider) ([]*TeamRating, error) {
	filter := func(_ context.Context, r *TeamRating) bool {
		return r.TeamId == t.ID
	}
	return database.GetAllWhere[*TeamRating](ctx, db, filter)
}

// GetRating returns the RatingId assigned to the given UserId on this team,
// or an error if no rating assignment exists for that user.
func (t *Team) GetRating(ctx context.Context, db database.Provider, userId database.UserId) (RatingId, error) {
	ratings, err := t.getRatings(ctx, db)
	if err != nil {
		return 0, err
	}
	for _, r := range ratings {
		if r.UserId == userId {
			return r.RatingId, nil
		}
	}
	return 0, fmt.Errorf("no rating assigned for user %s on team %s", userId, t.ID)
}

// GetRatingsMap reassembles the relationship rows into a user->rating map.
// This preserves the former Team.RatingsMap shape for in-process reads.
func (t *Team) GetRatingsMap(ctx context.Context, db database.Provider) (map[database.UserId]RatingId, error) {
	ratings, err := t.getRatings(ctx, db)
	if err != nil {
		return nil, err
	}
	result := make(map[database.UserId]RatingId, len(ratings))
	for _, r := range ratings {
		result[r.UserId] = r.RatingId
	}
	return result, nil
}

func (t *Team) GetCaptain(ctx context.Context, db database.Provider) (database.UserId, error) {
	assignments, err := t.getAssignments(ctx, db)
	if err != nil {
		return database.InvalidUserId, err
	}
	for _, a := range assignments {
		if a.Role == TeamRoleCaptain {
			return a.UserId, nil
		}
	}
	return database.InvalidUserId, fmt.Errorf("no captain found for team %s", t.ID)
}

func (t *Team) GetCoCaptains(ctx context.Context, db database.Provider) ([]database.UserId, error) {
	assignments, err := t.getAssignments(ctx, db)
	if err != nil {
		return nil, err
	}
	coCaptains := make([]database.UserId, 0)
	for _, a := range assignments {
		if a.Role == TeamRoleCoCaptain {
			coCaptains = append(coCaptains, a.UserId)
		}
	}
	return coCaptains, nil
}

func (t *Team) GetMembers(ctx context.Context, db database.Provider) ([]database.UserId, error) {
	assignments, err := t.getAssignments(ctx, db)
	if err != nil {
		return nil, err
	}
	members := make([]database.UserId, 0, len(assignments))
	for _, a := range assignments {
		members = append(members, a.UserId)
	}
	return members, nil
}

func (t *Team) IsTeamMember(ctx context.Context, db database.Provider, u database.UserId) (bool, error) {
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

func (t *Team) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	// Captains and co-captains can edit
	ids := make([]database.UserId, 0)

	// Get captain
	if captain, err := t.GetCaptain(ctx, db); err == nil {
		ids = append(ids, captain)
	}

	// Get co-captains
	if coCaptains, err := t.GetCoCaptains(ctx, db); err == nil {
		ids = append(ids, coCaptains...)
	}

	return ids
}

func (t *Team) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	// All team members can access
	members, err := t.GetMembers(ctx, db)
	if err != nil {
		return []database.UserId{database.EveryoneUserId}
	}
	return members
}

func (t *Team) StaticallyValid() error {
	return nil
}

func (t *Team) DynamicallyValid(ctx context.Context, db database.Provider) error {
	// If tempCaptain is set, skip captain validation (it will be created in PostCreate)
	if t.tempCaptain != database.InvalidUserId {
		// Still verify tempCaptain is a valid user
		if err := database.ExistsById(ctx, db, &User{}, t.tempCaptain.RecordId()); err != nil {
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
		err = database.ExistsById(ctx, db, &User{}, member.RecordId())
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Team) Type() string {
	return "team"
}

func (t *Team) GetId() database.RecordId {
	return t.ID.RecordId()
}

func (t *Team) SetId(id database.RecordId) {
	t.ID = TeamId(id)
}

func (t *Team) PreCreate(db database.Provider) error {
	return nil
}

func (t *Team) PostCreate(ctx context.Context, db database.Provider) error {
	// If a captain was set during creation, create the team assignment for them
	if t.tempCaptain != database.InvalidUserId {
		assignment := &TeamAssignment{
			TeamId: t.ID,
			UserId: t.tempCaptain,
			Role:   TeamRoleCaptain,
		}
		_, err := database.CreateOne(ctx, db, assignment)
		if err != nil {
			return err
		}
		// Clear the temp captain after creating the assignment
		t.tempCaptain = database.InvalidUserId
	}
	return nil
}

func (t *Team) PreUpdate() error {
	return nil
}

func (t *Team) PostUpdate() error {
	return nil
}

func (t *Team) PreDelete(ctx context.Context, db database.Provider) error {
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

func (t *Team) NewRecord() database.CrudRecord {
	return new(Team)
}

func AccessibleByTeamMembers(ctx context.Context, db database.Provider, t TeamId) []database.UserId {
	team, exists, err := database.GetOneById(ctx, db, &Team{}, t.RecordId())
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
	return members
}

func EditableByTeamCaptainOrCoCaptains(ctx context.Context, db database.Provider, t TeamId) []database.UserId {
	team, exists, err := database.GetOneById(ctx, db, &Team{}, t.RecordId())
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
