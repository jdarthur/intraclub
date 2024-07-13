package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
	"net/http"
)

type NewUserController struct{}

func (ctl *NewUserController) Register(c *gin.Context) {

	req := &model.User{}
	err := c.Bind(req)
	if err != nil {
		common.RespondWithBadRequest(c, err)
		return
	}

	if req.IsAdmin {
		err = fmt.Errorf("cannot create user and set is_admin=true")
		common.RespondWithBadRequest(c, err)
		return
	}

	created, err := common.Create(common.GlobalDbProvider, req)
	if err != nil {
		common.RespondWithBadRequest(c, err)
		return
	}

	c.JSON(http.StatusOK, created)
}
