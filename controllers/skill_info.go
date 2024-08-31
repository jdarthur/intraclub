package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
	"strings"
)

type SkillInfoController struct{}

func (s SkillInfoController) Model() common.CrudRecord {
	return &model.SkillInfo{}
}

func (s SkillInfoController) ValidateRequest(c common.CrudRecord, isUpdate bool, provider common.DbProvider) (common.CrudRecord, error) {

	record, ok := c.(*model.SkillInfo)
	if !ok {
		return nil, fmt.Errorf("invalid record: %T %+v", c, c)
	}

	if isUpdate {
		return nil, fmt.Errorf("update is not supported for record type '%s'", record.RecordType())
	}

	record.Captain = strings.TrimSpace(record.Captain)
	record.Level = strings.TrimSpace(record.Level)
	record.Line = strings.TrimSpace(record.Line)

	return c, nil
}

func (s SkillInfoController) GetAllFilter(c *gin.Context) (map[string]interface{}, error) {
	return nil, nil
}

type SkillInfoOptions struct {
	LeagueTypes   []model.LeagueType `json:"league_types"`
	KnownCaptains []string           `json:"known_captains"`
}

func GetSkillInfoOptions(c *gin.Context) {

	output := SkillInfoOptions{
		LeagueTypes: model.LeagueTypes,
	}

	captains, err := model.GetAllCaptains(common.GlobalDbProvider)
	if err != nil {
		common.RespondWithError(c, fmt.Errorf("error getting known captains: %s", err.Error()))
		return
	}

	output.KnownCaptains = captains

	c.JSON(200, gin.H{"resource": output})
}

func GetSkillInfoForUser(c *gin.Context) {
	userId := c.Param("id")
	if userId == "" {
		common.RespondWithApiError(c, common.ApiError{Code: common.EmptyObjectId})
		return
	}

	records, err := common.GetAllWhere(common.GlobalDbProvider, &model.SkillInfo{}, map[string]interface{}{"user_id": userId})
	if err != nil {
		common.RespondWithError(c, err)
		return
	}

	c.JSON(200, gin.H{"resource": records})
}
