package model

import (
	"context"
	"fmt"
	"strings"

	"intraclub/database"
)

type UserId database.RecordId

func (id UserId) MarshalJSON() ([]byte, error) {
	return id.RecordId().MarshalJSON()
}

func (id UserId) RecordId() database.RecordId {
	return database.RecordId(id)
}

func (id UserId) String() string {
	return id.RecordId().String()
}

func UserIdListToRecordIdList(input []UserId) []database.RecordId {
	output := make([]database.RecordId, 0, len(input))
	for _, id := range input {
		output = append(output, id.RecordId())
	}
	return output
}

func UserIdSliceToUserSlice(ctx context.Context, input []UserId, db database.DatabaseProvider) []*User {
	output := make([]*User, 0, len(input))
	for _, id := range input {
		user, err := database.GetExistingRecordById(ctx, db, &User{}, id.RecordId())
		if err == nil {
			output = append(output, user)
		}
	}
	return output
}

type User struct {
	ID          UserId       `json:"id"`
	FirstName   string       `json:"first_name"`
	LastName    string       `json:"last_name"`
	PhoneNumber PhoneNumber  `json:"phone_number"`
	Email       EmailAddress `json:"email"`
}

func (u *User) GetOwner() database.RecordId {
	// TODO implement me
	panic("implement me")
}

func (u *User) UniquenessEquivalent(other *User) error {
	if u.Email == other.Email {
		return fmt.Errorf("user with email address %s already exists", u.Email)
	}
	if u.FirstName == other.FirstName && u.LastName == other.LastName {
		return fmt.Errorf("user with name %s %s already exists", u.FirstName, u.LastName)
	}
	if u.PhoneNumber != "" && u.PhoneNumber == other.PhoneNumber {
		return fmt.Errorf("user with phone number %s already exists", u.PhoneNumber)
	}
	return nil
}

func (u *User) SetOwner(recordId database.RecordId) {
	// don't need to do anything as User records are self-owned
}

func (u *User) EditableBy(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return []database.RecordId{u.ID.RecordId(), database.SysAdminRecordId}
}

func (u *User) AccessibleTo(ctx context.Context, db database.DatabaseProvider) []database.RecordId {
	return []database.RecordId{database.EveryoneRecordId}
}

func NewUser() *User {
	return &User{}
}

func (u *User) Type() string {
	return "user"
}

func (u *User) GetId() database.RecordId {
	return u.ID.RecordId()
}

func (u *User) SetId(id database.RecordId) {
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

func (u *User) DynamicallyValid(ctx context.Context, db database.DatabaseProvider) error {
	return nil
}

func (u *User) BlankRecord() database.CrudRecord {
	return new(User)
}
