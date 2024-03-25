package model

import (
	"fmt"
	"github.com/google/uuid"
	"sync"
)

// OneTimePassword is the structure used for passwordless authentication
// in the web application.
type OneTimePassword struct {
	Username string
	UUID     string
}

// NewOneTimePassword creates a new random OneTimePassword for a given user.
// This username / UUID combination will be sent to the User associated
// with this username in a "magic email" link, and when they click on
// it, they will be issued a JWT that can be used for API auth.
func NewOneTimePassword(username string) (OneTimePassword, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return OneTimePassword{}, err
	}

	return OneTimePassword{
		Username: username,
		UUID:     u.String(),
	}, nil
}

// OneTimePasswordManager is the primary manager that is used to manage one-time-passwords
// as the auth mechanism behind the API. This involves the creation of OTPs and the validation
// of them at runtime. Once a JWT is issued for a successful OTP query, we will remove it from
// the map and treat that JWT as a valid authentication mechanism for a user until it expires
type OneTimePasswordManager struct {
	Map *sync.Map
}

// GetOneTimePassword retrieves a one-time password from the map. If it is not
// found, we will return an error message to the caller.
func (m *OneTimePasswordManager) GetOneTimePassword(username string) (OneTimePassword, error) {
	v, ok := m.Map.Load(username)
	if !ok {
		return OneTimePassword{}, fmt.Errorf("username %s has no active OTPs", username)
	}

	return v.(OneTimePassword), nil
}

// DeleteOneTimePassword removes a one-time password from the manager. This is
// called after use and prevents an OTP from being reused
func (m *OneTimePasswordManager) DeleteOneTimePassword(username string) {
	m.Map.Delete(username)
}

// CreateOneTimePassword creates a random one-time password. This will be used
// in a emailed "magic link" allowing a user to sign in and be given a JWT without
// using or storing any passwords on the server-side
func (m *OneTimePasswordManager) CreateOneTimePassword(username string) (OneTimePassword, error) {

	// create a random otp
	otp, err := NewOneTimePassword(username)
	if err != nil {
		return otp, err
	}

	// store it in the map
	m.Map.Store(username, otp)

	// return the OTP to the caller
	return otp, err
}

// ValidateUUID takes a username/uuid combination from an email link
// and validates that it is the correct combination. This will be used
// to generate a JWT that can be used for API calls on success
func (m *OneTimePasswordManager) ValidateUUID(username, uuid string) (err error) {

	// if username is not in map, return an error
	otp, err := m.GetOneTimePassword(username)
	if err != nil {
		return err
	}

	// if UUID doesn't match, return an error
	if uuid != otp.UUID {
		return fmt.Errorf("UUID %s doesn't match expected value for username %s", uuid, username)
	}

	// uuid/username combination is no longer valid after use
	m.DeleteOneTimePassword(username)

	return nil
}
