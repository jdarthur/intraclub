package controllers

import (
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
)

type UserController struct{}

func (u UserController) Model() common.CrudRecord {
	return &model.User{}
}

func (u UserController) ValidateRequest(c common.CrudRecord, isUpdate bool, db common.DbProvider) (common.CrudRecord, error) {
	record := c.(*model.User)
	return record, nil
}

func (u UserController) GetAllFilter(c *gin.Context) (map[string]interface{}, error) {
	return nil, nil
}
