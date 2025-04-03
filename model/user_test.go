package model

import (
	"fmt"
	"intraclub/common"
	"math/rand/v2"
	"testing"
)

func randomEmail() EmailAddress {
	return EmailAddress(fmt.Sprintf("user%d@email.com", rand.Uint64()))
}

func randomPhoneNumber() PhoneNumber {
	base := 100_000_0000
	random := rand.IntN(999_999_999) + base
	return PhoneNumber(fmt.Sprintf("%d", random))
}

func newStoredUser(t *testing.T, db common.DatabaseProvider) *User {
	user := NewUser()
	user.Email = randomEmail()
	user.FirstName = fmt.Sprintf("Test %d", rand.Uint64())
	user.LastName = "User"
	user.PhoneNumber = randomPhoneNumber()

	v, err := common.CreateOne(db, user)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func newSysAdmin(t *testing.T, db common.DatabaseProvider) *User {
	common.SysAdminCheck = IsUserSystemAdministrator
	sysAdmin := newStoredUser(t, db)
	err := sysAdmin.AssignRole(db, SystemAdministrator)
	if err != nil {
		t.Fatal(err)
	}
	return sysAdmin
}

func copyUser(u *User) *User {
	return &User{
		ID:          0,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		PhoneNumber: u.PhoneNumber,
		Email:       u.Email,
	}
}

func TestDuplicateUserEmail(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	user2 := copyUser(user)

	_, err := common.CreateOne(db, user2)
	if err == nil {
		t.Fatal("expected duplicate user error")
	}
	fmt.Println(err)
}

func TestDuplicateUserFirstAndLastName(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	user2 := copyUser(user)
	user2.Email = "new@email.com"

	_, err := common.CreateOne(db, user2)
	if err == nil {
		t.Fatal("expected duplicate user error")
	}
	fmt.Println(err)
}

func TestDuplicateUserPhoneNumber(t *testing.T) {
	db := common.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	user2 := copyUser(user)
	user2.Email = "new@email.com"
	user2.FirstName = "Test12345"

	_, err := common.CreateOne(db, user2)
	if err == nil {
		t.Fatal("expected duplicate user error")
	}
	fmt.Println(err)
}
