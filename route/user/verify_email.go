package user

import (
	"context"
	"errors"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/model"
)

// VerifyEmailBody is the request body for the VerifyEmail route
type VerifyEmailBody struct {
	Token string `json:"token"`
}

func (v *VerifyEmailBody) StaticallyValid() error {
	if v.Token == "" {
		return errors.New("token must be provided")
	}
	return nil
}

// VerifyEmail allows a user to verify their email address by submitting
// an EmailToken that was sent during self-registration
type VerifyEmail struct{}

func (v VerifyEmail) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, BaseRoute + "/verify_email"
}

func (v VerifyEmail) RequestBody() (*VerifyEmailBody, bool) {
	return &VerifyEmailBody{}, true
}

func (v VerifyEmail) Handler(req api.Request[*VerifyEmailBody]) (any, int, error) {
	if req.Token != nil {
		return nil, http.StatusBadRequest, errors.New("token must not be passed into verify email route")
	}

	_, err := verifyEmail(req.Context, req.DatabaseProvider, req.Body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return nil, http.StatusOK, nil
}

func verifyEmail(ctx context.Context, db database.Provider, body *VerifyEmailBody) (*model.User, error) {
	if body.Token == "" {
		return nil, errors.New("token must be provided")
	}

	// Find the EmailToken by its token value
	whereFunc := func(ctx context.Context, t *model.EmailToken) bool { return t.Token == body.Token }
	tokens, err := database.GetAllWhere[*model.EmailToken](ctx, db, whereFunc)
	if err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, errors.New("invalid token")
	}

	token := tokens[0]

	// Get the user associated with this token
	user, err := database.GetExistingRecordById(ctx, db, &model.User{}, token.UserId.RecordId())
	if err != nil {
		return nil, err
	}

	// Mark the user as verified
	user.Verified = true
	if err := database.UpdateOne(ctx, db, user); err != nil {
		return nil, err
	}

	// Delete the token so it can't be reused
	if _, _, err := database.DeleteOneById(ctx, db, &model.EmailToken{}, token.ID); err != nil {
		return nil, err
	}

	return user, nil
}
