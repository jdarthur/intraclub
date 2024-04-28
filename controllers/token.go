package controllers

import (
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
	"net/http"
	"sync"
)

// OneTimePasswordManager is the primary manager that is used to manage one-time-passwords
// as the auth mechanism behind the API. This involves the creation of OTPs and the validation
// of them at runtime. Once a JWT is issued for a successful OTP query, we will remove it from
// the map and treat that JWT as a valid authentication mechanism for a user until it expires
type OneTimePasswordManager struct {
	Map *sync.Map
}

// GetOneTimePassword retrieves a one-time password from the map. If it is not
// found, we will return an error message to the caller.
func (m *OneTimePasswordManager) GetOneTimePassword(userId string) (model.OneTimePassword, error) {
	v, ok := m.Map.Load(userId)
	if !ok {
		return model.OneTimePassword{}, common.ApiError{
			References: userId,
			Code:       common.UserHasNoActiveOneTimePasswords,
		}
	}

	return v.(model.OneTimePassword), nil
}

// DeleteOneTimePassword removes a one-time password from the manager. This is
// called after use and prevents an OTP from being reused
func (m *OneTimePasswordManager) DeleteOneTimePassword(userId string) {
	m.Map.Delete(userId)
}

// CreateOneTimePassword creates a random one-time password. This will be used
// in a emailed "magic link" allowing a user to sign in and be given a JWT without
// using or storing any passwords on the server-side
func (m *OneTimePasswordManager) CreateOneTimePassword(username string) (model.OneTimePassword, error) {

	// create a random otp
	otp, err := model.NewOneTimePassword(username)
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
		return common.ApiError{
			References: username,
			Code:       common.InvalidUuidForUsername,
		}
	}

	// uuid/username combination is no longer valid after use
	m.DeleteOneTimePassword(username)

	return nil
}

func (m *OneTimePasswordManager) GetToken(c *gin.Context) {
	req := model.OneTimePassword{}
	err := c.Bind(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid body for GetToken"})
		return
	}

	user, err := model.GetUserByEmail(common.GlobalDbProvider, req.Email)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}

	err = m.ValidateUUID(req.Email, req.UUID)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}

	token := model.NewToken(user.GetId())
	jwt, err := token.ToJwt()
	if err != nil {
		common.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": jwt})
}

func (m *OneTimePasswordManager) Create(c *gin.Context) {
	req := model.OneTimePassword{}
	err := c.Bind(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body for create"})
		return
	}

	_, err = model.GetUserByEmail(common.GlobalDbProvider, req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	otp, err := m.CreateOneTimePassword(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, otp)
}
