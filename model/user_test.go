package model

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"

	"intraclub/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestUserGetOwner(t *testing.T) {
	userId := database.UserId(database.NewRecordId())
	user := &User{ID: userId}

	assert.Equal(t, userId, user.GetOwner())
}

func TestUserSetOwner(t *testing.T) {
	user := &User{ID: database.UserId(123)}
	userId := database.UserId(456)

	user.SetOwner(userId)

	assert.Equal(t, database.UserId(123), user.GetOwner())
}

func TestUserSetOwner_DoesNotChangeId(t *testing.T) {
	originalId := database.UserId(789)
	user := &User{ID: originalId}

	user.SetOwner(database.UserId(456))
	user.SetOwner(database.UserId(999))

	assert.Equal(t, originalId, user.GetOwner())
}

func TestUserEditableBy(t *testing.T) {
	ctx := context.Background()
	db := database.NewUnitTestDBProvider()

	userId := database.UserId(database.NewRecordId())
	user := &User{ID: userId}

	editableBy := user.EditableBy(ctx, db)
	require.Len(t, editableBy, 2)
	assert.Equal(t, userId, editableBy[0])
	assert.Equal(t, database.SysAdminUserId, editableBy[1])
}

func TestUserAccessibleTo(t *testing.T) {
	ctx := context.Background()
	db := database.NewUnitTestDBProvider()

	user := &User{}

	accessibleTo := user.AccessibleTo(ctx, db)
	require.Len(t, accessibleTo, 1)
	assert.Equal(t, database.EveryoneUserId, accessibleTo[0])
}

func TestUserNewRecord(t *testing.T) {
	user := &User{}
	record := user.NewRecord()

	require.NotNil(t, record)
	newUser, ok := record.(*User)
	assert.True(t, ok)
	assert.NotNil(t, newUser)
}

func TestUserTrimValues(t *testing.T) {
	user := &User{
		FirstName:   "  John  ",
		LastName:    "  Doe  ",
		Email:       "  JOHN@Example.COM  ",
		PhoneNumber: "1234567890",
	}

	user.TrimValues()

	assert.Equal(t, "John", user.FirstName)
	assert.Equal(t, "Doe", user.LastName)
	assert.Equal(t, EmailAddress("john@example.com"), user.Email)
	assert.Equal(t, PhoneNumber("123-456-7890"), user.PhoneNumber)
}

func TestUserTrimValues_EmptyPhoneNumber(t *testing.T) {
	user := &User{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john@example.com",
		PhoneNumber: "",
	}

	user.TrimValues()

	assert.Equal(t, PhoneNumber(""), user.PhoneNumber)
}

func TestUserStaticallyValid_Valid(t *testing.T) {
	user := &User{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john@example.com",
		PhoneNumber: "123-456-7890",
	}

	err := user.StaticallyValid()
	assert.NoError(t, err)
}

func TestUserStaticallyValid_EmptyFirstName(t *testing.T) {
	user := &User{
		FirstName:   "",
		LastName:    "Doe",
		Email:       "john@example.com",
		PhoneNumber: "123-456-7890",
	}

	err := user.StaticallyValid()
	assert.Error(t, err)
	assert.Equal(t, "first name must not be empty", err.Error())
}

func TestUserStaticallyValid_EmptyLastName(t *testing.T) {
	user := &User{
		FirstName:   "John",
		LastName:    "",
		Email:       "john@example.com",
		PhoneNumber: "123-456-7890",
	}

	err := user.StaticallyValid()
	assert.Error(t, err)
	assert.Equal(t, "last name must not be empty", err.Error())
}

func TestUserStaticallyValid_InvalidEmail(t *testing.T) {
	user := &User{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "invalid-email",
		PhoneNumber: "123-456-7890",
	}

	err := user.StaticallyValid()
	assert.Error(t, err)
}

func TestUserStaticallyValid_InvalidPhoneNumber(t *testing.T) {
	user := &User{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john@example.com",
		PhoneNumber: "123456789012345",
	}

	err := user.StaticallyValid()
	assert.Error(t, err)
}

func TestUserDynamicallyValid(t *testing.T) {
	ctx := context.Background()
	db := database.NewUnitTestDBProvider()

	user := &User{}
	err := user.DynamicallyValid(ctx, db)
	assert.NoError(t, err)
}

func TestUserPostDeleteCascadesRoleAssignments(t *testing.T) {
	db := database.NewUnitTestDBProvider()
	user := newStoredUser(t, db)

	_, err := database.CreateOne(context.Background(), db, &UserRoleAssignment{
		UserId: user.ID,
		Role:   SystemAdministrator,
	})
	require.NoError(t, err)

	count := func() int {
		rows, err := database.GetAllWhere[*UserRoleAssignment](context.Background(), db, func(_ context.Context, r *UserRoleAssignment) bool {
			return r.UserId == user.ID
		})
		require.NoError(t, err)
		return len(rows)
	}
	require.Equal(t, 1, count())

	_, _, err = database.DeleteOneById(context.Background(), db, &User{}, user.ID.RecordId())
	require.NoError(t, err)

	require.Equal(t, 0, count())
}
