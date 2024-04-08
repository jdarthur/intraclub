package controllers

import (
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
	"net/http"
)

type adminUpdateUserRequest struct {
	UserId string `json:"user_id"`
	Email  string `json:"email"`
}

func UpdateUserId(c *gin.Context) {
	req := &adminUpdateUserRequest{}

	err := c.Bind(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = model.UpdateUserEmail(common.GlobalDbProvider, req.UserId, req.Email)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func SetUserAdmin(c *gin.Context) {
	req := &adminUpdateUserRequest{}

	err := c.Bind(req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = model.UpdateUserEmail(common.GlobalDbProvider, req.UserId, req.Email)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
