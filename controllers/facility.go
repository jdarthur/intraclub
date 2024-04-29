package controllers

import (
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
)

type FacilityController struct{}

func (f FacilityController) Model() common.CrudRecord {
	return new(model.Facility)
}

func (f FacilityController) ValidateRequest(c common.CrudRecord, isUpdate bool, provider common.DbProvider) (common.CrudRecord, error) {
	return c, nil
}

func (f FacilityController) GetAllFilter(c *gin.Context) (map[string]interface{}, error) {

	token, err := common.GetTokenFromAuthMiddleware(c)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"user_id": token.UserId}, nil
}
