package model

import (
	"context"
	"crypto/rand"
	"errors"

	"intraclub/database"
)

// errEmailTokenValueAlreadyExists is an error that indicates that a given token literal
// already exists in the database. This shouldn't ever actually happen as tokens are randomly
// generated with rand.Text(), but it's easy to check for.
var errEmailTokenValueAlreadyExists = errors.New("email token value already exists in DB")

// errEmailTokenAlreadyExists is an error that indicated that a given User already has an EmailToken
// present in the database. This is done as part of a validation to prevent us from creating duplicate
// email tokens and ensuring that we clean up email tokens when re-sending a verification email
var errEmailTokenAlreadyExists = errors.New("conflicting email token already exists for user")

// EmailToken is a random token that is sent via email to a particular User in order to verify
// that they own the EmailAddress that they set on their User
type EmailToken struct {
	ID     database.RecordId `json:"id"`      // store token in DB
	Token  string            `json:"token"`   // random token value
	UserId database.UserId   `json:"user_id"` // User ID that this token pertains to
}

func newEmailTokenForUser(userId database.UserId) *EmailToken {
	return &EmailToken{
		UserId: userId,
		Token:  rand.Text(),
	}
}

// CreateEmailTokenForUser creates an EmailToken for an existing User. If the User already
// has an EmailToken in the DB, we will delete it and create a new one.
func CreateEmailTokenForUser(ctx context.Context, db database.Provider, u database.UserId) (*EmailToken, error) {
	whereFunc := func(ctx context.Context, v *EmailToken) bool { return v.UserId == u }
	tokens, err := database.GetAllWhere[*EmailToken](ctx, db, whereFunc)
	if err != nil {
		return nil, err
	}

	if len(tokens) != 0 {
		_, _, err = database.DeleteOneById(ctx, db, &EmailToken{}, tokens[0].ID)
		if err != nil {
			return nil, err
		}
	}

	token := newEmailTokenForUser(u)
	return database.CreateOne(ctx, db, token)
}

// UniquenessEquivalent ensures that there is only a single EmailToken
// for any given UserId.
func (e *EmailToken) UniquenessEquivalent(other *EmailToken) error {
	if e.UserId == other.UserId {
		return errEmailTokenAlreadyExists
	} else if e.Token == other.Token {
		return errEmailTokenAlreadyExists
	}
	return nil
}

func (e *EmailToken) Type() string {
	return "email_verification_token"
}

func (e *EmailToken) GetId() database.RecordId {
	return e.ID
}

func (e *EmailToken) SetId(id database.RecordId) {
	e.ID = id
}

func (e *EmailToken) EditableBy(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{e.UserId}
}

func (e *EmailToken) AccessibleTo(ctx context.Context, db database.Provider) []database.UserId {
	return []database.UserId{e.UserId}
}

func (e *EmailToken) SetOwner(userId database.UserId) {
	e.UserId = userId
}

func (e *EmailToken) GetOwner() database.UserId {
	return e.UserId
}

func (e *EmailToken) NewRecord() database.CrudRecord {
	return &EmailToken{}
}

// StaticallyValid validates that this EmailToken has a non-empty
// Token value and that the UserId set is non-empty
func (e *EmailToken) StaticallyValid() error {
	if e.Token == "" {
		return errors.New("token is empty")
	} else if e.UserId == database.InvalidUserId {
		return errors.New("invalid user id")
	}
	return nil
}

// DynamicallyValid validates that this EmailToken corresponds to an actually
// existing User in the database.
func (e *EmailToken) DynamicallyValid(ctx context.Context, db database.Provider) error {
	return database.ExistsById[*User](ctx, db, &User{}, e.UserId.RecordId())
}
