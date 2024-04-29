package model

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
	"strings"
)

type User struct {
	ID        primitive.ObjectID `json:"user_id" bson:"_id"`
	IsAdmin   bool               `json:"is_admin" bson:"is_admin"`
	FirstName string             `json:"first_name" bson:"first_name"`
	LastName  string             `json:"last_name" bson:"last_name"`
	Email     string             `json:"email" bson:"email"`
}

func (u *User) SetUserId(userId string) {}

func (u *User) GetUserId() string {
	return u.ID.Hex()
}

type listOfUsers []*User

func (l listOfUsers) Get(index int) common.CrudRecord {
	return l[index]
}

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

func (u *User) ValidateDynamic(db common.DbProvider, isUpdate bool, previousState common.CrudRecord) error {

	if !isUpdate {
		return common.ValueMustBeGloballyUnique(db, &User{}, "email", u.Email)
	}

	return nil
}

func (u *User) ListOfRecords() common.ListOfCrudRecords {
	return make(listOfUsers, 0)
}

func (u *User) SetId(id primitive.ObjectID) {
	u.ID = id
}

func (u *User) GetId() primitive.ObjectID {
	return u.ID
}

func (u *User) RecordType() string {
	return "user"
}

func (u *User) OneRecord() common.CrudRecord {
	return new(User)
}

func GetUserByEmail(db common.DbProvider, email string) (*User, error) {
	users, err := common.GetAllWhere(db, &User{}, map[string]interface{}{"email": email})
	if err != nil {
		return nil, err
	}

	if users.Length() == 0 {
		return nil, common.ApiError{
			References: email,
			Code:       common.UserWithEmailDoesNotExist,
		}
	}

	if users.Length() > 1 {
		return nil, common.ApiError{
			References: email,
			Code:       common.MultipleUsersExistForEmail,
		}
	}

	return users.Get(0).(*User), nil
}
