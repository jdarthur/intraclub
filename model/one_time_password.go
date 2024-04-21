package model

import (
	"github.com/google/uuid"
)

// OneTimePassword is the structure used for passwordless authentication
// in the web application.
type OneTimePassword struct {
	Email string `json:"email"`
	UUID  string `json:"uuid"`
}

// NewOneTimePassword creates a new random OneTimePassword for a given user.
// This userId / UUID combination will be sent to the User associated
// with this username in a "magic email" link, and when they click on
// it, they will be issued a JWT that can be used for API auth.
func NewOneTimePassword(email string) (OneTimePassword, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return OneTimePassword{}, err
	}

	return OneTimePassword{
		Email: email,
		UUID:  u.String(),
	}, nil
}
