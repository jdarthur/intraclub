package controllers

import (
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
	"net/http"
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

func WhoAmI(c *gin.Context) {
	token, err := common.GetTokenFromAuthMiddleware(c)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}
	u := &model.User{}
	user, err := common.GetOneByStringId(common.GlobalDbProvider, u, token.UserId)
	if err != nil {
		common.RespondWithApiError(c, common.ApiError{
			References: []string{u.RecordType(), token.UserId},
			Code:       common.CrudRecordWithObjectIdDoesNotExist,
		})
		return
	}

	c.JSON(http.StatusOK, user)
}
