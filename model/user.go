package model

import (
	"context"
	"fmt"
	"strings"

	"intraclub/database"
)

type User struct {
	ID          database.UserId `json:"id"`
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	PhoneNumber PhoneNumber     `json:"phone_number"`
	Email       EmailAddress    `json:"email"`
	Verified    bool            `json:"verified"`
}

func (u *User) GetOwner() database.UserId {
	return u.ID
}

func (u *User) UniquenessEquivalent(other *User) error {
	if u.Email == other.Email {
		return fmt.Errorf("user with email address %s already exists", u.Email)
	} else if u.FirstName == other.FirstName && u.LastName == other.LastName {
		return fmt.Errorf("user with name %s %s already exists", u.FirstName, u.LastName)
	} else if u.PhoneNumber != "" && u.PhoneNumber == other.PhoneNumber {
		return fmt.Errorf("user with phone number %s already exists", u.PhoneNumber)
	}
	return nil
}

func (u *User) SetOwner(userId database.UserId) {
	// don't need to do anything as User records are self-owned
}

func (u *User) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{u.ID, database.SysAdminUserId}
}

func (u *User) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{database.EveryoneUserId}
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
	u.ID = database.UserId(id)
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

func (u *User) DynamicallyValid(ctx context.Context, db database.Provider) error {
	return nil
}

// PostDelete cascades deletion to this user's user_role_assignment rows.
// Without this, deleting a user would orphan those rows (see #97).
func (u *User) PostDelete(ctx context.Context, db database.Provider) error {
	assignments, err := database.GetAllWhere[*UserRoleAssignment](ctx, db, func(_ context.Context, a *UserRoleAssignment) bool {
		return a.UserId == u.ID
	})
	if err != nil {
		return err
	}
	for _, a := range assignments {
		if _, _, err := database.DeleteOneById(ctx, db, a, a.ID); err != nil {
			return err
		}
	}
	return nil
}

func (u *User) NewRecord() database.CrudRecord {
	return new(User)
}
