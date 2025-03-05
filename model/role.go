package model

import (
	"fmt"
	"intraclub/common"
)

type Role int

const (
	SystemAdministrator Role = iota
	Commissioner
	TeamCaptainOrCoCaptain
	TeamMember
	LeagueMember
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
	case LeagueMember:
		return "League Member"
	default:
		return "Invalid Role"
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
	case LeagueMember:
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

func (u *UserRoleAssignment) SetOwner(recordId common.RecordId) {
	// don't need to do anything as the Owner field this record type
	// will necessarily be present in the Create request
}

func (u *UserRoleAssignment) EditableBy(common.DatabaseProvider) []common.RecordId {
	return common.SysAdminAndUsers() // only changeable by system administrator
}

func (u *UserRoleAssignment) AccessibleTo(common.DatabaseProvider) []common.RecordId {
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

func (u *UserRoleAssignment) DynamicallyValid(db common.DatabaseProvider, existing common.DatabaseValidatable) error {
	err := common.ExistsById(db, &User{}, u.UserId.RecordId())
	if err != nil {
		return err
	}

	referenceType := u.Role.GetReferenceType()
	if referenceType != nil {
		return common.ExistsById(db, referenceType, u.ReferenceId)
	}

	return nil
}

func IsUserAssignedToTeam(db common.DatabaseProvider, userId UserId, teamId TeamId) (bool, error) {
	err := common.ExistsById(db, &User{}, userId.RecordId())
	if err != nil {
		return false, err
	}

	team, exists, err := common.GetOneById(db, &Team{}, teamId.RecordId())
	if err != nil {
		return false, err
	}
	if !exists {
		return false, fmt.Errorf("team with ID %s does not exist", teamId)
	}
	return team.IsTeamMember(userId), nil
}

func IsUserAssignedToSeason(db common.DatabaseProvider, userId UserId, seasonId common.RecordId) (bool, error) {
	err := common.ExistsById(db, &User{}, userId.RecordId())
	if err != nil {
		return false, err
	}

	season, exists, err := common.GetOneById(db, &Season{}, seasonId)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, fmt.Errorf("season with ID %s does not exist", seasonId)
	}

	return season.IsUserIdASeasonParticipant(db, userId)
}

func IsUserSystemAdministrator(db common.DatabaseProvider, userId common.RecordId) (bool, error) {
	user, exists, err := common.GetOneById(db, &User{}, userId)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, fmt.Errorf("user with ID %s does not exist", userId)
	}

	return user.HasRole(db, SystemAdministrator)
}

func (u *User) HasRole(db common.DatabaseProvider, role Role) (bool, error) {

	filter := func(c *UserRoleAssignment) bool {
		return c.UserId == u.ID
	}

	roles, err := common.GetAllWhere(db, &UserRoleAssignment{}, filter)

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

func (u *User) AssignRole(db common.DatabaseProvider, r Role) error {
	return u.AssignRoleWithReference(db, r, common.InvalidRecordId)
}

func (u *User) AssignRoleWithReference(db common.DatabaseProvider, r Role, referenceId common.RecordId) error {
	hasRole, err := u.HasRole(db, r)
	if err != nil {
		return err
	}
	if hasRole {
		return nil
	}

	assignment := UserRoleAssignment{
		UserId:      u.ID,
		Role:        r,
		ReferenceId: referenceId,
	}

	_, err = common.CreateOne(db, &assignment)
	return err
}
