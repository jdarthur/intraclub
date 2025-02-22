package model

import (
	"fmt"
	"intraclub/common"
	"math/rand"
	"testing"
)

func randomEmail() EmailAddress {
	return EmailAddress(fmt.Sprintf("user%d@email.com", rand.Uint32()))
}

func newStoredUser(t *testing.T, db common.DatabaseProvider) *User {
	user := NewUser()
	user.Email = randomEmail()
	user.FirstName = "Test"
	user.LastName = "User"

	v, err := common.CreateOne(db, user)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
