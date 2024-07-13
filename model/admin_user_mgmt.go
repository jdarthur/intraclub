package model

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
)

func AsAdminUser(c *gin.Context) {
	token, err := common.GetTokenFromAuthMiddleware(c)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}

	id, err := common.TryParsingObjectId(token.UserId)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}

	user, err := getUser(common.GlobalDbProvider, id)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}

	if !user.IsAdmin {
		common.RespondWithApiError(c, common.ApiError{
			References: []any{token.UserId},
			Code:       common.UserIsNotAdmin,
		})
		return
	}

	c.Next()
}

func UserIsAdmin(provider common.DbProvider, userId primitive.ObjectID) (bool, error) {
	user, err := getUser(provider, userId)
	if err != nil {
		return false, err
	}

	return user.IsAdmin, nil
}

func MarkUserAsAdministrator(provider common.DbProvider, userId primitive.ObjectID) error {
	user, err := getUser(provider, userId)
	if err != nil {
		return err
	}

	user.IsAdmin = true
	return common.Update(provider, user)
}

func UpdateUserEmail(provider common.DbProvider, userId primitive.ObjectID, email string) error {
	user, err := getUser(provider, userId)
	if err != nil {
		return err
	}

	user.Email = email
	return common.Update(provider, user)

}

func getUser(provider common.DbProvider, userId primitive.ObjectID) (*User, error) {
	search := &User{ID: userId}

	r, exists, err := common.GetOne(provider, search)
	if !exists {
		return nil, common.RecordDoesNotExist(search)
	}
	if err != nil {
		return nil, err
	}

	return r.(*User), nil
}
