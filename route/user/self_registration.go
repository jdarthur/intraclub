package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"intraclub/api"
	"intraclub/database"
	"intraclub/mailer"
	"intraclub/model"
)

// SelfRegister allows a user to self-register to the system
type SelfRegister struct {
	BaseURL string
}

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

	appURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}

	cfg := mailer.Config{
		FromDomain: appURL.Host,
		Hostname:   "mail." + appURL.Host,
	}
	m, err := mailer.New(cfg)
	if err != nil {
		return err
	}

	message := newEmailMessage(u.Email, t, c.BaseURL)
	return m.Send(ctx, message)
}

func newEmailMessage(addr model.EmailAddress, t *model.EmailToken, baseURL string) mailer.Message {
	verifyURL := fmt.Sprintf("%s/verify?token=%s", baseURL, t.Token)
	body := fmt.Sprintf(
		`<!DOCTYPE html>
<html>
<head>
<style>
  body { font-family: Arial, sans-serif; background: #f4f4f4; padding: 40px; }
  .container { max-width: 600px; margin: auto; background: #fff; padding: 30px; border-radius: 8px; }
  .button { display: inline-block; padding: 14px 28px; background: #4CAF50; color: #fff; text-decoration: none; border-radius: 4px; }
</style>
</head>
<body>
  <div class="container">
    <h2>Verify Your Email</h2>
    <p>Thank you for signing up! Please click the button below to verify your email address and activate your account.</p>
    <p><a href="%s" class="button">Verify Email</a></p>
    <p>If the button doesn't work, copy and paste this link into your browser:</p>
    <p>%s</p>
  </div>
</body>
</html>`,
		verifyURL, verifyURL,
	)
	return mailer.Message{
		From:    "noreply@rcintra.club",
		To:      []string{string(addr)},
		Subject: "Verify your account",
		Text:    body,
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
