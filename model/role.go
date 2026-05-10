package model

import (
	"context"
	"fmt"
	"intraclub/common"
)

type Role int

const (
	SystemAdministrator Role = iota
	Commissioner
	TeamCaptainOrCoCaptain
	TeamMember
	SeasonMember
	BlurbReporter
	RoleInvalid
)

func (r Role) String() string {
	switch r {
	case SystemAdministrator:
		return "System Administrator"
	case Commissioner:
		return "Commissioner"
	case TeamCaptainOrCoCaptain:
		return "Team Captain or Co-Captain"
	case TeamMember:
		return "Team Member"
	case SeasonMember:
		return "Season Member"
	case BlurbReporter:
		return "Blurb Reporter"
	default:
		return "InvalidScoreCountingType Role"
	}
}

func (r Role) GetReferenceType() common.CrudRecord {
	switch r {
	case SystemAdministrator:
		return nil
	case Commissioner:
		return &Season{}
	case TeamCaptainOrCoCaptain:
		return &Team{}
	case TeamMember:
		return &Team{}
	case SeasonMember:
		return &Season{}
	default:
		return nil
	}
}

func (r Role) Valid() bool {
	return r < RoleInvalid
}

type UserRoleAssignment struct {
	ID          common.RecordId // ID of this assignment
	UserId      UserId          // ID of the User being assigned a role
	Role        Role            // Role being assigned to this user
	ReferenceId common.RecordId // ID of referenced record base on role, e.g. Team ID for a TeamMember role
}

func (u *UserRoleAssignment) GetOwner() common.RecordId {
	return common.InvalidRecordId
}

func (u *UserRoleAssignment) SetOwner(recordId common.RecordId) {
	// don't need to do anything as the Owner field this record type
	// will necessarily be present in the Create request
}

func (u *UserRoleAssignment) EditableBy(ctx context.Context, db common.DatabaseProvider) []common.RecordId {
	return common.SysAdminAndUsers() // only changeable by system administrator
}

func (u *UserRoleAssignment) AccessibleTo(ctx context.Context, db common.DatabaseProvider) []common.RecordId {
	//TODO implement me
	panic("implement me")
}

func (u *UserRoleAssignment) Type() string {
	return "user_role_assignment"
}

func (u *UserRoleAssignment) GetId() common.RecordId {
	return u.ID
}

func (u *UserRoleAssignment) SetId(id common.RecordId) {
	u.ID = id
}

func (u *UserRoleAssignment) StaticallyValid() error {
	if !u.Role.Valid() {
		return fmt.Errorf("role %d is not valid", u.Role)
	}
	return nil
}

func (u *UserRoleAssignment) DynamicallyValid(ctx context.Context, db common.DatabaseProvider) error {
	err := common.ExistsById(ctx, db, &User{}, u.UserId.RecordId())
	if err != nil {
		return err
	}

	referenceType := u.Role.GetReferenceType()
	if referenceType != nil {
		return common.ExistsById(ctx, db, referenceType, u.ReferenceId)
	}

	return nil
}

func (u *UserRoleAssignment) BlankRecord() common.CrudRecord {
	return new(UserRoleAssignment)
}

func IsUserAssignedToTeam(ctx context.Context, db common.DatabaseProvider, userId UserId, teamId TeamId) (bool, error) {
	err := common.ExistsById(ctx, db, &User{}, userId.RecordId())
	if err != nil {
		return false, err
	}

	team, exists, err := common.GetOneById(ctx, db, &Team{}, teamId.RecordId())
	if err != nil {
		return false, err
	}
	if !exists {
		return false, fmt.Errorf("team with ID %s does not exist", teamId)
	}
	return team.IsTeamMember(ctx, db, userId)
}

func IsUserAssignedToSeason(ctx context.Context, db common.DatabaseProvider, userId UserId, seasonId common.RecordId) (bool, error) {
	err := common.ExistsById(ctx, db, &User{}, userId.RecordId())
	if err != nil {
		return false, err
	}

	season, exists, err := common.GetOneById(ctx, db, &Season{}, seasonId)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, fmt.Errorf("season with ID %s does not exist", seasonId)
	}

	return season.IsUserIdASeasonParticipant(ctx, db, userId)
}

func IsUserSystemAdministrator(ctx context.Context, db common.DatabaseProvider, userId common.RecordId) (bool, error) {
	user, exists, err := common.GetOneById(ctx, db, &User{}, userId)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, fmt.Errorf("user with ID %s does not exist", userId)
	}

	return user.HasRole(ctx, db, SystemAdministrator)
}

func (u *User) HasRole(ctx context.Context, db common.DatabaseProvider, role Role) (bool, error) {

	filter := func(_ context.Context, c *UserRoleAssignment) bool {
		return c.UserId == u.ID
	}

	roles, err := common.GetAllWhere[*UserRoleAssignment](ctx, db, filter)

	if err != nil {
		return false, err
	}
	for _, assignment := range roles {
		if assignment.Role == role {
			return true, nil
		}
	}
	return false, nil
}

func (u *User) AssignRole(ctx context.Context, db common.DatabaseProvider, r Role) error {
	return u.AssignRoleWithReference(ctx, db, r, common.InvalidRecordId)
}

func (u *User) AssignRoleWithReference(ctx context.Context, db common.DatabaseProvider, r Role, referenceId common.RecordId) error {
	hasRole, err := u.HasRole(ctx, db, r)
	if err != nil {
		return err
	}
	if hasRole {
		return nil
	}

	if r != SystemAdministrator && referenceId != common.InvalidRecordId {
		err = common.ExistsById(ctx, db, r.GetReferenceType(), referenceId)
		if err != nil {
			return err
		}
	}

	assignment := UserRoleAssignment{
		UserId:      u.ID,
		Role:        r,
		ReferenceId: referenceId,
	}

	_, err = common.CreateOne(ctx, db, &assignment)
	return err
}
