package model

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Token struct {
	UserId string `json:"user_id"`
}

func ParseToken(tokenString string) (*Token, error) {
	token := Token{}
	err := json.Unmarshal([]byte(tokenString), &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func NewToken(userId primitive.ObjectID) *Token {
	return &Token{UserId: userId.Hex()}
}

func (t *Token) ToJwt() (string, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func GetTokenFromContext(c *gin.Context) (*Token, error) {
	token, exists := c.Get("token")
	if !exists {
		return nil, errors.New("token was not present")
	}

	return token.(*Token), nil

}
