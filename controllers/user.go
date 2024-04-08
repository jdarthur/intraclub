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

func (u UserController) ValidateRequest(c *gin.Context, isUpdate bool, db common.DbProvider) (common.CrudRecord, error) {

	m := &model.User{}
	err := c.Bind(m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (u UserController) GetAllFilter(c *gin.Context) (map[string]interface{}, error) {
	return nil, nil
}
