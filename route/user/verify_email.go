package user

import (
	"intraclub/model"
)

type verifyEmailRequest struct {
	EmailAddress model.EmailAddress `json:"email"`
	Token        string             `json:"token"`
}

type VerifyEmail struct{}
