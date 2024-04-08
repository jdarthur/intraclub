package model

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Token struct {
	UserId string `json:"userId"`
}

func ParseToken(token string) (*Token, error) {

	// parse JWT into a &Token

	return &Token{UserId: ""}, nil
}

func NewToken(userId string) *Token {
	return &Token{UserId: userId}
}

func (t *Token) ToJwt() string {
	return ""
}

func WithToken(c *gin.Context) {

	t, exists := c.Get("x-session-token")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No token found"})
		return
	}

	token, err := ParseToken(t.(string))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token was not valid"})
	}

	c.Set("token", token)
	c.Next()
}

func GetTokenFromContext(c *gin.Context) (*Token, error) {
	token, exists := c.Get("token")
	if !exists {
		return nil, errors.New("token was not present")
	}

	return token.(*Token), nil

}
