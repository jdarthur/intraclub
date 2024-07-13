package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
	"net/http"
	"sort"
)

type LeagueController struct{}

func (l LeagueController) Model() common.CrudRecord {
	return &model.League{}
}

func (l LeagueController) ValidateRequest(c common.CrudRecord, isUpdate bool, provider common.DbProvider) (common.CrudRecord, error) {

	league := c.(*model.League)
	weeks, err := league.GetWeeks(provider)
	if err != nil {
		return nil, err
	}

	sort.Slice(weeks, func(i, j int) bool {
		return weeks[i].Date.Time.Before(weeks[j].Date.Time)
	})

	newWeekIds := make([]string, 0, len(weeks))
	for _, week := range weeks {
		fmt.Println(week)
		newWeekIds = append(newWeekIds, week.ID.Hex())
	}

	league.Weeks = newWeekIds
	fmt.Println("league validateRequest", league)

	return league, nil
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

type WeekIdsRequest struct {
	WeekIds []string `json:"week_ids"`
}

func GetWeeksByIds(c *gin.Context) {

	request := WeekIdsRequest{}
	err := c.Bind(&request)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}

	weeks := make([]*model.Week, 0)
	for _, w := range request.WeekIds {
		week, err := common.GetOneByStringId(common.GlobalDbProvider, &model.Week{}, w)
		if err != nil {
			common.RespondWithError(c, err)
			return
		}
		weeks = append(weeks, week.(*model.Week))
	}

	c.JSON(http.StatusOK, gin.H{common.ResourceKey: weeks})
}
