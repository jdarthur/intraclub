package model

import (
	"context"
	"testing"

	"intraclub/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmailTokenForUser_GeneratesToken(t *testing.T) {
	userId := database.UserId(database.NewRecordId())
	token := newEmailTokenForUser(userId)

	require.NotNil(t, token)
	assert.NotEmpty(t, token.Token)
	assert.Equal(t, userId, token.UserId)
}

func TestCreateEmailTokenForUser_CreatesToken(t *testing.T) {
	ctx := context.Background()
	db := database.NewUnitTestDBProvider()

	user := newStoredUser(t, db)
	token, err := CreateEmailTokenForUser(ctx, db, user.ID)
	require.NoError(t, err)
	require.NotNil(t, token)
	assert.NotEmpty(t, token.Token)
	assert.Equal(t, user.ID, token.UserId)
}

func TestCreateEmailTokenForUser_DeletesExistingToken(t *testing.T) {
	ctx := context.Background()
	db := database.NewUnitTestDBProvider()

	user := newStoredUser(t, db)

	token1, err := CreateEmailTokenForUser(ctx, db, user.ID)
	require.NoError(t, err)

	token2, err := CreateEmailTokenForUser(ctx, db, user.ID)
	require.NoError(t, err)

	require.NotEqual(t, token1.Token, token2.Token)

	tokens, err := database.GetAllWhere[*EmailToken](ctx, db, func(ctx context.Context, v *EmailToken) bool {
		return v.UserId == user.ID
	})
	require.NoError(t, err)
	assert.Len(t, tokens, 1)
	assert.Equal(t, token2.Token, tokens[0].Token)
}

func TestCreateEmailTokenForUser_NoExistingToken(t *testing.T) {
	ctx := context.Background()
	db := database.NewUnitTestDBProvider()

	user := newStoredUser(t, db)

	tokens, err := database.GetAllWhere[*EmailToken](ctx, db, func(ctx context.Context, v *EmailToken) bool {
		return v.UserId == user.ID
	})
	require.NoError(t, err)
	assert.Len(t, tokens, 0)

	_, err = CreateEmailTokenForUser(ctx, db, user.ID)
	require.NoError(t, err)

	tokens, err = database.GetAllWhere[*EmailToken](ctx, db, func(ctx context.Context, v *EmailToken) bool {
		return v.UserId == user.ID
	})
	require.NoError(t, err)
	assert.Len(t, tokens, 1)
}

func TestEmailTokenUniquenessEquivalent_MatchingUserId(t *testing.T) {
	userId := database.UserId(database.NewRecordId())
	token1 := &EmailToken{UserId: userId, Token: "abc123"}
	token2 := &EmailToken{UserId: userId, Token: "def456"}

	err := token1.UniquenessEquivalent(token2)
	assert.Error(t, err)
	assert.ErrorIs(t, err, errEmailTokenAlreadyExists)
}

func TestEmailTokenUniquenessEquivalent_DifferentUserId(t *testing.T) {
	token1 := &EmailToken{UserId: database.UserId(database.NewRecordId()), Token: "abc123"}
	token2 := &EmailToken{UserId: database.UserId(database.NewRecordId()), Token: "def456"}

	err := token1.UniquenessEquivalent(token2)
	assert.NoError(t, err)
}

func TestEmailTokenUniquenessEquivalent_SameTokenDifferentUser(t *testing.T) {
	token1 := &EmailToken{UserId: database.UserId(database.NewRecordId()), Token: "shared-token"}
	token2 := &EmailToken{UserId: database.UserId(database.NewRecordId()), Token: "shared-token"}

	err := token1.UniquenessEquivalent(token2)
	assert.Error(t, err)
	assert.ErrorIs(t, err, errEmailTokenAlreadyExists)
}

func TestEmailTokenStaticallyValid_Valid(t *testing.T) {
	token := &EmailToken{Token: "valid-token", UserId: database.UserId(database.NewRecordId())}

	err := token.StaticallyValid()
	assert.NoError(t, err)
}

func TestEmailTokenStaticallyValid_EmptyToken(t *testing.T) {
	token := &EmailToken{Token: "", UserId: database.UserId(database.NewRecordId())}

	err := token.StaticallyValid()
	assert.Error(t, err)
	assert.Equal(t, "token is empty", err.Error())
}

func TestEmailTokenStaticallyValid_InvalidUserId(t *testing.T) {
	token := &EmailToken{Token: "valid-token", UserId: database.InvalidUserId}

	err := token.StaticallyValid()
	assert.Error(t, err)
	assert.Equal(t, "invalid user id", err.Error())
}

func TestEmailTokenStaticallyValid_EmptyBoth(t *testing.T) {
	token := &EmailToken{Token: "", UserId: database.InvalidUserId}

	err := token.StaticallyValid()
	assert.Error(t, err)
	assert.Equal(t, "token is empty", err.Error())
}

func TestEmailTokenDynamicallyValid_ValidUser(t *testing.T) {
	ctx := context.Background()
	db := database.NewUnitTestDBProvider()

	user := newStoredUser(t, db)
	token := &EmailToken{Token: "valid-token", UserId: user.ID}

	err := token.DynamicallyValid(ctx, db)
	assert.NoError(t, err)
}

func TestEmailTokenDynamicallyValid_InvalidUser(t *testing.T) {
	ctx := context.Background()
	db := database.NewUnitTestDBProvider()

	token := &EmailToken{Token: "valid-token", UserId: database.UserId(database.NewRecordId())}

	err := token.DynamicallyValid(ctx, db)
	assert.Error(t, err)
}

func TestEmailTokenType(t *testing.T) {
	token := &EmailToken{}
	assert.Equal(t, "email_verification_token", token.Type())
}

func TestEmailTokenGetIdAndSetId(t *testing.T) {
	token := &EmailToken{}
	assert.Equal(t, database.RecordId(0), token.GetId())

	newId := database.NewRecordId()
	token.SetId(newId)
	assert.Equal(t, newId, token.GetId())
}

func TestEmailTokenGetOwnerAndSetOwner(t *testing.T) {
	userId := database.UserId(database.NewRecordId())
	token := &EmailToken{}

	token.SetOwner(userId)
	assert.Equal(t, userId, token.GetOwner())
}

func TestEmailTokenEditableBy(t *testing.T) {
	ctx := context.Background()
	db := database.NewUnitTestDBProvider()

	userId := database.UserId(database.NewRecordId())
	token := &EmailToken{UserId: userId}

	editableBy := token.EditableBy(ctx, db)
	assert.Len(t, editableBy, 1)
	assert.Equal(t, userId, editableBy[0])
}

func TestEmailTokenAccessibleTo(t *testing.T) {
	ctx := context.Background()
	db := database.NewUnitTestDBProvider()

	userId := database.UserId(database.NewRecordId())
	token := &EmailToken{UserId: userId}

	accessibleTo := token.AccessibleTo(ctx, db)
	assert.Len(t, accessibleTo, 1)
	assert.Equal(t, userId, accessibleTo[0])
}

func TestEmailTokenNewRecord(t *testing.T) {
	token := &EmailToken{}
	record := token.NewRecord()

	require.NotNil(t, record)
	emailToken, ok := record.(*EmailToken)
	assert.True(t, ok)
	assert.NotNil(t, emailToken)
}
