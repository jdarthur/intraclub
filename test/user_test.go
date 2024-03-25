package test

import (
	"fmt"
	"intraclub/common"
	"intraclub/model"
	"testing"
)

func init() {
	InitializeTestDbProvider()
}

func TestEmailEmpty(t *testing.T) {
	user := &model.User{
		FirstName: "test",
		LastName:  "test2",
	}

	_, err := common.Create(common.GlobalDbProvider, user)
	if err == nil {
		t.Fatal(fmt.Errorf("expected empty email to throw an error"))
	} else {
		ValidateErrorContains(t, err, "email must not be empty")
	}
}

func TestEmailInvalid(t *testing.T) {
	user := &model.User{
		FirstName: "test",
		LastName:  "test2",
		Email:     "email",
	}

	_, err := common.Create(common.GlobalDbProvider, user)
	if err == nil {
		t.Fatal(fmt.Errorf("expected invalid email to throw an error"))
	} else {
		ValidateErrorContains(t, err, "contain an @")
	}
}

func TestEmailStartsWithAt(t *testing.T) {
	user := &model.User{
		FirstName: "test",
		LastName:  "test2",
		Email:     "@email",
	}

	_, err := common.Create(common.GlobalDbProvider, user)
	if err == nil {
		t.Fatal(fmt.Errorf("expected invalid email to throw an error"))
	} else {
		ValidateErrorContains(t, err, "start with @")
	}
}

func TestEmailEndsWithAt(t *testing.T) {
	user := &model.User{
		FirstName: "test",
		LastName:  "test2",
		Email:     "email@",
	}

	_, err := common.Create(common.GlobalDbProvider, user)
	if err == nil {
		t.Fatal(fmt.Errorf("expected invalid email to throw an error"))
	} else {
		ValidateErrorContains(t, err, "end with @")
	}
}

func TestEmailMultiAt(t *testing.T) {
	user := &model.User{
		FirstName: "test",
		LastName:  "test2",
		Email:     "e@email@email.com",
	}

	_, err := common.Create(common.GlobalDbProvider, user)
	if err == nil {
		t.Fatal(fmt.Errorf("expected invalid email to throw an error"))
	} else {
		ValidateErrorContains(t, err, "multiple @s")
	}
}

func TestValidUser(t *testing.T) {
	user := &model.User{
		FirstName: "test",
		LastName:  "test2",
		Email:     "email@email.com",
	}

	_, err := common.Create(common.GlobalDbProvider, user)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFirstNameEmpty(t *testing.T) {
	user := &model.User{
		LastName: "test2",
		Email:    "email@email.com",
	}

	_, err := common.Create(common.GlobalDbProvider, user)
	if err == nil {
		t.Fatal(fmt.Errorf("expected empty first name to throw an error"))
	} else {
		ValidateErrorContains(t, err, "first name must not be empty")
	}
}

func TestLastNameEmpty(t *testing.T) {
	user := &model.User{
		FirstName: "test",
		Email:     "email@email.com",
	}

	_, err := common.Create(common.GlobalDbProvider, user)
	if err == nil {
		t.Fatal(fmt.Errorf("expected empty first name to throw an error"))
	} else {
		ValidateErrorContains(t, err, "last name must not be empty")
	}
}

func createUser() *model.User {
	user := &model.User{
		FirstName: "jim",
		LastName:  "bibby",
		Email:     "jim@bibby.net",
	}

	v, err := common.Create(common.GlobalDbProvider, user)
	if err != nil {
		panic(err)
	}

	return v.(*model.User)
}
