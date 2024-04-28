package model

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
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
		return "", fmt.Errorf("error marshalling token: %v", err)
	}

	return string(b), nil
}

func GetTokenFromContext(c *gin.Context) (*Token, error) {
	token, exists := c.Get("token")
	if !exists {
		return nil, common.ApiError{
			Code: common.TokenWasNotPresentOnGinContext,
		}
	}

	return token.(*Token), nil

}
