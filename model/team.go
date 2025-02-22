package model

import (
	"fmt"
	"intraclub/common"
)

type TeamId common.RecordId

func (t TeamId) RecordId() common.RecordId {
	return common.RecordId(t)
}

func (t TeamId) String() string {
	return t.RecordId().String()
}

type Team struct {
	ID         TeamId
	Captain    UserId
	CoCaptains []UserId
	Members    []UserId
}

func (t *Team) SetOwner(recordId common.RecordId) {
	// don't need to do anything as Captain will not necessarily
	// be the same as the RecordId that was passed into the
	// Create request for this type. The Captain for a given
	// Team will be set after creation via the draft initialization
}

func NewTeam() *Team {
	return &Team{}
}

func (t *Team) EditableBy(common.DatabaseProvider) []common.RecordId {
	captain := []common.RecordId{t.Captain.RecordId()}
	l := UserIdListToRecordIdList(t.CoCaptains)
	return append(captain, l...)
}

func (t *Team) AccessibleTo(common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{common.EveryoneRecordId}
}

func (t *Team) StaticallyValid() error {
	return nil
}

func (t *Team) DynamicallyValid(db common.DatabaseProvider, existing common.DatabaseValidatable) error {
	err := common.ExistsById(db, &User{}, t.Captain.RecordId())
	if err != nil {
		return err
	}

	if !t.IsTeamMember(t.Captain) {
		return fmt.Errorf("captain ID %s not found in members", t.Captain)
	}

	for _, coCaptain := range t.CoCaptains {
		err = common.ExistsById(db, &User{}, coCaptain.RecordId())
		if err != nil {
			return err
		}

		if !t.IsTeamMember(coCaptain) {
			return fmt.Errorf("co-captain ID %s not found in members", coCaptain)
		}
	}

	for _, member := range t.Members {
		err = common.ExistsById(db, &User{}, member.RecordId())
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Team) IsTeamMember(u UserId) bool {
	for _, member := range t.Members {
		if member == u {
			return true
		}
	}
	return false
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

func AccessibleByTeamMembers(db common.DatabaseProvider, t TeamId) []common.RecordId {
	team, exists, err := common.GetOneById(db, &Team{}, t.RecordId())
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if !exists {
		fmt.Println("Team does not exist")
		return nil
	}
	return UserIdListToRecordIdList(team.Members)
}

func EditableByTeamCaptainOrCoCaptains(db common.DatabaseProvider, t TeamId) []common.RecordId {
	team, exists, err := common.GetOneById(db, &Team{}, t.RecordId())
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if !exists {
		fmt.Println("Team does not exist")
		return nil
	}
	return team.EditableBy(db)
}
