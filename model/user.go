package model

import (
	"fmt"
	"intraclub/common"
)

type User struct {
	UserId       string
	FirstName    string
	LastName     string
	Email        string
	Username     string
	Password     string
	PasswordHash string
}

func (u *User) ValidateStatic() error {
	if u.FirstName == "" {
		return fmt.Errorf("first name must not be empty")
	}
	if u.LastName == "" {
		return fmt.Errorf("last name must not be empty")
	}
	if u.Username == "" {
		return fmt.Errorf("username must not be empty")
	}
	if u.Email == "" {
		return fmt.Errorf("email must not be empty")
	}

	return nil
}

func (u *User) ValidateDynamic() error {
	//TODO implement me
	panic("implement me")
}

func (u *User) ListOfRecords() interface{} {
	return make([]User, 0)
}

func (u *User) SetId(id string) {
	u.UserId = id
}

func (u *User) GetId() string {
	return u.UserId
}

func (u *User) RecordType() string {
	return "user"
}

func (u *User) OneRecord() common.CrudRecord {
	return new(User)
}
