package model

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"

	"intraclub/database"
)

func randomEmail() EmailAddress {
	return EmailAddress(fmt.Sprintf("user%d@email.com", rand.Uint64()))
}

func randomPhoneNumber() PhoneNumber {
	base := 100_000_0000
	random := rand.IntN(999_999_999) + base
	return PhoneNumber(fmt.Sprintf("%d", random))
}

func newStoredUser(t *testing.T, db database.Provider) *User {
	user := NewUser()
	user.Email = randomEmail()
	user.FirstName = fmt.Sprintf("Test %d", rand.Uint64())
	user.LastName = "User"
	user.PhoneNumber = randomPhoneNumber()

	v, err := database.CreateOne(context.Background(), db, user)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func newSysAdmin(t *testing.T, db database.Provider) *User {
	database.SysAdminCheck = IsUserSystemAdministrator
	sysAdmin := newStoredUser(t, db)
	err := sysAdmin.AssignRole(context.Background(), db, SystemAdministrator)
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
	db := database.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	user2 := copyUser(user)

	_, err := database.CreateOne(context.Background(), db, user2)
	if err == nil {
		t.Fatal("expected duplicate user error")
	}
	fmt.Println(err)
}

func TestDuplicateUserFirstAndLastName(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	user2 := copyUser(user)
	user2.Email = "new@email.com"

	_, err := database.CreateOne(context.Background(), db, user2)
	if err == nil {
		t.Fatal("expected duplicate user error")
	}
	fmt.Println(err)
}

func TestDuplicateUserPhoneNumber(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	user := newStoredUser(t, db)
	user2 := copyUser(user)
	user2.Email = "new@email.com"
	user2.FirstName = "Test12345"

	_, err := database.CreateOne(context.Background(), db, user2)
	if err == nil {
		t.Fatal("expected duplicate user error")
	}
	fmt.Println(err)
}

func TestUniquenessEquivalent_MatchingEmail(t *testing.T) {
	user1 := &User{Email: "test@example.com", FirstName: "John", LastName: "Doe", PhoneNumber: "123-456-7890"}
	user2 := &User{Email: "test@example.com", FirstName: "Jane", LastName: "Smith", PhoneNumber: "098-765-4321"}

	err := user1.UniquenessEquivalent(user2)
	if err == nil {
		t.Fatal("expected error for matching email")
	}
}

func TestUniquenessEquivalent_DifferentEmail(t *testing.T) {
	user1 := &User{Email: "john@example.com", FirstName: "John", LastName: "Doe", PhoneNumber: "123-456-7890"}
	user2 := &User{Email: "jane@example.com", FirstName: "Jane", LastName: "Smith", PhoneNumber: "098-765-4321"}

	err := user1.UniquenessEquivalent(user2)
	if err != nil {
		t.Fatalf("expected no error for different email/name/phone, got: %v", err)
	}
}

func TestUniquenessEquivalent_MatchingName(t *testing.T) {
	user1 := &User{Email: "john@example.com", FirstName: "John", LastName: "Doe", PhoneNumber: "123-456-7890"}
	user2 := &User{Email: "jane@example.com", FirstName: "John", LastName: "Doe", PhoneNumber: "098-765-4321"}

	err := user1.UniquenessEquivalent(user2)
	if err == nil {
		t.Fatal("expected error for matching first and last name")
	}
}

func TestUniquenessEquivalent_FirstNameOnlyMatch(t *testing.T) {
	user1 := &User{Email: "john@example.com", FirstName: "John", LastName: "Doe", PhoneNumber: "123-456-7890"}
	user2 := &User{Email: "johnsmith@example.com", FirstName: "John", LastName: "Smith", PhoneNumber: "098-765-4321"}

	err := user1.UniquenessEquivalent(user2)
	if err != nil {
		t.Fatalf("expected no error for matching first name only, got: %v", err)
	}
}

func TestUniquenessEquivalent_LastNameOnlyMatch(t *testing.T) {
	user1 := &User{Email: "john@example.com", FirstName: "John", LastName: "Doe", PhoneNumber: "123-456-7890"}
	user2 := &User{Email: "janedoe@example.com", FirstName: "Jane", LastName: "Doe", PhoneNumber: "098-765-4321"}

	err := user1.UniquenessEquivalent(user2)
	if err != nil {
		t.Fatalf("expected no error for matching last name only, got: %v", err)
	}
}

func TestUniquenessEquivalent_MatchingPhoneNumber(t *testing.T) {
	user1 := &User{Email: "john@example.com", FirstName: "John", LastName: "Doe", PhoneNumber: "123-456-7890"}
	user2 := &User{Email: "jane@example.com", FirstName: "Jane", LastName: "Smith", PhoneNumber: "123-456-7890"}

	err := user1.UniquenessEquivalent(user2)
	if err == nil {
		t.Fatal("expected error for matching phone number")
	}
}

func TestUniquenessEquivalent_EmptyFields(t *testing.T) {
	user1 := &User{Email: "", FirstName: "", LastName: "", PhoneNumber: ""}
	user2 := &User{Email: "", FirstName: "", LastName: "", PhoneNumber: ""}

	err := user1.UniquenessEquivalent(user2)
	if err == nil {
		t.Fatal("expected error for matching empty fields")
	}
}
