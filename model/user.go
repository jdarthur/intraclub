package model

import (
	"fmt"
	"intraclub/common"
	"strings"
)

type User struct {
	ID        string `json:"user_id"`
	IsAdmin   bool   `json:"is_admin"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type listOfUsers []*User

func (l listOfUsers) Length() int {
	return len(l)
}

func (l listOfUsers) Value() interface{} {
	return l
}

func (u *User) ValidateStatic() error {
	if u.FirstName == "" {
		return fmt.Errorf("first name must not be empty")
	}
	if u.LastName == "" {
		return fmt.Errorf("last name must not be empty")
	}
	if u.Email == "" {
		return fmt.Errorf("email must not be empty")
	}

	if !strings.Contains(u.Email, "@") {
		return fmt.Errorf("email must contain an @")
	}

	if u.Email[0] == '@' {
		return fmt.Errorf("email must not start with @")
	}

	if u.Email[len(u.Email)-1] == '@' {
		return fmt.Errorf("email must not end with @")
	}

	if len(strings.Split(u.Email, "@")) > 2 {
		return fmt.Errorf("email must not contain multiple @s")
	}

	return nil
}

func (u *User) ValidateDynamic(provider common.DbProvider) error {
	return nil
}

func (u *User) ListOfRecords() common.ListOfCrudRecords {
	return make(listOfUsers, 0)
}

func (u *User) SetId(id string) {
	u.ID = id
}

func (u *User) GetId() string {
	return u.ID
}

func (u *User) RecordType() string {
	return "user"
}

func (u *User) OneRecord() common.CrudRecord {
	return new(User)
}
