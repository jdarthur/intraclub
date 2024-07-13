package controllers

import (
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
)

type WeekController struct{}

func (w WeekController) Model() common.CrudRecord {
	return &model.Week{}
}

func (w WeekController) ValidateRequest(c common.CrudRecord, isUpdate bool, provider common.DbProvider) (common.CrudRecord, error) {
	return c, nil
}

func (w WeekController) GetAllFilter(c *gin.Context) (map[string]interface{}, error) {
	return nil, nil
}
