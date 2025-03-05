package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"intraclub/common"
	"sync"
	"time"
)

var LoginTokenLength = 64
var LoginTokenDefaultExpirationTime = time.Minute * 15
var LoginTokenPurgeTimePeriod = time.Hour * 24 * 7
var TokenPrintfLayout = "2006-01-02_15:04:05"
var BaseUrl = "https://localhost:5000"

type LoginToken struct {
	UserId   UserId
	Expiry   time.Time
	UsedAt   time.Time
	Token    string
	ReturnTo string
}

// NewLoginToken creates a new one-time-use login token for a particular UserId with
// a random token value and which expires LoginTokenDefaultExpirationTime from now
func NewLoginToken(userId UserId) (*LoginToken, error) {
	b := make([]byte, LoginTokenLength)
	n, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	if n != LoginTokenLength {
		return nil, fmt.Errorf("expected %d bytes, got %d", LoginTokenLength, n)
	}

	return &LoginToken{
		UserId: userId,
		Expiry: time.Now().Add(LoginTokenDefaultExpirationTime),
		Token:  hex.EncodeToString(b),
	}, nil
}

func (l *LoginToken) Expired() bool {
	return time.Now().After(l.Expiry)
}

func (l *LoginToken) NeedsPurge() bool {
	return l.UsedOrExpiredForTimePeriod(LoginTokenPurgeTimePeriod)
}

func (l *LoginToken) String() string {

	if l.StaticallyValid() != nil {
		return "invalid token"
	}

	usedString := "no"
	if !l.UsedAt.IsZero() {
		usedString = fmt.Sprintf("yes (at %s)", l.UsedAt.Format(TokenPrintfLayout))
	}
	expiry := l.Expiry.Format(TokenPrintfLayout)

	returnTo := ""
	if l.ReturnTo != "" {
		returnTo = fmt.Sprintf(", return to '%s'", l.ReturnTo)
	}

	return fmt.Sprintf("%s: user=%s, expiry=%v, used=%s%s", l.Token[:16], l.UserId, expiry, usedString, returnTo)
}

// UsedOrExpiredForTimePeriod returns true if a token was used or expired
// X period of time ago, e.g. was used 7 days ago. This is used in the token
// manager to purge tokens from the sync.Map that it uses to store tokens
func (l *LoginToken) UsedOrExpiredForTimePeriod(t time.Duration) bool {
	now := time.Now()
	if now.Sub(l.UsedAt) > t {
		// if token was used X period of time ago, return true
		return true
	}
	if now.Sub(l.Expiry) > t {
		// if token expired X period of time ago, return true
		return true
	}

	// otherwise, don't purge this token
	return false
}

// StaticallyValid validates that this token is:
//   - not expired
//   - non-empty
//   - not already-used
func (l *LoginToken) StaticallyValid() error {
	if l.Expired() {
		return fmt.Errorf("token is expired")
	}
	if l.Token == "" {
		return fmt.Errorf("token is empty")
	}

	if !l.UsedAt.IsZero() {
		return fmt.Errorf("token has already been used")
	}
	if len(l.Token) != LoginTokenLength*2 {
		return fmt.Errorf("token has wrong length (%d, should be %d)", len(l.Token), LoginTokenLength*2)
	}

	return nil
}

// DynamicallyValid validates that this token corresponds to an existing user
func (l *LoginToken) DynamicallyValid(db common.DatabaseProvider) error {
	return common.ExistsById(db, &User{}, l.UserId.RecordId())
}

type StartLoginTokenManager struct {
	tokens sync.Map // map[string]*LoginToken
}

func (m *StartLoginTokenManager) GetToken(token string) (*LoginToken, bool) {
	t, exists := m.tokens.Load(token)
	if exists {
		return t.(*LoginToken), true
	}
	return nil, false
}

func (m *StartLoginTokenManager) AddToken(token *LoginToken) {
	m.tokens.Store(token.Token, token)
}

func (m *StartLoginTokenManager) IsTokenValid(db common.DatabaseProvider, token *LoginToken) error {

	token, exists := m.GetToken(token.Token)
	if !exists {
		return fmt.Errorf("token %s does not exist\n", token.Token)
	}

	err := common.Validate(db, token, nil)
	if err != nil {
		return err
	}

	return nil
}

func (m *StartLoginTokenManager) AuditTokens() {
	toDelete := make([]string, 0)
	m.tokens.Range(func(k, v interface{}) bool {
		token := v.(*LoginToken)
		if token.NeedsPurge() {
			toDelete = append(toDelete, token.Token)
		}
		return true
	})

	for _, token := range toDelete {
		m.DeleteToken(token)
	}
}

func (m *StartLoginTokenManager) DeleteToken(token string) {
	m.tokens.Delete(token)
}

// RequestForLoginToken is a request to generate a new LoginToken
// for a particular UserId. This will create an Email which is sent
// to the User.Email with a clickable link/button containing a token.
// When clicked, this will hit the StartLoginTokenManager.LoginViaToken
// endpoint, attaching the token as a path parameter. If the token is
// valid, then this function will return a JWT which can be attached by
// the web application to future API requests for authenticated calls
type RequestForLoginToken struct {
	UserId   UserId `json:"user_id"`
	ReturnTo string `json:"return_to"`
}

// RequestToken requests that a
func (m *StartLoginTokenManager) RequestToken(db common.DatabaseProvider, req *RequestForLoginToken) (*LoginToken, error) {

	token, err := NewLoginToken(req.UserId)
	if err != nil {
		return nil, err
	}

	err = common.Validate(db, token, nil)
	if err != nil {
		return nil, err
	}

	token.ReturnTo = req.ReturnTo
	m.AddToken(token)

	email, err := m.GenerateTokenEmail(db, token)
	if err != nil {
		return nil, err
	}

	return token, email.Send()
}

func (m *StartLoginTokenManager) GenerateTokenEmail(db common.DatabaseProvider, token *LoginToken) (*Email, error) {
	user, exists, err := common.GetOneById(db, &User{}, token.UserId.RecordId())
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("user '%s' does not exist", token.UserId.String())
	}

	email := NewDefaultEmail()
	email.To = []EmailAddress{user.Email}
	email.Subject = "Log in to Intraclub"
	email.Body = fmt.Sprintf("<a href=%s/create_jwt_from_login_token/%s>Log in</a>", BaseUrl, token.Token)
	return email, nil
}

type LoginResponse struct {
	UserId   UserId `json:"user_id"`
	JWT      string `json:"jwt"`
	ReturnTo string `json:"return_to"`
	Error    error  `json:"error"`
}

func (m *StartLoginTokenManager) LoginViaToken(token string) LoginResponse {
	t, exists := m.GetToken(token)
	if !exists {
		return LoginResponse{
			Error: fmt.Errorf("loginToken '%s' does not exist", token),
		}
	}

	err := t.StaticallyValid()
	if err != nil {
		return LoginResponse{
			Error: err,
		}
	}

	jwt, err := common.GenerateToken(t.UserId.RecordId())
	if err != nil {
		return LoginResponse{
			Error: err,
		}
	}

	t.UsedAt = time.Now()

	return LoginResponse{
		JWT:      jwt,
		UserId:   t.UserId,
		ReturnTo: t.ReturnTo,
	}
}
