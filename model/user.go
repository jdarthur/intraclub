package model

import (
	"fmt"
	"intraclub/common"
	"strings"
)

type UserId common.RecordId

func (id UserId) RecordId() common.RecordId {
	return common.RecordId(id)
}

func (id UserId) String() string {
	return id.RecordId().String()
}

func UserIdListToRecordIdList(input []UserId) []common.RecordId {
	output := make([]common.RecordId, 0, len(input))
	for _, id := range input {
		output = append(output, id.RecordId())
	}
	return output
}

type User struct {
	ID          UserId
	FirstName   string
	LastName    string
	PhoneNumber PhoneNumber
	Email       EmailAddress
}

func (u *User) SetOwner(recordId common.RecordId) {
	// don't need to do anything as User records are self-owned
}

func (u *User) EditableBy(common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{u.ID.RecordId(), common.SysAdminRecordId}
}

func (u *User) AccessibleTo(common.DatabaseProvider) []common.RecordId {
	return []common.RecordId{common.EveryoneRecordId}
}

func NewUser() *User {
	return &User{}
}

func (u *User) Type() string {
	return "user"
}

func (u *User) GetId() common.RecordId {
	return u.ID.RecordId()
}

func (u *User) SetId(id common.RecordId) {
	u.ID = UserId(id)
}

func (u *User) TrimValues() {
	u.FirstName = strings.TrimSpace(u.FirstName)
	u.LastName = strings.TrimSpace(u.LastName)
	u.Email = EmailAddress(strings.ToLower(string(u.Email)))
	u.Email = EmailAddress(strings.TrimSpace(string(u.Email)))

	u.PhoneNumber = PhoneNumber(strings.TrimSpace(string(u.PhoneNumber)))
	if u.PhoneNumber != "" {
		u.PhoneNumber = u.PhoneNumber.AddDashes()
	}

}

func (u *User) StaticallyValid() error {
	u.TrimValues()

	if u.FirstName == "" {
		return fmt.Errorf("first name must not be empty")
	}

	if u.LastName == "" {
		return fmt.Errorf("last name must not be empty")
	}

	err := u.Email.StaticallyValid()
	if err != nil {
		return err
	}

	err = u.PhoneNumber.StaticallyValid()
	if err != nil {
		return err
	}

	return nil
}

func (u *User) DynamicallyValid(db common.DatabaseProvider, existing common.DatabaseValidatable) error {
	return nil
}
