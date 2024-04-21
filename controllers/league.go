package controllers

import (
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
)

type LeagueController struct{}

func (l LeagueController) Model() common.CrudRecord {
	return &model.League{}
}

func (l LeagueController) ValidateRequest(c common.CrudRecord, isUpdate bool, provider common.DbProvider) (common.CrudRecord, error) {
	return c, nil

}

func (l LeagueController) GetAllFilter(c *gin.Context) (map[string]interface{}, error) {
	return nil, nil
}
