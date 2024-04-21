package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"intraclub/model"
	"net/http"
)

var TokenHeaderKey = "x-session-token"
var TokenContextKey = "token"

func WithToken(c *gin.Context) {

	t := c.Request.Header.Get(TokenHeaderKey)
	if t == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No token found"})
		return
	}

	token, err := model.ParseToken(t)
	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token was not valid"})
		return
	}

	c.Set(TokenContextKey, token)
	c.Next()
}

// GetTokenFromAuthMiddleware retrieves the "token" key that was set on
// the WithToken middleware function. This function must be called after
// WithToken in the order of the middleware chain.
func GetTokenFromAuthMiddleware(c *gin.Context) (*model.Token, error) {
	token, exists := c.Get(TokenContextKey)
	if !exists {
		return nil, errors.New("token was not found in the gin context")
	}

	return token.(*model.Token), nil
}
