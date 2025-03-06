package route

import (
	"github.com/gin-gonic/gin"
	"intraclub/common"
	"intraclub/model"
	"net/http"
)

var UserImportBaseRoute = "/user_import"

type CsvImportResult struct {
	CreatedCount    int
	Created         []*model.User
	AlreadyExisting []*model.User
}

func HandleCsvImport(c *gin.Context) {
	userList, err := model.ParseUserCsvFromReader(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, existing, err := model.ParseAndCreateCsvUsers(common.GlobalDatabaseProvider, userList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, CsvImportResult{
		CreatedCount:    len(created),
		Created:         created,
		AlreadyExisting: existing,
	})
}
