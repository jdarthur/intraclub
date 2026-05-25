package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"intraclub/api"
	"intraclub/database"
	"intraclub/mailer"
	"intraclub/model"
)

// SelfRegister allows a user to self-register to the system
type SelfRegister struct{}

func (c SelfRegister) Path() (api.HttpMethod, string) {
	return api.HttpMethodPost, BaseRoute
}

func (c SelfRegister) RequestBody() (*model.User, bool) {
	return &model.User{}, true
}

func (c SelfRegister) Handler(req api.Request[*model.User]) (any, int, error) {
	if req.Body.ID.RecordId() != database.InvalidRecordId {
		return nil, http.StatusBadRequest, errors.New("user ID must not be passed into create user route")
	}
	if req.Token != nil {
		return nil, http.StatusBadRequest, errors.New("token must not be passed into create user route")
	}

	user, err := selfRegister(req.Context, req.DatabaseProvider, req.Body, c.f)

	if err != nil {
		// check for uniqueness constraint error
		errorCode := http.StatusInternalServerError
		if errors.Is(err, database.ErrUniquenessConstraintViolated) {
			errorCode = http.StatusBadRequest
		}
		return nil, errorCode, err
	}
	return user, http.StatusCreated, nil
}

func (c SelfRegister) f(ctx context.Context, u *model.User, t *model.EmailToken) error {
	fmt.Println("send token to email:", u.Email)

	cfg := mailer.Config{
		FromDomain:   "rcintra.club",
		Hostname:     "mail.rcintra.club",
		DKIMSelector: "default",
		DKIMKeyPath:  "/etc/intraclub/dkim.key",
	}
	m, err := mailer.New(cfg)
	if err != nil {
		return err
	}

	message := newEmailMessage(u.Email, t)
	return m.Send(ctx, message)
}

func newEmailMessage(addr model.EmailAddress, t *model.EmailToken) mailer.Message {
	return mailer.Message{
		From:    "noreply@rcintra.club",
		To:      []string{string(addr)},
		Subject: "Verify your account",
		Text:    "",
	}
}

func selfRegister(ctx context.Context, db database.Provider, user *model.User, f sendTokenFunction) (*model.User, error) {
	created, err := database.CreateOne(ctx, db, user)
	if err != nil {
		return nil, err
	}

	token, err := model.CreateEmailTokenForUser(ctx, db, created.ID)
	if err != nil {
		return nil, err
	}

	err = f(ctx, created, token)
	return created, err
}

type sendTokenFunction func(ctx context.Context, u *model.User, t *model.EmailToken) error
