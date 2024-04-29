package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
)

type Token struct {
	UserId string `json:"user_id"`
}

var TokenContextKey = "token"

func (t *Token) ToJwt() (string, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("error marshalling token: %v", err)
	}

	return string(b), nil
}

// GetTokenFromAuthMiddleware retrieves the "token" key that was set on
// the WithToken middleware function. This function must be called after
// WithToken in the order of the middleware chain.
func GetTokenFromAuthMiddleware(c *gin.Context) (*Token, error) {
	token, exists := c.Get(TokenContextKey)
	if !exists {
		return nil, errors.New("token was not found in the gin context")
	}

	return token.(*Token), nil
}
