package model

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"intraclub/common"
	"net/http"
)

func AsAdminUser(c *gin.Context) {
	token, err := GetTokenFromContext(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userId := token.UserId
	id, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := getUser(common.GlobalDbProvider, id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !user.IsAdmin {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Errorf("user %s is not admin", userId)})
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
