package model

import (
	"fmt"
	"github.com/gin-gonic/gin"
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

	user, err := getUser(common.GlobalDbProvider, userId)
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

func UserIsAdmin(provider common.DbProvider, userId string) (bool, error) {
	user, exists, err := common.GetOne(provider, &User{ID: userId})
	if !exists {
		return false, common.RecordDoesNotExist(&User{ID: userId})
	}
	if err != nil {
		return false, err
	}

	return user.(*User).IsAdmin, nil
}

func MarkUserAsAdministrator(provider common.DbProvider, userId string) error {
	user, err := getUser(provider, userId)
	if err != nil {
		return err
	}

	user.IsAdmin = true
	return common.Update(provider, user)
}

func UpdateUserEmail(provider common.DbProvider, userId, email string) error {
	user, err := getUser(provider, userId)
	if err != nil {
		return err
	}

	user.IsAdmin = true
	return common.Update(provider, user)

}

func getUser(provider common.DbProvider, userId string) (*User, error) {
	r, exists, err := common.GetOne(provider, &User{ID: userId})
	if !exists {
		return nil, common.RecordDoesNotExist(&User{ID: userId})
	}
	if err != nil {
		return nil, err
	}

	return r.(*User), nil
}
