package controllers

import (
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
	"net/http"
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

func GetWeeksForLeague(c *gin.Context) {
	v := common.IdValidation(c, &model.League{}, common.GlobalDbProvider)
	if v == nil {
		return
	}

	league := v.(*model.League)
	weeks, err := league.GetWeeks(common.GlobalDbProvider)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{common.ResourceKey: weeks})
}
