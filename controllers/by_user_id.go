package controllers

import (
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
	"net/http"
)

// GetTeamsForUser returns all of the model.Team s that this model.User is a
// a member of. It also marks them as active or inactive.
func GetTeamsForUser(c *gin.Context) {
	id := c.Param("id")

	teams, err := model.GetTeamsForUserId(common.GlobalDbProvider, id)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{common.ResourceKey: teams})
}

// GetLeaguesForUser returns all of the model.League s that this model.User is a
// a member of. It also marks them as active or inactive.
func GetLeaguesForUser(c *gin.Context) {
	id := c.Param("id")

	leagues, err := model.GetLeaguesByUserId(common.GlobalDbProvider, id)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{common.ResourceKey: leagues})
}

func GetCommissionedLeaguesForUser(c *gin.Context) {
	id := c.Param("id")

	leagues, err := model.GetCommissionedLeaguesByUserId(common.GlobalDbProvider, id)
	if err != nil {
		common.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{common.ResourceKey: leagues})

}
