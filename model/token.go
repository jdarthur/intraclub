package model

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
)

func ParseToken(tokenString string) (*common.Token, error) {
	token := common.Token{}
	err := json.Unmarshal([]byte(tokenString), &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func NewToken(userId primitive.ObjectID) *common.Token {
	return &common.Token{UserId: userId.Hex()}
}
