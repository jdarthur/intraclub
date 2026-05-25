package user

import (
	"context"
	"net/http"
	"testing"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelfRegisterAndVerifyEmail(t *testing.T) {
	ctx := context.Background()
	db := database.NewUnitTestDBProvider()

	user := &model.User{
		FirstName:   "Test",
		LastName:    "User",
		Email:       model.EmailAddress("test@example.com"),
		PhoneNumber: model.PhoneNumber("123-456-7890"),
	}

	var savedToken *model.EmailToken
	created, err := selfRegister(ctx, db, user, func(ctx context.Context, u *model.User, t *model.EmailToken) error {
		savedToken = t
		return nil
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotNil(t, savedToken)
	assert.Equal(t, "test@example.com", string(created.Email))
	assert.False(t, created.Verified)

	// Verify the email using the token
	verifiedUser, err := verifyEmail(ctx, db, &VerifyEmailBody{Token: savedToken.Token})
	require.NoError(t, err)
	require.NotNil(t, verifiedUser)
	assert.True(t, verifiedUser.Verified)

	// Verify the user in the DB is marked as verified
	fetchedUser, err := database.GetExistingRecordById(ctx, db, &model.User{}, created.ID.RecordId())
	require.NoError(t, err)
	assert.True(t, fetchedUser.Verified)

	// Verify the token was deleted
	tokens, err := database.GetAllWhere[*model.EmailToken](ctx, db, func(ctx context.Context, t *model.EmailToken) bool {
		return t.UserId == created.ID
	})
	require.NoError(t, err)
	assert.Len(t, tokens, 0)
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	ctx := context.Background()
	db := database.NewUnitTestDBProvider()

	_, err := verifyEmail(ctx, db, &VerifyEmailBody{Token: "nonexistent-token"})
	require.Error(t, err)
	assert.Equal(t, "invalid token", err.Error())
}

func TestVerifyEmail_EmptyToken(t *testing.T) {
	ctx := context.Background()
	db := database.NewUnitTestDBProvider()

	_, err := verifyEmail(ctx, db, &VerifyEmailBody{Token: ""})
	require.Error(t, err)
	assert.Equal(t, "token must be provided", err.Error())
}

func TestVerifyEmail_TokenUsedOnce(t *testing.T) {
	ctx := context.Background()
	db := database.NewUnitTestDBProvider()

	user := &model.User{
		FirstName:   "Test",
		LastName:    "User",
		Email:       model.EmailAddress("test2@example.com"),
		PhoneNumber: model.PhoneNumber("111-222-3333"),
	}

	var savedToken *model.EmailToken
	_, err := selfRegister(ctx, db, user, func(ctx context.Context, u *model.User, t *model.EmailToken) error {
		savedToken = t
		return nil
	})
	require.NoError(t, err)

	// First verification should succeed
	_, err = verifyEmail(ctx, db, &VerifyEmailBody{Token: savedToken.Token})
	require.NoError(t, err)

	// Second verification with the same token should fail
	_, err = verifyEmail(ctx, db, &VerifyEmailBody{Token: savedToken.Token})
	require.Error(t, err)
	assert.Equal(t, "invalid token", err.Error())
}

func TestVerifyEmail_NonexistentUser(t *testing.T) {
	ctx := context.Background()
	db := database.NewUnitTestDBProvider()

	// Create a user, create a token, then delete the user
	user := &model.User{
		FirstName:   "Ghost",
		LastName:    "User",
		Email:       model.EmailAddress("ghost@example.com"),
		PhoneNumber: model.PhoneNumber("000-000-0000"),
	}
	created, err := database.CreateOne(ctx, db, user)
	require.NoError(t, err)

	token, err := model.CreateEmailTokenForUser(ctx, db, created.ID)
	require.NoError(t, err)

	// Delete the user so the token becomes orphaned
	_, _, err = database.DeleteOneById(ctx, db, &model.User{}, created.ID.RecordId())
	require.NoError(t, err)

	_, err = verifyEmail(ctx, db, &VerifyEmailBody{Token: token.Token})
	require.Error(t, err)
}

func TestVerifyEmail_AlreadyVerified(t *testing.T) {
	ctx := context.Background()
	db := database.NewUnitTestDBProvider()

	user := &model.User{
		FirstName:   "Test",
		LastName:    "User",
		Email:       model.EmailAddress("test3@example.com"),
		PhoneNumber: model.PhoneNumber("444-555-6666"),
		Verified:    true,
	}
	_, err := database.CreateOne(ctx, db, user)
	require.NoError(t, err)

	token, err := model.CreateEmailTokenForUser(ctx, db, user.ID)
	require.NoError(t, err)

	// Verification should still succeed even if user is already verified
	verifiedUser, err := verifyEmail(ctx, db, &VerifyEmailBody{Token: token.Token})
	require.NoError(t, err)
	require.NotNil(t, verifiedUser)
	assert.True(t, verifiedUser.Verified)
}

func TestVerifyEmailBodyStaticallyValid_ValidToken(t *testing.T) {
	body := &VerifyEmailBody{Token: "some-token"}
	err := body.StaticallyValid()
	assert.NoError(t, err)
}

func TestVerifyEmailBodyStaticallyValid_EmptyToken(t *testing.T) {
	body := &VerifyEmailBody{Token: ""}
	err := body.StaticallyValid()
	require.Error(t, err)
	assert.Equal(t, "token must be provided", err.Error())
}

func TestVerifyEmail_Path(t *testing.T) {
	v := VerifyEmail{}
	method, path := v.Path()
	assert.Equal(t, api.HttpMethodPost, method)
	assert.Equal(t, "/user/verify_email", path)
}

func TestVerifyEmail_RequestBody(t *testing.T) {
	v := VerifyEmail{}
	body, usesBody := v.RequestBody()
	assert.True(t, usesBody)
	assert.NotNil(t, body)
	assert.IsType(t, &VerifyEmailBody{}, body)
}

func TestVerifyEmail_Handler_WithAuthToken(t *testing.T) {
	v := VerifyEmail{}
	req := api.Request[*VerifyEmailBody]{
		Body:  &VerifyEmailBody{Token: "some-token"},
		Token: &api.AuthToken{UserId: database.UserId(database.NewRecordId())},
	}
	_, statusCode, err := v.Handler(req)
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, "token must not be passed into verify email route", err.Error())
}

func TestVerifyEmail_Handler_Success(t *testing.T) {
	ctx := context.Background()
	db := database.NewUnitTestDBProvider()

	user := &model.User{
		FirstName:   "Test",
		LastName:    "User",
		Email:       model.EmailAddress("test4@example.com"),
		PhoneNumber: model.PhoneNumber("777-888-9999"),
	}
	created, err := database.CreateOne(ctx, db, user)
	require.NoError(t, err)

	token, err := model.CreateEmailTokenForUser(ctx, db, created.ID)
	require.NoError(t, err)

	v := VerifyEmail{}
	req := api.Request[*VerifyEmailBody]{
		Context:          ctx,
		DatabaseProvider: db,
		Body:             &VerifyEmailBody{Token: token.Token},
	}
	resp, statusCode, handlerErr := v.Handler(req)
	assert.Nil(t, handlerErr)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.Nil(t, resp)
}

func TestVerifyEmail_Handler_EmptyToken(t *testing.T) {
	v := VerifyEmail{}
	req := api.Request[*VerifyEmailBody]{
		Body: &VerifyEmailBody{Token: ""},
	}
	_, statusCode, err := v.Handler(req)
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, "token must be provided", err.Error())
}
