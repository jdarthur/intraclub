package controllers

import (
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
	"net/http"
)

type UserParseResult struct {
	Success bool        `json:"success"`
	Error   string      `json:"error"`
	User    *model.User `json:"user"`
}

// ParseUserCsv takes a CSV file as form data and
// parses it into a list of model.User s. It then
// attempts to create all of the users in the list
// with a UserParseResult for each which will include
// any errors we encountered during creation
func ParseUserCsv(c *gin.Context) {
	file := c.PostForm("file")
	users, err := model.ParseUserCsvFromString(file)
	if err != nil {
		common.RespondWithBadRequest(c, err)
		return
	}

	output := make([]UserParseResult, 0)
	for _, user := range users {
		created, err := common.Create(common.GlobalDbProvider, user)
		if err != nil {
			output = append(output, UserParseResult{Success: false, Error: err.Error(), User: user})
		} else {
			output = append(output, UserParseResult{Success: true, User: created.(*model.User)})
		}
	}

	c.JSON(http.StatusCreated, gin.H{common.ResourceKey: output})
}
